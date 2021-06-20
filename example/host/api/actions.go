package actions

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"
)

type App struct {
	isSeeded  bool
	lastHello int
}

//goplug:generate
func (a *App) GetRandomInt(n int) (int, error) {
	if !a.isSeeded {
		rand.Seed(time.Now().UnixNano())
		a.isSeeded = true
	}

	return rand.Intn(n), nil
}

//goplug:generate
func (a *App) PrintHello() error {
	fmt.Println("Hellooooooo", strconv.Itoa(a.lastHello))
	a.lastHello++
	return nil
}
