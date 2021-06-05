package main

import (
	"encoding/json"
	"fmt"
	"goplug"
)

func main() {
	plug := goplug.GoPlug{
		PluginFolder: "./cmd/plugin-bin",
	}

	plug.RegisterOnCommand(func(cmd string, data []byte) error {
		if cmd == "print" {
			message := ""
			err := json.Unmarshal(data, &message)
			if err != nil {
				return err
			}

			fmt.Println(message)
		}

		return nil
	})

	err := plug.Run()
	if err != nil {
		panic(err)
	}

	plug.Wait()
}
