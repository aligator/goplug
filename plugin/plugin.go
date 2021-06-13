package plugin

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/aligator/goplug/goplug"
	"os"
)

type Plugin struct {
	goplug.PluginInfo
}

func (p Plugin) Init() {
	init := flag.Bool("init", false, "")
	flag.Parse()

	// Return the plugin info on init.
	if *init {
		res, err := json.Marshal(p.PluginInfo)
		if err != nil {
			panic(err)
		}
		fmt.Print(string(res))
		os.Exit(0)
	}
}
