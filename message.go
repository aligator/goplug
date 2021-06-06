package goplug

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
)

// parseMessage splits the message by the first " ".
// The first part is the command the rest is the payload
// which should be valid json.
func parseMessage(message string) (string, []byte, error) {
	split := strings.SplitN(message, " ", 2)
	if len(split) != 2 {
		return "", nil, errors.New("invalid message")
	}

	return split[0], []byte(split[1]), nil
}

func onLog(data []byte) error {
	var logMessage string
	err := json.Unmarshal(data, &logMessage)
	if err != nil {
		return err
	}
	log.Print(logMessage)
	return nil
}

func onRegister(g *GoPlug, p *plugin, data []byte) error {

	registerMessage := &RegisterMessage{}
	err := json.Unmarshal(data, registerMessage)
	if err != nil {
		return err
	}

	if _, ok := g.registeredPlugins[registerMessage.ID]; ok {
		err := fmt.Errorf("a plugin with the id %v is already registered", registerMessage.ID)
		// Kill the process as it is of no use anymore.
		_ = p.Process.Kill()
		return err
	}

	// Register the plugin.
	p.ID = registerMessage.ID
	g.registeredPlugins[registerMessage.ID] = p
	return nil
}

// onMessage gets called when the plugin sent a new message.
func (g *GoPlug) onMessage(p *plugin) func(message []byte) {
	return func(message []byte) {
		cmd, data, err := parseMessage(string(message))
		if err != nil {
			fmt.Println(err)
			return
		}

		// First handle GoPlug specific messages.

		// If id is not set yet, the first message must be
		// a "register" command containing the id.
		if p.ID == "" {
			if cmd != "register" {
				fmt.Println(errors.New("the first message has to be a 'register' command"))
				return
			}

			err := onRegister(g, p, data)
			if err != nil {
				fmt.Println(err)
			}
			return
		}

		// log can be used to print log messages inside of a plugin.
		if cmd == "log" {
			err := onLog(data)
			if err != nil {
				fmt.Println(err)
			}
			return
		}

		// All other messages are forwarded to all listeners
		for _, listener := range g.onCommandListener {
			err := listener(p, cmd, data)
			if err != nil {
				fmt.Println(err)
				continue
			}
		}
	}
}
