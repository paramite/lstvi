package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/paramite/lstvi/endpoints"
)

func main() {
	memInitCap := flag.Int("initial-capacity", 100, "Initial capacity for data storage.")
	config := flag.String("config", "/etc/lstvi_config.json", "Path to configuration file.")
	flag.Parse()

	dispatcher, err := endpoints.NewDispatcher(*memInitCap, *config)
	if err != nil {
		fmt.Printf("Failed to start server: %s\n", err.Error())
		flag.Usage()
		os.Exit(1)
	}
	dispatcher.Start()
}
