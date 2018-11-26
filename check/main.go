package main

import (
	"encoding/json"
	"github.com/devinotelecom/concourse-git-resource/controller"
	log "github.com/sirupsen/logrus"
	"os"
)

func main() {
	var input controller.Payload

	err := json.NewDecoder(os.Stdin).Decode(&input)
	if err != nil {
		log.Fatalln(err)
	}

	controller.Init(input,"")

	err = json.NewEncoder(os.Stdout).Encode(controller.Check(input,""))
	if err != nil {
		log.Fatalln(err)
	}
}
