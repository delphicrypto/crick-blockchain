package crickchain


func (cli *CLI) printLast(dbFile string) {
	bc := NewBlockchain(dbFile)
	defer bc.db.Close()

	bci := bc.Iterator()
	block := bci.Next()
	block.NicePrint(bc)

}


