package main

import (
	"log"
	"net/http"
	"os/exec"

	registry "brocade.be/base/registry"
	handle "brocade.be/qui/lib/handle"
)

const port = ":8081"
const baseURL = "http://localhost" + port + "/"

func main() {
	go func() {
		quiExe := registry.Registry["qtechng-qui-exe"]
		cmd := exec.Command(quiExe, baseURL)
		err := cmd.Run()
		if err != nil {
			log.Fatalf("browser error %v", err)
		}
	}()

	http.HandleFunc("/", handle.Start)
	http.HandleFunc("/result", handle.Result)

	err := http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatalf("server error %v", err)
	}
}
