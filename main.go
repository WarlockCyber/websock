package main

import (
	"log"
)

func main() {
	if err := initConfig(); err != nil {
		log.Fatal(err)
	}

	initWS()

	if err := startServer(); err != nil {
		log.Fatal(err)
	}
}
