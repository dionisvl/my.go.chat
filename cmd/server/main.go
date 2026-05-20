package main

import (
	"mygochat/internal/app"
)

func main() {
	if err := app.New().Run(); err != nil {
		panic(err)
	}
}
