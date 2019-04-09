package crickchain

import (
	"fmt"
	"log"
)

func (cli *CLI) listAddresses(dbFile string) {
	wallets, err := NewWallets(dbFile)
	if err != nil {
		log.Panic(err)
	}
	addresses := wallets.GetAddresses()

	for _, address := range addresses {
		fmt.Println(address)
	}
}
