package ssh

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"runtime"
	"strings"
	"sync"

	"golang.org/x/crypto/ssh"

	qagent "github.com/xanzy/ssh-agent"

	qregistry "brocade.be/base/registry"
)

// Payload something to send over SSH
type Payload interface {
	GetID() string
	GetUID() string
	GetCMD() string
	GetOrigin() string
	SetOrigin(origin string)
	Send(encoder *gob.Encoder) error
}

// SSHcmd send commands over SSH
//  payload stands for the action which have to be executed over the SSH link
//  catchOut contains - on the initiating machine - the writing on stdout of the command on the target
//  catchErr contains - on the initiating machine - the writing on stderr of the command on the target
func SSHcmd(payload Payload, whowhere string) (catchOut *bytes.Buffer, catchErr *bytes.Buffer, err error) {
	catchOut = &bytes.Buffer{}
	catchErr = &bytes.Buffer{}
	payload.SetOrigin("")
	user, host := parseRemote(whowhere, payload.GetUID())
	if user == "" || host == "" {
		err = fmt.Errorf("no host and/or user specified")
		return
	}
	cop, _, err := qagent.New()
	if err != nil {
		if runtime.GOOS == "windows" {
			err = fmt.Errorf("cannot find SSH agent. On windows, work with PuTTY and Pageant: `%s`", err)
			return
		}
		err = fmt.Errorf("cannot find SSH agent `%s`", err)
		return
	}

	auth := ssh.PublicKeysCallback(cop.Signers)

	sshConfig := &ssh.ClientConfig{
		User:            user,
		Auth:            []ssh.AuthMethod{auth},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	conn, e := ssh.Dial("tcp", host, sshConfig)

	if e != nil {
		err = fmt.Errorf("failed to dial `%s@%s`:\n%s", user, host, e)
		return
	}
	defer conn.Close()

	session, e := conn.NewSession()

	if e != nil {
		err = fmt.Errorf("failed to create session on `%s@%s`:\n%s", user, host, e)
		return
	}
	defer session.Close()

	// stream redirection

	ssherr, err := session.StderrPipe()
	if err != nil {
		return
	}

	sshout, err := session.StdoutPipe()
	if err != nil {
		return
	}

	sshin, err := session.StdinPipe()
	if err != nil {
		return
	}

	wg := &sync.WaitGroup{}
	wg.Add(3)

	go func() {
		defer wg.Done()
		io.Copy(catchErr, ssherr)
	}()

	go func() {
		defer wg.Done()
		io.Copy(catchOut, sshout)
	}()

	go func() {
		defer wg.Done()
		defer sshin.Close()
		enc := gob.NewEncoder(sshin)
		err = payload.Send(enc)
		if err != nil {
			return
		}
	}()

	err = session.Run(payload.GetCMD())
	wg.Wait()
	return
}

func parseRemote(remote string, uid string) (user string, host string) {
	if strings.Contains(remote, "@") {
		user = strings.SplitN(remote, "@", 2)[0]
		host = remote[len(user)+1:]
	}
	if user == "" {
		user = uid
	}
	if user == "" {
		user = qregistry.Registry["ssh-default-user"]
	}
	if host == "" {
		host = qregistry.Registry["ssh-default-host"]
	}
	return
}
