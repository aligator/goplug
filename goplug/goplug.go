package goplug

import (
	"encoding/json"
	"fmt"
	"github.com/aligator/checkpoint"
	"github.com/aligator/goplug/errutil"
	"io/fs"
	"io/ioutil"
	"os/exec"
	"path"
	"sync"
)

type PluginType string

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
}

type plugin struct {
	PluginInfo
	*exec.Cmd
}

type GoPlug struct {
	PluginFolder      string
	plugins           []plugin
	registeredPlugins map[string]*plugin
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

	g.registeredPlugins = make(map[string]*plugin)

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

			p := plugin{}

			// Start the plugin with the -init flag.
			// The plugin should return some information about it.
			filePath := path.Join(g.PluginFolder, entry.Name())

			cmd := exec.Command(filePath, "-init")
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

			fmt.Println("info", p.PluginInfo)
		}()
	}

	wg.Wait()
	close(errCh)

	err, _ = <-allErrorsCh
	return err
}
