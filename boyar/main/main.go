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

	configCache := make(boyar.BoyarConfigCache)

	if *daemonizePtr {
		for true {
			if err := boyar.RunOnce(*keyPairConfigPathPtr, *configUrlPtr, configCache); err != nil {
				fmt.Println("ERROR:", err)
			}

			for vcid, hash := range configCache {
				fmt.Println(fmt.Sprintf("Latest successful configuration for vchain %s: %s", vcid, hash))
			}

			<-time.After(time.Duration(*pollingIntervalPtr) * time.Second)
		}
	} else {
		if err := boyar.RunOnce(*keyPairConfigPathPtr, *configUrlPtr, configCache); err != nil {
			fmt.Println("ERROR:", err)
			os.Exit(1)
		}
	}

}
