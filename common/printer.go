package common

import (
	"encoding/json"
	"fmt"
)

type printer interface {
	PrintData(interface{})
	PrintLog(string)
}

type Printer struct{}

func (p *Printer) PrintData(data interface{}) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		panic(fmt.Sprintln("JSON print error:", err))
	}
	fmt.Println(string(jsonData))
}
