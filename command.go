package main

import (
	"encoding/base64"
	_ "encoding/base64"
	"fmt"
	"log"
	"strconv"
	"strings"
)

const ShellClass = 1

type Command struct {
	class   int
	content string
}

func NewCommand(class int, content string) *Command {
	return &Command{
		class:   class,
		content: content,
	}
}

func (c *Command) Execute() {

}

func (c *Command) Encode() string {
	encodedContent := fmt.Sprintf("%d:%s", c.class,
		base64.StdEncoding.EncodeToString([]byte(c.content)))
	encodedCommand := base64.StdEncoding.EncodeToString([]byte(encodedContent))
	return encodedCommand
}

func DecodeCommand(encodedCommand string) *Command {
	decodedCommand, err := base64.StdEncoding.DecodeString(encodedCommand)
	if err != nil {
		log.Fatal("An error occurred in DecodeCommand() method while decoding command: ", err)
	}

	splitCommand := strings.Split(string(decodedCommand), ":")
	if len(splitCommand) != 2 {
		log.Fatal("An error occurred in DecodeCommand() method while splitting command: len of the splitCommand " +
			"is not equal to 2")
	}

	decodedClass, err := strconv.Atoi(splitCommand[0])
	if err != nil {
		log.Fatal("An error occurred in DecodeCommand() method while decoding class: ", err)
	}

	decodedContent, err := base64.StdEncoding.DecodeString(splitCommand[1])
	if err != nil {
		log.Fatal("An error occurred in DecodeCommand() method while decoding content: ", err)
	}

	return &Command{
		class:   decodedClass,
		content: string(decodedContent),
	}
}
