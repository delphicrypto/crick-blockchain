package crickchain

import (
	"fmt"
	"log"
)

func (cli *CLI) startNode(nodeID string, dbFile string, minerAddress string) {
	fmt.Printf("Starting node %s\n", dbFile)
	if len(minerAddress) > 0 {
		if ValidateAddress(minerAddress) {
			fmt.Println("Mining is on. Address to receive rewards: ", minerAddress)
		} else {
			log.Panic("Wrong miner address!")
		}
	}
	StartServer(nodeID, dbFile, minerAddress)
}
