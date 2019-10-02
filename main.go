package main

import (
	"./app"
)

func main() {
	s := &app.Server{}

	s.Run()
}
