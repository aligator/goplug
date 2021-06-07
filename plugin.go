package goplug

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
)

type commandFn = func(data []byte) error

// RegisterMessage is the message which is the
// payload for the 'register' command.
type RegisterMessage struct {
	ID string `json:"id"`
}

// Plugin is the base struct to be used to build plugins.
// It already contains the basic methods to communicate with
// the plugin-host.
type Plugin struct {
	ID       string
	commands map[string]commandFn

	shouldCloseSig chan bool
	actualCloseSig chan bool
}

func (p *Plugin) ShouldClose() bool {
	select {
	case _, open := <-p.shouldCloseSig:
		return !open
	default:
		return false
	}
}

func (p *Plugin) Close() {
	close(p.actualCloseSig)
}

// Register registers the plugin with the given ID.
// This ID has to be unique. If there is already another plugin
// registered with the same id, the registration will fail and the
// plugin process killed.
//
// Register has to be the first message sent by any plugin.
func (p *Plugin) Register() error {
	return p.Send("register", RegisterMessage{
		ID: p.ID,
	})
}

// Send any command with any json-marshal-able payload.
func (p *Plugin) Send(cmd string, payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	// As GoPlug communicates over stdout, a simple fmt.Println is sufficient.
	_, err = fmt.Println(cmd, string(data))
	return err
}

// Logger returns a new log.Logger configured to log through
// the plugin messaging system. This is needed stdout is already
// used for communication. When logging using this logger, the logs are
// sent using the 'log' command.
func (p *Plugin) Logger() *log.Logger {
	return log.New(&PluginLogWriter{
		Plugin: p,
	}, p.ID+" ", 0)
}

// RegisterCommand can be used to register commands, this plugin listens to.
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
//  data := message.(*DoPrintMessage)
//	return listener(data.Text)
func (p *Plugin) RegisterCommand(cmd string, factory func() interface{}, listener func(message interface{}) error) {
	if p.commands == nil {
		p.commands = make(map[string]commandFn)
	}

	p.commands[cmd] = func(message []byte) error {
		if factory == nil {
			return listener(nil)
		}

		data := factory()
		err := json.Unmarshal(message, &data)
		if err != nil {
			return err
		}

		return listener(data)
	}
}

// Run marks the plugin as initialized and starts the message-reading loop.
// You have to setup all events before calling this method.
// This function only exits on error of if the 'close' command was received.
func (p *Plugin) Run() error {
	p.shouldCloseSig = make(chan bool)
	p.actualCloseSig = make(chan bool)
	p.Send("initialized", nil)

	reader := bufio.NewReader(os.Stdin)

	go func() {
		<-p.actualCloseSig
		// ToDo: Could not find a better way to close the reader.ReadLine
		//       loop instantly.
		os.Exit(0)
	}()

	for {
		message, _, err := reader.ReadLine()
		if err != nil {
			fmt.Println(err)
			return err
		}

		cmd, data, err := parseMessage(string(message))
		if err != nil {
			return err
		}

		if cmd == "kill" {
			return nil
		}

		if cmd == "close" {
			close(p.shouldCloseSig)
		}

		if p.commands == nil {
			continue
		}

		if listener, ok := p.commands[cmd]; ok {
			go func() {
				_ = listener(data)
			}()
			// TODO: error handling
			//err := listener(data)
			//if err != nil {
			//	return err
			//}
		}
	}

	return nil
}

// OnAllInitialized is an event which notifies that all plugins are initialized.
func (p *Plugin) OnAllInitialized(listener func() error) {
	p.RegisterCommand("allInitialized", nil, func(message interface{}) error {
		return listener()
	})
}
