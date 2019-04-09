package crickchain

import (
	"fmt"
	"encoding/hex"
)

func (cli *CLI) printProblemGraphs(dbFile string) {
	bc := NewBlockchain(dbFile)
	defer bc.db.Close()

	hashes := bc.GetProblemGraphHashes()

	for i, h := range hashes {
		pg, err := bc.GetProblemGraphFromHash(h)
		if err == nil {
			fmt.Println("Problem ", i)
			pg.NicePrint(bc)			
		} else {
			fmt.Println(err)
		}
	}	
}

func (cli *CLI) printProblemGraph(dbFile string, hash string) {
	bc := NewBlockchain(dbFile)
	defer bc.db.Close()

	h, err := hex.DecodeString(hash)
	if err != nil {
	    panic(err)
	}
	pg, err := bc.GetProblemGraphFromHash(h)
	if err == nil {
		pg.NicePrint(bc)
		text := ProblemToString(pg)
		dbFile := "jsgraph/data/graph.js"
		WriteToFile(dbFile, text)
		textsol := "var cliques = ["

		allSolutions := bc.GetAllSolutions(&pg)
		for i, s := range allSolutions {
			textsol += CliqueToString(s)
			if i < len(allSolutions) - 1{
				textsol += ",\n "
			}
		}

		textsol += "];\n"
		dbFilesol := "jsgraph/data/sol.js"
		WriteToFile(dbFilesol, textsol)
	} else {
		fmt.Println(err)
	}
}	