package main

import (
	"encoding/json"
	"fmt"
	"github.com/bitrise-io/go-utils/fileutil"
)

type jsonModel struct {
	Engines struct {
		Npm string
	}
}

func main() {
	content, _ := fileutil.ReadStringFromFile("package.json")
	var m jsonModel;

	_ = json.Unmarshal([]byte(content), &m)
	fmt.Printf("detected npm version: %s\n", m.Engines.Npm)
}