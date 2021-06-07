package message

import (
	"errors"
	"strings"
)

// Parse splits the message by the first " ".
// The first part is the command the rest is the payload
// which should be valid json.
func Parse(message string) (string, []byte, error) {
	split := strings.SplitN(message, " ", 2)
	if len(split) != 2 {
		return "", nil, errors.New("invalid message")
	}

	return split[0], []byte(split[1]), nil
}

// RegisterMessage is the message which is the
// payload for the 'register' command.
type RegisterMessage struct {
	ID string `json:"id"`
}
