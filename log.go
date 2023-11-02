package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
)

type Log struct {
	id          int
	term        int
	command     Command
	commitIndex int
}

func NewLog(id int, term int, command *Command, commitIndex int) *Log {
	return &Log{
		id:          id,
		term:        term,
		command:     *command,
		commitIndex: commitIndex,
	}
}

func (l *Log) Encode() string {
	return fmt.Sprintf("%d:%d:%s:%d",
		l.id, l.term, l.command.Encode(), l.commitIndex)
}

func DecodeLog(encodedLog string) *Log {
	splitLog := strings.Split(encodedLog, ":")
	if len(splitLog) != 4 {
		log.Fatal("An error occurred in DecodeLog() method while splitting log: " +
			"len of the splitCommand should be equal to 4")
	}

	id, err := strconv.Atoi(splitLog[0])
	if err != nil {
		log.Fatal("An error occurred in DecodeLog() method while turning id to int: ", err)
	}

	term, err := strconv.Atoi(splitLog[1])
	if err != nil {
		log.Fatal("An error occurred in DecodeLog() method while turning term to int: ", err)
	}

	commitIndex, err := strconv.Atoi(splitLog[3])
	if err != nil {
		log.Fatal("An error occurred in DecodeLog() method while turning commitIndex to int: ", err)
	}

	return &Log{
		id:          id,
		term:        term,
		command:     *DecodeCommand(splitLog[2]),
		commitIndex: commitIndex,
	}
}

func (l *Log) ToString() string {
	return fmt.Sprintf("---------- LOG %d\nTerm %d\n%s\nCommitIndex %d", l.id, l.term, l.command.ToString(),
		l.commitIndex)
}
