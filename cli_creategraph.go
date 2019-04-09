package crickchain


func (cli *CLI) createGraph(dbFile string, nodes int, edges int) {
	bc := NewBlockchain(dbFile)
	defer bc.db.Close()

	pg := NewProblemGraph(nodes, edges)
	pg.NicePrint(bc)
	bc.AddProblemGraph(pg)
}
