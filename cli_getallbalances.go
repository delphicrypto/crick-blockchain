package crickchain

import (
	"fmt"
	"log"
)

func (cli *CLI) getAllBalances(dbFile, walletFile string) {
	bc := NewBlockchain(dbFile)
	UTXOSet := UTXOSet{bc}
	defer bc.db.Close()

	wallets, err := NewWallets(walletFile)
	if err != nil {
		log.Panic(err)
	}
	addresses := wallets.GetAddresses()

	for _, address := range addresses {
		balance := 0
		pubKeyHash := Base58Decode([]byte(address))
		pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
		UTXOs := UTXOSet.FindUTXO(pubKeyHash)

		for _, out := range UTXOs {
			balance += out.Value
		}
		fmt.Printf("Balance of '%s': %d\n", address, balance)
	}
}
