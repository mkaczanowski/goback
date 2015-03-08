package util

import (
	"bufio"
	"log"
)

var Debug bool

func sendCmd(w *bufio.Writer, cmd string, flush bool) error {
	_, err := w.Write([]byte(cmd+"\n"))

	if Debug == true {
		log.Println("Send comand: ", cmd)
	}

	if flush {
		w.Flush()
	}

	return err
}

func MustSendCmd(w *bufio.Writer, cmd string, flush bool) {
	err := sendCmd(w, cmd, flush)

	if err != nil {
		panic(err)
	}
}
