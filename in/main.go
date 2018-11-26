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

	err := json.NewDecoder(os.Stdin).Decode(&input)
	if err != nil {
		log.Fatalln(err)
	}

	controller.Init(input, path)

	controller.Checkout(path, input.Version.Ref)

	metadata := controller.GetMetaData(input.Version.Ref, path)

	result := controller.MetadataJson{controller.Ref{input.Version.Ref}, metadata}

	err = json.NewEncoder(os.Stdout).Encode(result)
	if err != nil {
		log.Fatalln(err)
	}
}
