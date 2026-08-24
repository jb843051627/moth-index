package app

import "os"

type Config struct {
	DBPath  string
	Address string
	Workers int
}

func ConfigFromEnv() Config {
	path := os.Getenv("MOTH_INDEX_DB")
	if path == "" {
		path = "./moth-index.db"
	}
	address := os.Getenv("MOTH_INDEX_ADDR")
	if address == "" {
		address = ":8080"
	}
	workers := 2
	return Config{DBPath: path, Address: address, Workers: workers}
}

func (c Config) Valid() bool { return c.DBPath != "" && c.Address != "" && c.Workers > 0 }
