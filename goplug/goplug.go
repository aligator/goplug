package goplug

import (
	"encoding/json"
	"fmt"
	"github.com/aligator/goplug/plugin"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"time"
)

type PluginInfo struct {
	// ID is set on the first message sent from the internalPlugin during registration.
	// it identifies the internalPlugin uniquely.
	ID string
}

// internalPlugin is the internal representation of a internalPlugin binary.
type internalPlugin struct {
	PluginInfo
	*exec.Cmd

	// initializedSig gets closed when initialized
	initializedSig chan bool

	// stdinPipe is the pipe to send data to the internalPlugin.
	stdinPipe io.WriteCloser

	finishedSig    chan bool
	lastMessageSig chan bool
}

// GoPlug loads plugins from the PluginFolder.
// Use Init to execute them.
type GoPlug struct {
	PluginFolder string

	// plugins contains all known plugins found in the internalPlugin folder.
	// They may not be all valid internalPlugin binaries.
	// When they got started and sent the 'register' command they get added to
	// the registeredPlugins map.
	plugins []internalPlugin

	// registeredPlugins contains references to all plugins which already
	// registered themselves.
	registeredPlugins map[string]*internalPlugin
	onCommandListener []func(p *internalPlugin, cmd string, message []byte) (interface{}, error)
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
	//       "***.internalPlugin" ("***.internalPlugin.exe" on windows).
	//       This could be made configurable...
	return true
}

// RegisterCommand can be used to register commands, plugins can send.
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
func (g *GoPlug) RegisterCommand(name string, factory func() interface{}, listener func(p PluginInfo, message interface{}) error) {
	g.onCommandListener = append(g.onCommandListener, func(p *internalPlugin, cmd string, message []byte) (interface{}, error) {
		if cmd != name {
			return nil, nil
		}

		if factory == nil {
			return nil, listener(p.PluginInfo, nil)
		}

		data := factory()
		err := json.Unmarshal(message, &data)
		if err != nil {
			return nil, err
		}

		return nil, listener(p.PluginInfo, data)
	})
}

func (g *GoPlug) RegisterFunc(name string, factory func() interface{}, listener func(p PluginInfo, message interface{}) (interface{}, error)) {
	g.onCommandListener = append(g.onCommandListener, func(p *internalPlugin, cmd string, message []byte) (interface{}, error) {
		if cmd != name {
			return nil, nil
		}

		// First get the FuncCommand as it includes the funcNum which is needed to send the result
		funcData := plugin.FuncCommand{}
		err := json.Unmarshal(message, &funcData)
		if err != nil {
			return nil, err
		}

		if factory == nil {
			// TODO:
			return listener(p.PluginInfo, funcData)
		}

		data := factory()
		err = json.Unmarshal([]byte(funcData.Payload), &data)
		if err != nil {
			return nil, err
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

	g.registeredPlugins = make(map[string]*internalPlugin)

	// ToDo: collect errors and return them all.
	for _, entry := range entries {
		if !isValidPlugin(entry) {
			continue
		}

		// Start the internalPlugin.
		filePath := path.Join(g.PluginFolder, entry.Name())
		p := internalPlugin{
			Cmd: &exec.Cmd{
				Stderr: os.Stderr,
				Path:   filePath,
				// ToDo: maybe add something as arg to indicate that the binary
				//       should be started in "internalPlugin-mode".
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
	// Note: this 'close' is only a "please close" to the plugins.
	// As there may be plugins which want to run much longer
	// (e.g. notifications after some time).
	err := g.Send("close", nil)
	if err != nil {
		return err
	}

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
