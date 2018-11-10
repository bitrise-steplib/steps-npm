package main

import (
	"fmt"
	"github.com/bitrise-io/go-utils/command"
)

func main() {
	fmt.Println("hello world")
	command.NewWithStandardOuts("echo", "hello world -- from go-utils command").Run()
}