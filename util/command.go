package util

import (
	"bufio"
)

func sendCmd(w *bufio.Writer, cmd string, flush bool) error {
	//log.Println("Send comand: ", cmd)
	_, err := w.Write([]byte(cmd))
	w.WriteRune(rune(13))

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
