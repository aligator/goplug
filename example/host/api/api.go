package api

import (
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"time"
)

type App struct {
	isSeeded  bool
	lastHello int
}

// GetRandomInt returns, a non-negative pseudo-random number in [0,n) from the
// default Source. Returns an error if n <= 0.
//goplug:generate
func (a *App) GetRandomInt(n int) (res int, err error) {
	// As rand.Intn panics if n <= 0, but this would crash the host,
	// recover from it.
	defer func() {
		if e := recover(); e != nil {
			fmt.Println("Recovered in GetRandomInt", e)
			err = errors.New("n <= 0 is not allowed")
		}
	}()

	if !a.isSeeded {
		rand.Seed(time.Now().UnixNano())
		a.isSeeded = true
	}

	return rand.Intn(n), nil
}

// PrintHello to stdout.
//goplug:generate
func (a *App) PrintHello() error {
	fmt.Println("Hellooooooo", strconv.Itoa(a.lastHello))
	a.lastHello++
	return nil
}
