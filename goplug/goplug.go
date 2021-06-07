package goplug

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"time"
)

type PluginInfo struct {
	// ID is set on the first message sent from the plugin during registration.
	// it identifies the plugin uniquely.
	ID string
}

// plugin is the internal representation of a plugin binary.
type plugin struct {
	PluginInfo
	*exec.Cmd

	// initializedSig gets closed when initialized
	initializedSig chan bool

	// stdinPipe is the pipe to send data to the plugin.
	stdinPipe io.WriteCloser

	finishedSig    chan bool
	lastMessageSig chan bool
}

// GoPlug loads plugins from the PluginFolder.
// Use Init to execute them.
type GoPlug struct {
	PluginFolder string

	// plugins contains all known plugins found in the plugin folder.
	// They may not be all valid plugin binaries.
	// When they got started and sent the 'register' command they get added to
	// the registeredPlugins map.
	plugins []plugin

	// registeredPlugins contains references to all plugins which already
	// registered themselves.
	registeredPlugins map[string]*plugin
	onCommandListener []func(p *plugin, cmd string, data []byte) error
}

func isValidPlugin(info fs.FileInfo) bool {
	if info.IsDir() {
		return false
	}

	if info.Name() == ".gitkeep" {
		return false
	}

	// ToDo: implement checks
	//       Maybe invent a custom filename rule, such as
	//       "***.plugin" ("***.plugin.exe" on windows).
	//       This could be made configurable...
	return true
}

// RegisterOnCommand can be used to register commands, plugins can send.
// The cmd should be a unique string.
//
// The factory is used to create a new instance of whatever the message should
// be parsed to (using json.Unmarshal).
// It has to return a pointer.
//
// listener is the actual function to call when the command occurs.
// The message is already parsed from json and you can
// safely assume that it is of the type, the factory returns.
// So you can safely convert and use it like this:
//  data := message.(*YourMessageType)
//	return listener(data.Text)
func (g *GoPlug) RegisterOnCommand(registerCmd string, factory func() interface{}, listener func(p PluginInfo, message interface{}) error) {
	g.onCommandListener = append(g.onCommandListener, func(p *plugin, cmd string, message []byte) error {
		if cmd != registerCmd {
			return nil
		}

		if factory == nil {
			return listener(p.PluginInfo, nil)
		}

		data := factory()
		err := json.Unmarshal(message, &data)
		if err != nil {
			return err
		}

		return listener(p.PluginInfo, data)
	})
}

// Init initializes and starts all plugins.
// It blocks until all plugins are initialized.
func (g *GoPlug) Init() error {
	entries, err := ioutil.ReadDir(g.PluginFolder)
	if err != nil {
		return err
	}

	g.registeredPlugins = make(map[string]*plugin)

	// ToDo: collect errors and return them all.
	for _, entry := range entries {
		if !isValidPlugin(entry) {
			continue
		}

		// Start the plugin.
		filePath := path.Join(g.PluginFolder, entry.Name())
		p := plugin{
			Cmd: &exec.Cmd{
				Stderr: os.Stderr,
				Path:   filePath,
				// ToDo: maybe add something as arg to indicate that the binary
				//       should be started in "plugin-mode".
				Args: []string{filePath},
			},
			initializedSig: make(chan bool),
			finishedSig:    make(chan bool),
			lastMessageSig: make(chan bool),
		}

		p.stdinPipe, err = p.StdinPipe()
		if err != nil {
			return err
		}

		p.Stdout = writer{
			onMessage: g.onMessage(&p),
		}

		g.plugins = append(g.plugins, p)

		// Run in extra go routine to be able
		// to do something directly after it closes
		// (e.g. set the finishedSig)
		go func() {
			err = p.Run()
			if err != nil {
				fmt.Println(err)
			}

			// Wait for the last message to be received.
			<-p.lastMessageSig

			close(p.finishedSig)
		}()
	}

	deadline := time.Now().Add(3 * time.Second)
	done := make(chan bool)
	go func() {
		for _, p := range g.plugins {
			var initialized bool

			for {
				select {
				case _, ok := <-p.initializedSig:
					if !ok {
						initialized = true
					}
				default:
				}

				if initialized || deadline.Before(time.Now()) {
					break
				}
			}

			if !initialized {
				// Process was not initializedSig in time.
				p.Process.Kill()
				close(p.finishedSig)
				if _, ok := g.registeredPlugins[p.ID]; ok {
					delete(g.registeredPlugins, p.ID)
				}
			}
		}
		close(done)
	}()

	<-done

	return g.Send("allInitialized", nil)
}

// Send a command to all plugins.
// The message can be of any type which is marshal-able to json.
func (g *GoPlug) Send(cmd string, message interface{}) error {
	// ToDo: add a way to only send to plugins which actually listen to it.
	//       To do this Plugin.OnDoPrint has to send a message to GoPlug to
	//       register it there.
	var data []byte

	if message != nil {
		var err error
		data, err = json.Marshal(message)
		if err != nil {
			return err
		}
	}

	// ToDo: collect errors and return them all.
	for _, p := range g.plugins {
		// Only send to still running plugins.
		select {
		case <-p.finishedSig:
		default:
			_, err := p.stdinPipe.Write([]byte(cmd + " " + string(data) + "\n"))
			if err != nil {
				fmt.Println(err)
				continue
			}
		}
	}

	return nil
}

// Close sends a 'close' message to all plugins and waits until all plugins are
// stopped.
func (g *GoPlug) Close() error {
	err := g.Send("close", nil)
	if err != nil {
		return err
	}

	// ToDo: only wait for a certain time and kill all plugins not stopped after
	//       that.

	for _, p := range g.plugins {
		_, _ = <-p.finishedSig
		//err = p.stdinPipe.Close()
		//if err != nil {
		//	fmt.Println(err)
		//	continue
		//}
	}
	return nil
}
