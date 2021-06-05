package main

import (
	"goplug/cmd/server/plugin"
	"time"
)

func main() {
	p := plugin.TestPlugin{}
	p.Register("TestPlugin")
	time.Sleep(5 * time.Second)
	p.Print("Hello World!!!!")
}
