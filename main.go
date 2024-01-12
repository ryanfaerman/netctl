package main

import (
	"errors"
	"fmt"
)

var (
	ErrNeedsCallsign = errors.New("user needs a callsign")
	ErrNeedsName     = errors.New("user needs a name")
)

func main() {
	err := someFunc()

	if errors.Is(err, ErrNeedsCallsign) {
		fmt.Println("please provide a callsign")
	}

	if errors.Is(err, ErrNeedsName) {
		fmt.Println("please provide a name")
	}

}

func someFunc() error {
	return errors.Join(ErrNeedsCallsign, ErrNeedsName)
}
