package main

import (
	"fmt"
	"github.com/aligator/goplug/cmd/server/plugin"
	"os"
	"strconv"
	"time"
)

func main() {
	p := plugin.New("SlowPrintPlugin")
	err := p.Register()
	if err != nil {
		panic(err)
	}

	logger := p.Logger()
	p.OnDoPrint(func(toPrint string) error {
		time.Sleep(1 * time.Second)

		err = p.Print("This is the SlowPrintPlugin " + toPrint + " " + strconv.Itoa(*random))
		p.Close()
		return err
	})

	p.OnAllInitialized(func() error {
		logger.Println("All plugins initialized")
		return nil
	})

	logger.Println("start RUN")
	err = p.Run()
	fmt.Fprint(os.Stderr, "eeeennnd")
	if err != nil {
		logger.Println(fmt.Sprint(err))
	}
}
