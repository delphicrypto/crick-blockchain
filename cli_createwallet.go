package crickchain

import (
	"fmt"
)

func (cli *CLI) createWallet(dbFile string) {
	wallets, _ := NewWallets(dbFile)
	address := wallets.CreateWallet()
	wallets.SaveToFile(dbFile)

	fmt.Printf("Your new address: %s\n", address)
}


func (cli *CLI) quickstart(dbFile, walletFile string) {
	wallets, _ := NewWallets(walletFile)
	address := wallets.CreateWallet()
	wallets.SaveToFile(walletFile)

	fmt.Printf("Your new address: %s\n", address)
	cli.createBlockchain(address, dbFile)
}
