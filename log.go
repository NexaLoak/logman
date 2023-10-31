package main

import "fmt"

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
