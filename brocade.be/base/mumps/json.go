package mumps

import (
	"encoding/json"
	"io"
	"os/exec"

	qregistry "brocade.be/base/registry"
)

func Reader(action string, payload map[string]string) (io.Reader, io.Reader, error) {

	cmd := exec.Command(qregistry.Registry["m-exe"], "%Iter^stdjapi")
	cmd.Dir = qregistry.Registry["m-db"]

	// cmd := exec.Command("yottadb", "-run", "%Iter^stdjapi")
	// cmd.Dir = qregistry.Registry["m-db"]

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, nil, err
	}
	outs, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, err
	}
	errs, err := cmd.StderrPipe()
	if err != nil {
		return nil, nil, err
	}
	cmd.Start()

	go func() {
		defer stdin.Close()
		data := map[string]interface{}{
			"action":    action,
			"payload":   payload,
			"preamble":  "",
			"postamble": "",
			"error":     "<.?-error-!.>",
		}
		jdata, _ := json.Marshal(data)
		io.WriteString(stdin, string(jdata)+"\n")
	}()

	return outs, errs, nil
}
