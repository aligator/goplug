package goplug

import "fmt"

type HostControl struct {
	GoPlug GoPlug
}

func (a *HostControl) Print(args string, reply *interface{}) error {
	_, err := fmt.Print(args)
	return err
}
