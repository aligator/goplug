package goplug

import (
	"encoding/json"
	"fmt"
)

type RegisterMessage struct {
	ID string `json:"id"`
}

type PluginImpl struct {
	ID string
}

func (p *PluginImpl) Register(ID string) error {
	return p.Send("register", RegisterMessage{
		ID: ID,
	})
}

func (p *PluginImpl) Send(cmd string, payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	fmt.Println(cmd, string(data))
	return nil
}

type Plugin interface {
	Register(ID string) error
	Send(cmd string, payload interface{}) error
}
