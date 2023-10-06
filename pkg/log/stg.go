package log

import (
	"fmt"
)

type Logger interface {
	Err(err error, msg string)
	Debug(msg string)
}

type Simple struct {
}

func (l *Simple) Err(err error, msg string) {
	println(fmt.Sprintf("error: %v msg: %s", err, msg))
}
func (l *Simple) Debug(msg string) {
	println(fmt.Sprintf("debug: msg: %s", msg))
}
