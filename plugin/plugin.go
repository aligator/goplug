package plugin

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/aligator/goplug/common"
	"github.com/aligator/goplug/goplug"
	"net/rpc"
	"net/rpc/jsonrpc"
	"os"
)

type Plugin struct {
	goplug.PluginInfo
	client *rpc.Client
}

func (p *Plugin) Init() {
	init := flag.Bool("init", false, "")
	flag.Parse()

	// Return the plugin info on init just using stdout.
	if *init {
		res, err := json.Marshal(p.PluginInfo)
		if err != nil {
			panic(err)
		}
		fmt.Print(string(res))
		os.Exit(0)
	}

	// If it is a one shot plugin, it needs to be able to communicate
	// with the host using rpc to query data.
	if p.PluginType == goplug.OneShot {
		p.client = jsonrpc.NewClient(common.CombinedReadWriter{
			In:  os.Stdin,
			Out: os.Stdout,
		})
	}
}

func (p *Plugin) Call(serviceMethod string, args interface{}, reply interface{}) error {
	return p.client.Call("Host."+serviceMethod, args, reply)
}

func (p *Plugin) Print(text string) error {
	return p.client.Call("HostControl.Print", text, nil)
}
