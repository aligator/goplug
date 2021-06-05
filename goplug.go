package goplug

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os/exec"
	"path"
	"strings"
)

type event string

type plugin struct {
	*exec.Cmd
	id string
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

	// ToDo: implement checks

	return true
}

func (g *GoPlug) RegisterOnCommand(listener func(cmd string, data []byte) error) {
	g.onCommandListener = append(g.onCommandListener, listener)
}

func (g *GoPlug) onMessage(p *plugin) func(message []byte) {
	return func(message []byte) {
		// Commands always consist of one command name and json separated by an ' '.
		split := strings.SplitN(string(message), " ", 2)
		if len(split) != 2 {
			fmt.Println(errors.New("invalid message"))
			return
		}

		// First handle GoPlug specific messages.

		// If id is not set yet, the first message must be
		// a "register" command containing the id.
		if p.id == "" {
			if split[0] != "register" {
				fmt.Println(errors.New("the first message has to be a 'register' command"))
				return
			}

			registerMessage := &RegisterMessage{}
			err := json.Unmarshal([]byte(split[1]), registerMessage)
			if err != nil {
				fmt.Println(err)
				return
			}

			p.id = registerMessage.ID
			return
		}

		// All other messages are forwarded to all listeners
		for _, listener := range g.onCommandListener {
			err := listener(split[0], []byte(split[1]))
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

		p.Stdout = writer{
			onMessage: g.onMessage(&p),
		}

		g.processes = append(g.processes, p)

		err = p.Start()
		if err != nil {
			return err
		}
	}

	return nil
}

func (g *GoPlug) Wait() {
	for _, p := range g.processes {
		p.Wait()
	}
}
