package main

import (
	"github.com/koan6gi/notifier/internal/app"
	"log"
)

func main() {
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
