package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/devinotelecom/concourse-git-resource/controller"
	"encoding/json"
	"os"
)

func main() {
	var input controller.Payload

	err := json.NewDecoder(os.Stdin).Decode(&input)
	if err != nil {
		log.Fatalln(err)
	}

	controller.Init(input.Source.Url, input.Source.Branch, input.Source.PrivateKey, "")

	if input.Source.TagFilter != "" {
		err = json.NewEncoder(os.Stdout).Encode(controller.LastTag(""))
		if err != nil {
			log.Fatalln(err)
		}
	} else {
		err = json.NewEncoder(os.Stdout).Encode(controller.LastCommit("", input.Source.Branch))
		if err != nil {
			log.Fatalln(err)
		}
	}
}
