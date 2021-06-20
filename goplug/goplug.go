package goplug

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"net/rpc"
	"net/rpc/jsonrpc"
	"os"
	"os/exec"
	"path"
	"sync"

	"github.com/aligator/checkpoint"
	"github.com/aligator/goplug/common"
	"github.com/aligator/goplug/errutil"
)

type PluginType string

var (
	ErrPluginDoesNotExist = errors.New("plugin does not exist")
	ErrCallingPlugin      = errors.New("could not call the plugin")
)

const (
	// OneShot is a plugin which gets called when a specific event happens.
	// This can be for example a subcommand in command line tools.
	// This type of plugin can query information from the host. And does only
	// run when needed.
	OneShot = PluginType("one_shot")

	// DataSource is a plugin which provides data from any datasource.
	// This type of plugin gets queried by the host and runs in the background
	// as long as the host needs it.
	DataSource = PluginType("data_source")
)

type PluginInfo struct {
	ID         string     `json:"id"`
	PluginType PluginType `json:"plugin_type"`

	// Metadata contains additional information which is host-specific.
	// It may just be another json string.
	Metadata string `json:"metadata"`
}

type plugin struct {
	PluginInfo
	filePath string
}

type GoPlug struct {
	PluginFolder string
	Host         Host
	Actions      interface{}

	plugins             []plugin
	oneShotPlugins      map[string]*plugin
	oneShotPluginsMutex sync.Mutex
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

// Init initializes and starts all plugins.
// It blocks until all plugins are initialized.
func (g *GoPlug) Init() error {
	entries, err := ioutil.ReadDir(g.PluginFolder)
	if err != nil {
		return err
	}

	g.oneShotPluginsMutex.Lock()
	g.oneShotPlugins = make(map[string]*plugin)
	g.oneShotPluginsMutex.Unlock()

	errCh := make(chan error)
	allErrorsCh := errutil.Collect(errCh)

	wg := sync.WaitGroup{}
	wg.Add(len(entries))
	for i := range entries {
		entry := entries[i]
		go func() {
			defer wg.Done()

			if !isValidPlugin(entry) {
				return
			}

			// Start the plugin with the -init flag.
			// The plugin should return some information about it.
			filePath := path.Join(g.PluginFolder, entry.Name())

			p := plugin{
				filePath: filePath,
			}

			cmd := exec.Command(filePath, "-init")
			cmd.Stderr = os.Stderr
			res, err := cmd.Output()
			if err != nil {
				errCh <- checkpoint.From(err)
				return
			}

			err = json.Unmarshal(res, &p.PluginInfo)
			if err != nil {
				errCh <- checkpoint.From(err)
				return
			}

			if p.PluginType != OneShot {
				log.Println(p.ID, "- currently only one_shot plugins are supported")
				return
			}

			g.oneShotPluginsMutex.Lock()
			g.oneShotPlugins[p.ID] = &p
			g.oneShotPluginsMutex.Unlock()

			err = g.Host.RegisterOneShot(p.PluginInfo, func(args []string) error {
				// Run
				return g.oneShot(p.ID, args)
			})
			if err != nil {
				errCh <- checkpoint.From(err)
				return
			}
		}()
	}

	wg.Wait()
	close(errCh)

	err, _ = <-allErrorsCh
	return err
}

func (g *GoPlug) oneShot(ID string, args []string) error {
	if p, ok := g.oneShotPlugins[ID]; !ok {
		return checkpoint.From(fmt.Errorf("PluginID: %v: %w", ID, ErrPluginDoesNotExist))
	} else {
		// Start rpc.
		cmd := exec.Command(p.filePath, args...)

		outPipe, err := cmd.StdoutPipe()
		if err != nil {
			return checkpoint.Wrap(fmt.Errorf("PluginID: %v: %w", ID, err), ErrCallingPlugin)
		}

		inPipe, err := cmd.StdinPipe()
		if err != nil {
			return checkpoint.Wrap(fmt.Errorf("PluginID: %v: %w", ID, err), ErrCallingPlugin)
		}

		cmd.Stderr = os.Stderr

		codec := jsonrpc.NewServerCodec(common.CombinedReadWriter{
			In:  outPipe,
			Out: inPipe,
		})

		s := rpc.NewServer()

		err = s.RegisterName("Host", g.Actions)
		if err != nil {
			return checkpoint.Wrap(fmt.Errorf("PluginID: %v: %w", ID, err), ErrCallingPlugin)
		}

		err = s.RegisterName("HostControl", &HostControl{})
		if err != nil {
			return checkpoint.Wrap(fmt.Errorf("PluginID: %v: %w", ID, err), ErrCallingPlugin)
		}

		go func() {
			s.ServeCodec(codec)
		}()

		// Start plugin
		err = cmd.Run()
		if err != nil {
			return err
		}

		return nil
	}
}
