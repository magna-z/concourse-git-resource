package main

import (
	"io/ioutil"
	"os"

	"concourse-git-resource/common"
	"concourse-git-resource/resource"
)

const (
	CheckResourceBin = "/opt/resource/check"
	InResourceBin    = "/opt/resource/in"
	OutResourceBin   = "/opt/resource/out"
)

func main() {
	stdin, _ := ioutil.ReadAll(os.Stdin)
	printer := &common.Printer{}

	switch os.Args[0] {
	case CheckResourceBin:
		p := resource.NewCheckPayload(stdin)
		resource.Check(p, printer)
	case InResourceBin:
		p := resource.NewInPayload(stdin)
		resource.In(p, os.Args[1], printer)
	case OutResourceBin:
		p := resource.NewOutPayload(stdin)
		resource.Out(p, os.Args[1], printer)
	}
}
