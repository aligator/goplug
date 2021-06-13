package goplug

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aligator/goplug/message"
	"log"
)

func onLog(data []byte) error {
	var logMessage string
	err := json.Unmarshal(data, &logMessage)
	if err != nil {
		return err
	}
	log.Print(logMessage)
	return nil
}

func onInitialized(p *internalPlugin) error {
	close(p.initializedSig)
	return nil
}

func onRegister(g *GoPlug, p *internalPlugin, data []byte) error {

	registerMessage := &message.RegisterMessage{}
	err := json.Unmarshal(data, registerMessage)
	if err != nil {
		return err
	}

	if _, ok := g.registeredPlugins[registerMessage.ID]; ok {
		err := fmt.Errorf("a internalPlugin with the id %v is already registered", registerMessage.ID)
		// Kill the process as it is of no use anymore.
		_ = p.Process.Kill()
		return err
	}

	// Register the internalPlugin.
	p.ID = registerMessage.ID
	g.registeredPlugins[registerMessage.ID] = p
	return nil
}

// onMessage gets called when the internalPlugin sent a new message.
func (g *GoPlug) onMessage(p *internalPlugin) func(payload []byte) {
	return func(payload []byte) {
		cmd, data, err := message.Parse(string(payload))
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

		if cmd == "lastMessage" {
			close(p.lastMessageSig)
		}

		// log can be used to print log messages inside of a internalPlugin.
		if cmd == "log" {
			err := onLog(data)
			if err != nil {
				fmt.Println(err)
			}
			return
		}

		if cmd == "initialized" {
			err := onInitialized(p)
			if err != nil {
				fmt.Println(err)
			}
			return
		}

		// All other messages are forwarded to all listeners
		for _, listener := range g.onCommandListener {
			resData, err := listener(p, cmd, data)
			if err != nil {
				fmt.Println(err)
				continue
			}
			if resData == nil {
				return
			}

			// ToDo: send only to the plugin which requested this (p)
			err = g.Send(cmd, resData)
			if err != nil {
				fmt.Println(err)
				continue
			}
		}
	}
}
