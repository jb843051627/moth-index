package main

import (
	"log"
	"os"

	"github.com/jb843051627/moth-index/internal/app"
)

func main() {
	cfg := app.ConfigFromEnv()
	if err := app.Run(cfg); err != nil {
		log.New(os.Stderr, "moth-index: ", log.LstdFlags).Fatal(err)
	}
}
