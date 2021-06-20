package actions

import (
	"math/rand"
	"time"

	"github.com/aligator/goplug/example/host/lol"
)

type App struct {
	isSeeded bool
}

//goplug:generate
func (a App) GetRandomInt(n some.Test) (int, error) {
	if !a.isSeeded {
		rand.Seed(time.Now().UnixNano())
		a.isSeeded = true
	}

	return rand.Intn(n.N), nil
}
