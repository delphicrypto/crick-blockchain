package crickchain

import "fmt"

func (cli *CLI) reindexUTXO(dbFile string) {
	bc := NewBlockchain(dbFile)
	defer bc.db.Close()
	
	UTXOSet := UTXOSet{bc}
	UTXOSet.Reindex()

	count := UTXOSet.CountTransactions()
	fmt.Printf("Done! There are %d transactions in the UTXO set.\n", count)
}
