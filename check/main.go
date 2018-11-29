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

	config := controller.Config{
		Input: &input,
		Path: "/tmp/git-resource-request/",
	}

	controller.Init(config)

	err = json.NewEncoder(os.Stdout).Encode(controller.Check(config))
	if err != nil {
		log.Fatalln(err)
	}
}
