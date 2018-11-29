package main

import (
	"encoding/json"
	"fmt"
	"github.com/devinotelecom/concourse-git-resource/controller"
	log "github.com/sirupsen/logrus"
	"os"
	"strings"
)

func main() {
	var input controller.Payload

	path := strings.Trim(fmt.Sprint(os.Args[1:]), "[]")
	if path == "" {
		path = "/tmp/git-resource-request/"
	}

	err := json.NewDecoder(os.Stdin).Decode(&input)
	if err != nil {
		log.Fatalln(err)
	}

	config := controller.Config{
		Input: &input,
		Path: path,
	}

	controller.Init(config)

	controller.Checkout(config)

	metadata := controller.GetMetaData(path, input)

	result := controller.MetadataJson{controller.Ref{input.Version.Ref}, metadata}

	err = json.NewEncoder(os.Stdout).Encode(result)
	if err != nil {
		log.Fatalln(err)
	}
}
