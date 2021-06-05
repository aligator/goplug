package goplug

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
)

type RegisterMessage struct {
	ID string `json:"id"`
}

type Plugin struct {
	ID       string
	commands map[string]func(data []byte) error
}

func (p *Plugin) Register(ID string) error {
	return p.Send("register", RegisterMessage{
		ID: ID,
	})
}

func (p *Plugin) Send(cmd string, payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	fmt.Println(cmd, string(data))
	return nil
}

func (p *Plugin) Log(message string) error {
	return p.Send("log", message)
}

// RegisterCommand can be used to register commands, this plugin listens to.
// The cmd should be a unique string.
//
// The factory is used to create a new instance of whatever the message should be parsed to (using json.Unmarshal).
// It has to return a pointer.
//
// listener is the actual function to call when the command occurs. The message is already parsed from json and you can
// safely assume that it is of the type, the factory returns. So you can safely convert and use it like this:
//  data := message.(*DoPrintMessage)
//	return listener(data.Text)
func (p *Plugin) RegisterCommand(cmd string, factory func() interface{}, listener func(message interface{}) error) {
	if p.commands == nil {
		p.commands = make(map[string]func(data []byte) error)
	}

	p.commands[cmd] = func(message []byte) error {
		data := factory()
		err := json.Unmarshal(message, &data)
		if err != nil {
			return err
		}

		return listener(data)
	}
}

func (p *Plugin) Run() error {
	reader := bufio.NewReader(os.Stdin)
	for {
		message, _, err := reader.ReadLine()
		if err != nil {
			return err
		}

		cmd, data, err := parseMessage(string(message))
		if err != nil {
			return err
		}

		if cmd == "close" {
			break
		}

		if p.commands == nil {
			continue
		}

		if listener, ok := p.commands[cmd]; ok {
			err := listener(data)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
