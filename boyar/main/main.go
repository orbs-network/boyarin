package main

import (
	"flag"
	"fmt"
	"github.com/orbs-network/boyarin/boyar"
	"os"
	"time"
)

func main() {
	configUrlPtr := flag.String("config-url", "", "http://my-config/config.json")
	keyPairConfigPathPtr := flag.String("keys", "", "path to public/private key pair in json format")

	daemonizePtr := flag.Bool("daemonize", false, "do not exit the program and keep polling for changes")
	pollingIntervalPtr := flag.Uint("polling-interval", 60, "how often to poll for configuration in daemon mode (in seconds)")

	flag.Parse()

	if *daemonizePtr {
		successfulConfigHash := ""

		for true {
			if configHash, err := boyar.RunOnce(*keyPairConfigPathPtr, *configUrlPtr, successfulConfigHash); err != nil {
				fmt.Println("ERROR:", err)
				fmt.Println("Latest successful configuration", successfulConfigHash)
			} else {
				fmt.Println("Successfully updated to latest configuration:", configHash)
				successfulConfigHash = configHash
			}
			<-time.After(time.Duration(*pollingIntervalPtr) * time.Second)
		}
	} else {
		if _, err := boyar.RunOnce(*keyPairConfigPathPtr, *configUrlPtr, ""); err != nil {
			fmt.Println("ERROR:", err)
			os.Exit(1)
		}
	}

}
