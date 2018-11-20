package main

import (
	"encoding/json"
	"os"
	log "github.com/sirupsen/logrus"
	"github.com/devinotelecom/concourse-git-resource/controller"
	"fmt"
	"strings"
)

func main()  {
	var input controller.Payload

	arg := fmt.Sprint(os.Args[1:])
	path := strings.Trim(arg, "[]")

	err := json.NewDecoder(os.Stdin).Decode(&input)
	if err != nil {
		log.Fatalln(err)
	}

	controller.Init(input.Source.Url, input.Source.Branch, input.Source.PrivateKey, path)

	controller.CheckoutCommit(input.Version.Ref, path)

	metadata := controller.GetMetaData(input.Version.Ref, input.Source.Branch, path)

	result := controller.MetadataArry{controller.Ref{input.Version.Ref}, metadata}

	err = json.NewEncoder(os.Stdout).Encode(result)
	if err != nil {
		log.Fatalln(err)
	}
}