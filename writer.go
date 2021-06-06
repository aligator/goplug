package goplug

import (
	"bytes"
)

// writer is just calls onMessage for each line (separated by \n).
type writer struct {
	onMessage func(message []byte)
}

func (w writer) Write(p []byte) (n int, err error) {
	messages := bytes.Split(p, []byte("\n"))

	for _, m := range messages {
		if len(m) == 0 {
			continue
		}
		w.onMessage(m)
	}

	return len(p), nil
}
