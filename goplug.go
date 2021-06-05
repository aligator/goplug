package goplug

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"log"
	"os/exec"
	"path"
	"strings"
)

type plugin struct {
	*exec.Cmd
	id        string
	stdinPipe io.WriteCloser
}

type GoPlug struct {
	PluginFolder string

	processes         []plugin
	onCommandListener []func(cmd string, data []byte) error
}

func isValidPlugin(info fs.FileInfo) bool {
	if info.IsDir() {
		return false
	}

	if info.Name() == ".gitkeep" {
		return false
	}

	// ToDo: implement checks

	return true
}

func parseMessage(message string) (string, []byte, error) {
	// Commands always consist of one command name and json separated by an ' '.
	split := strings.SplitN(message, " ", 2)
	if len(split) != 2 {
		return "", nil, errors.New("invalid message")
	}

	return split[0], []byte(split[1]), nil
}

func (g *GoPlug) RegisterOnCommand(registerCmd string, factory func() interface{}, listener func(message interface{}) error) {
	g.onCommandListener = append(g.onCommandListener, func(cmd string, message []byte) error {
		if cmd != registerCmd {
			return nil
		}

		data := factory()
		err := json.Unmarshal(message, &data)
		if err != nil {
			return err
		}

		return listener(data)
	})
}

func (g *GoPlug) onMessage(p *plugin) func(message []byte) {
	return func(message []byte) {
		cmd, data, err := parseMessage(string(message))
		if err != nil {
			fmt.Println(err)
			return
		}

		// First handle GoPlug specific messages.

		if cmd == "log" {
			var logMessage string
			err := json.Unmarshal(data, &logMessage)
			if err != nil {
				fmt.Println(err)
				return
			}

			log.Println(logMessage)
			return
		}

		// If id is not set yet, the first message must be
		// a "register" command containing the id.
		if p.id == "" {
			if cmd != "register" {
				fmt.Println(errors.New("the first message has to be a 'register' command"))
				return
			}

			registerMessage := &RegisterMessage{}
			err := json.Unmarshal([]byte(data), registerMessage)
			if err != nil {
				fmt.Println(err)
				return
			}

			p.id = registerMessage.ID
			return
		}

		// All other messages are forwarded to all listeners
		for _, listener := range g.onCommandListener {
			err := listener(cmd, data)
			if err != nil {
				fmt.Println(err)
				continue
			}
		}
	}
}

func (g *GoPlug) Run() error {
	entries, err := ioutil.ReadDir(g.PluginFolder)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if !isValidPlugin(entry) {
			continue
		}

		// Start the plugin.
		filePath := path.Join(g.PluginFolder, entry.Name())
		p := plugin{
			Cmd: &exec.Cmd{
				Path: filePath,
				Args: []string{filePath},
			},
		}

		p.stdinPipe, err = p.StdinPipe()
		if err != nil {
			return err
		}

		p.Stdout = writer{
			onMessage: g.onMessage(&p),
		}

		g.processes = append(g.processes, p)

		err = p.Start()
		if err != nil {
			fmt.Println(err)
			continue
		}
	}

	return nil
}

func (g *GoPlug) SendAll(cmd string, message interface{}) error {
	var data []byte

	if message != nil {
		var err error
		data, err = json.Marshal(message)
		if err != nil {
			return err
		}
	}

	for _, p := range g.processes {
		_, err := p.stdinPipe.Write([]byte(cmd + " " + string(data) + "\n"))
		if err != nil {
			fmt.Println(err)
			continue
		}
	}

	return nil
}

func (g *GoPlug) Wait() error {
	err := g.SendAll("close", nil)
	if err != nil {
		return err
	}

	for _, p := range g.processes {
		err := p.Wait()
		if err != nil {
			fmt.Println(err)
			continue
		}

		p.stdinPipe.Close()
	}
	return nil
}
