package errutil

import (
	"errors"
	"fmt"
)

var (
	ErrCumulated = errors.New("several errors happened")
)

// ErrorList is a special error which
// just combines any amount of errors to one.
// It is fully compatible with errors.Is and errors.As.
type ErrorList []error

func (e ErrorList) Error() string {
	result := "List of errors:\n"
	for _, err := range e {
		if err != nil {
			result += fmt.Sprintf("%v\n", err.Error())
		}
	}
	return result
}

func (e ErrorList) Unwrap() error {
	if len(e) == 1 {
		return nil
	}

	return e[:len(e)-1]
}

func (e ErrorList) Is(target error) bool {
	return errors.Is(e[len(e)-1], target)
}

func (e ErrorList) As(target interface{}) bool {
	return errors.As(e[len(e)-1], target)
}

// Collect errors as long as the errCh is open.
// When it gets closed it sends all errors to the resulting channel
// as one combined error and closes that channel also.
// The resulting error is an ErrorList and therefore
// supports errors.Is and errors.As for all collected errors.
func Collect(errCh <-chan error) <-chan error {
	allCh := make(chan error)
	go func() {
		var errors ErrorList

		for {
			select {
			case err, ok := <-errCh:
				if err != nil {
					errors = append(errors, err)
				}
				if !ok {
					if len(errors) > 0 {
						allCh <- errors
					}
					close(allCh)
					return
				}
			}
		}
	}()

	return allCh
}
