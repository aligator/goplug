package actions

import (
	"fmt"
	"math/rand"
	"time"
)

type App struct {
	isSeeded bool
}

//goplug:generate
func (a App) GetRandomInt(n int) (int, error) {
	if !a.isSeeded {
		rand.Seed(time.Now().UnixNano())
		a.isSeeded = true
	}

	return rand.Intn(n), nil
}

//goplug:generate
func (a App) PrintHello() error {
	fmt.Println("Hellooooooo")
	return nil
}
