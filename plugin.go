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

func (p *Plugin) RegisterCommand(cmd string, listener func(message []byte) error) {
	if p.commands == nil {
		p.commands = make(map[string]func(data []byte) error)
	}
	p.commands[cmd] = listener
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
			p.Log("do close")
			break
		}

		if p.commands == nil {
			continue
		}

		if listener, ok := p.commands[cmd]; ok {
			p.Log(fmt.Sprint(data))
			err := listener(data)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
