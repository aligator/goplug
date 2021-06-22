package goplug

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/rpc"
	"net/rpc/jsonrpc"
	"os"

	"github.com/aligator/goplug/common"
)

// Client is the basis of all Plugins.
// To use it the Init method has to be called.
// Which starts the client.
type Client struct {
	PluginInfo
	client *rpc.Client
}

// Init starts the client and connects to jsonrpc.
// If the flag "-init" is passed, it only returns its
// plugin information to stdout as json and exits.
func (c *Client) Init() error {
	init := flag.Bool("init", false, "")
	flag.Parse()

	// Return the plugin info on init just using stdout.
	if *init {
		res, err := json.Marshal(c.PluginInfo)
		if err != nil {
			panic(err)
		}
		fmt.Print(string(res))
		os.Exit(0)
	}

	// If it is a one shot plugin, it needs to be able to communicate
	// with the host using rpc to query data.
	if c.PluginType == OneShot {
		c.client = jsonrpc.NewClient(common.CombinedReadWriter{
			In:  os.Stdin,
			Out: os.Stdout,
		})
	}

	return nil
}

// Call can be used to execute commands which are sent to the host.
func (c *Client) Call(serviceMethod string, args interface{}, reply interface{}) error {
	return c.client.Call("Host."+serviceMethod, args, reply)
}
