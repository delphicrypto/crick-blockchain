package crickchain

import (
	"fmt"
)

func (cli *CLI) getDifficulty(dbFile string) {
	bc := NewBlockchain(dbFile)
	defer bc.db.Close()
	target := bc.CurrentTarget(false)
	diff := targetToDifficulty(target)
	rTarget := bc.CurrentTarget(true)
	rDiff := targetToDifficulty(rTarget)
	fmt.Printf("Normal  Target: %064x\r\n", target)
	fmt.Printf("Reduced Target: %064x\r\n", rTarget)
	fmt.Printf("Normal  Difficulty: %d\r\n", diff)	
	fmt.Printf("Reduced Difficulty: %d\r\n", rDiff)
}
