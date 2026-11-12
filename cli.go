package main

import (
	"flag"
	"log"

	"startromagnaapi/config"
	"startromagnaapi/internal/gtfs"
	"startromagnaapi/internal/repository"
)

// Select an execution mode before loading configuration or opening connections.
func runCLI(args []string, serve func(), updateStatic func() error) int {
	flags := flag.NewFlagSet("StartRomagnaAPI", flag.ContinueOnError)
	staticOnly := flags.Bool("update-static-gtfs", false, "Update static GTFS once and exit without starting HTTP, scheduler or realtime updates")
	if err := flags.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return 0
		}
		return 2
	}
	if flags.NArg() != 0 {
		log.Print("Unexpected positional arguments; use --help for usage")
		return 2
	}
	if *staticOnly {
		if err := updateStatic(); err != nil {
			log.Printf("Failed to update static GTFS: %v", err)
			return 1
		}
		return 0
	}
	serve()
	return 0
}

func runStaticGTFSUpdate() error {
	config.LoadStaticGTFSConf()
	repository.InitStaticConnection()
	defer repository.DB_STATIC.Close()
	if err := gtfs.UpdateStatic(); err != nil {
		return err
	}
	log.Print("Static GTFS update completed")
	return nil
}
