package main

import (
	"fmt"

	qagent "github.com/xanzy/ssh-agent"
	"golang.org/x/crypto/ssh"
)

func main() {
	user := "rphilips"
	host := "dev.anet.be:22"
	cop, _, err := qagent.New()

	if err != nil {
		fmt.Println("No agent:", err)
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
		fmt.Println("failed to dial:", host, e)
		return
	}

	defer conn.Close()

	session, e := conn.NewSession()

	if e != nil {
		fmt.Println("failed to create session on", host, e)
		return
	}
	defer session.Close()
	out, err := session.CombinedOutput("uptime")
	if err != nil {
		fmt.Println("failed to run uptime", host, e)
	}
	fmt.Println(string(out))
}
