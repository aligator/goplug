package main

import (
	"fmt"
	"github.com/aligator/goplug/cmd/server/plugin"
)

func main() {
	p := plugin.New("FastPrintPlugin")
	err := p.Register()
	if err != nil {
		panic(err)
	}

	logger := p.Logger()
	p.OnDoPrint(func(toPrint string) error {
		return p.Print("This is the FastPrintPlugin " + toPrint)
	})

	p.OnAllInitialized(func() error {
		logger.Println("All plugins initialized")
		return nil
	})

	p.OnShouldClose(func() error {
		p.Close()
		return nil
	})

	logger.Println("start RUN")
	err = p.Run()
	if err != nil {
		logger.Println(fmt.Sprint(err))
	}
}
