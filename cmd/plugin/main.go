package main

import (
	"fmt"
	"github.com/aligator/goplug/cmd/server/plugin"
	"os"
	"strconv"
)

func main() {
	p := plugin.New("SlowPrintPlugin")
	err := p.Register()
	if err != nil {
		panic(err)
	}

	logger := p.Logger()
	p.OnDoPrint(func(toPrint string) error {
		//time.Sleep(1 * time.Second)

		// This simulates a "function" -> send fnRand and get result fnRand.
		// Will be made more easy when it works.
		p.WG.Add(1)
		res := make(chan int)
		p.RegisterCommand("fnRand", func() interface{} {
			var val int
			return &val
		}, func(message interface{}) error {
			val := message.(*int)
			res <- *val
			return nil
		})
		p.Send("fnRand", nil)
		val := <-res
		p.WG.Done()

		err := p.Print("This is the SlowPrintPlugin " + toPrint + " " + strconv.Itoa(val))
		return err
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
	fmt.Fprint(os.Stderr, "eeeennnd")
	if err != nil {
		logger.Println(fmt.Sprint(err))
	}
}
