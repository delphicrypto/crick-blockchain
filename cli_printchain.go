package crickchain


func (cli *CLI) printChain(dbFile string) {
	bc := NewBlockchain(dbFile)
	defer bc.db.Close()

	bci := bc.Iterator()
	for {
		block := bci.Next()
		block.NicePrint(bc)

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
}
