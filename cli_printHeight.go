package crickchain


func (cli *CLI) printHeight(dbFile string, height int) {
	bc := NewBlockchain(dbFile)
	defer bc.db.Close()

	block, _ := bc.GetBlockFromHeight(height)
	block.NicePrint(bc)

}


