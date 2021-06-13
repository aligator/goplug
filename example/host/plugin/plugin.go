package plugin

import (
	"encoding/json"
	"github.com/aligator/goplug/common"
	"github.com/aligator/goplug/goplug"
	"github.com/aligator/goplug/plugin"
	"net/rpc"
	"net/rpc/jsonrpc"
	"os"
)

type TestMetadata struct {
	Command string `json:"command"`
}

type GetRandomIntRequest struct {
	N int `json:"n"`
}

type GetRandomIntResponse struct {
	Rand int `json:"rand"`
}

type Plugin struct {
	plugin         plugin.Plugin
	client         *rpc.Client
	subCommand     string
	subCommandFunc func(args []string) error
}

func New(info goplug.PluginInfo) Plugin {
	return Plugin{
		plugin: plugin.Plugin{
			PluginInfo: info,
		},
	}
}

// SetSubCommand - for this example support only one subcommand.
// This is host implementation specific
func (p *Plugin) SetSubCommand(name string, subCommand func(args []string) error) error {
	meta := TestMetadata{
		Command: name,
	}

	metaJson, err := json.Marshal(meta)
	if err != nil {
		return err
	}

	p.subCommandFunc = subCommand
	p.subCommand = name
	p.plugin.Metadata = metaJson

	return nil
}

func (p *Plugin) Run() {
	p.client = jsonrpc.NewClient(common.CombinedReadWriter{
		In:  os.Stdin,
		Out: os.Stdout,
	})

	p.plugin.Init()

	if os.Args[1] == p.subCommand {
		err := p.subCommandFunc(os.Args[1:])
		if err != nil {
			panic(err)
		}
	}
}

func (p *Plugin) GetRandomInt(n int) int {
	response := GetRandomIntResponse{}
	err := p.client.Call("Host.GetRandomInt", GetRandomIntRequest{
		N: n,
	}, &response)
	if err != nil {
		panic(err)
	}
	return response.Rand
}
