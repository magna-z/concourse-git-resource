package common

import (
	"encoding/json"
	"fmt"
)

type Source struct {
	Url        string
	PrivateKey string `json:"private_key"`
	Login      string
	Password   string
	Branch     string
	TagRegex   string `json:"tag_regex"`
	Paths      []string
}

type Version struct {
	Reference string `json:"ref"`
}

type Payload struct {
	Source  Source
	Version Version
}

func Parse(payload interface{}, stdin []byte) {
	err := json.Unmarshal(stdin, &payload)
	if err != nil {
		panic(fmt.Sprintln("Unmarshal payload error:", err))
	}
}
