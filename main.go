package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/orbs-network/boyarin/config"
	"github.com/orbs-network/boyarin/strelets"
	"os"
)

func printHelp() {
	fmt.Println("strelets provision-virtual-chain [params]")
	fmt.Println("strelets remove-virtual-chain [params]")
	fmt.Println()
	flag.Usage()
}

func main() {
	root := "_tmp"

	if len(os.Args) < 2 {
		printHelp()
	}

	ctx := context.Background()

	switch os.Args[1] {
	case "provision-virtual-chain":
		input, err := config.GetProvisionVirtualChainInput(os.Args[2:])
		if err != nil {
			panic(err)
		}

		str := strelets.NewStrelets(root)
		if err := str.ProvisionVirtualChain(ctx, input); err != nil {
			panic(err)
		}
	case "remove-virtual-chain":
		input := config.GetRemoveVirtualChainInput(os.Args[2:])

		str := strelets.NewStrelets(root)
		if err := str.RemoveVirtualChain(ctx, input); err != nil {
			panic(err)
		}
	default:
		printHelp()
	}
}
