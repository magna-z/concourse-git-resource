package common

import (
	"encoding/json"
	"fmt"
	"os"
)

type printer interface {
	PrintData(interface{})
	PrintLog(string)
}

type Printer struct{}

func (Printer) PrintData(data interface{}) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		panic(fmt.Sprintln("JSON print error:", err))
	}
	fmt.Println(string(jsonData))
}

func (Printer) PrintLog(message string) {
	fmt.Fprintln(os.Stderr, message)
}
