package main

import (
	"log"

	"Go-Next-WebRTC/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}