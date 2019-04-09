package crickchain

import (
	"fmt"
	"encoding/hex"
	"log"
)

func (cli *CLI) mineblock(dbFile string) *Block {
	bc := NewBlockchain(dbFile)
	defer bc.db.Close()
	var txs []*Transaction
	
	newBlock := bc.MineBlock(txs, []byte{}, []int{}, []byte{})
	fmt.Println("Block mined classically")
	fmt.Printf("New block hash: %x\r\n", newBlock.Hash)
	return newBlock
}

func (cli *CLI) mineblockWithNewProblem(dbFile string, nodes int, density float64) *Block {
	bc := NewBlockchain(dbFile)
	defer bc.db.Close()

	kclique := []int{}
	edges := int(float64(nodes*(nodes-1)/2) * density)
	pg := NewProblemGraph(nodes, edges)
	bc.AddProblemGraph(pg)
	//we mine the problem with an initial solution
	for k := 8; k >= 3; k-- {
		kclique = pg.FindKClique(k)
		if len(kclique) == k {
			break
		}
	}
	
	var txs []*Transaction
	newBlock := bc.MineBlock(txs, pg.Hash, kclique, pg.Hash)

	fmt.Println("Block mined with problem")
	fmt.Printf("New block hash: %x\r\n", newBlock.Hash)
	return newBlock
}

func (cli *CLI) mineblockWithSolution(dbFile string, pgHash string) *Block {
	bc := NewBlockchain(dbFile)
	defer bc.db.Close()
	
	hash, err := hex.DecodeString(pgHash)
	if err != nil {
		fmt.Println("Invalid problem hash")
		log.Panic(err)
	}
	pg, err := bc.GetProblemGraphFromHash(hash)
	if err != nil {
		fmt.Println("Invalid problem hash")
		log.Panic(err)
	}

	height := bc.GetBestHeight()
	bestSolution := bc.GetBestSolution(&pg, height)
	kclique := pg.FindKClique(len(bestSolution) + 1)

	var txs []*Transaction

	newBlock := bc.MineBlock(txs, hash, kclique, []byte{})
	fmt.Println("Block mined with solution")
	fmt.Printf("New block hash: %x\r\n", newBlock.Hash)
	return newBlock
}

func (cli *CLI) mineblockParallel(dbFile string) *Block {
	bc := NewBlockchain(dbFile)
	defer bc.db.Close()
	var txs []*Transaction
	height := bc.GetBestHeight()
	hashes := bc.GetProblemGraphHashes()
	if !(len(hashes) > 0) {
		log.Panic("No Problem to mine solution for. Create one first")
	}
	var bestPG *ProblemGraph
	var bestSol []int
	bestRatio := 100.00
	var solHash []byte
	for _, h := range hashes {
		pg, err := bc.GetProblemGraphFromHash(h)
		if err != nil {
			log.Panic(err)
		}
		//evaluate expected difficulty for solution
		sol := bc.GetBestSolution(&pg, height)
		expected := float64(pg.Graph.Order() -1) * pg.Graph.Density()
		ratio := float64(len(sol))/expected
		fmt.Println(ratio)
		if ratio < bestRatio {
			bestRatio = ratio
			bestPG = &pg
			solHash = pg.Hash
			bestSol = sol
		}
		
	}

	//blockChannel := make(chan *Block, 1)
	// go func() {
	// 	newBlock := bc.MineBlock(txs, []byte{}, []int{}, []byte{})
	// 	fmt.Println("Block mined classically")
 //        blockChannel <- newBlock
 //    }()
 //    go func() {
 //    	kclique := bestPG.FindKClique(len(bestSol) + 1)
	// 	newBlock := bc.MineBlock(txs, solHash, kclique, []byte{})
	// 	fmt.Println("Block mined with solution")
		
 //        blockChannel <- newBlock
 //    }()
 //    select {
	//     case block := <-blockChannel:
	// 		//close(blockChannel)
	// 		fmt.Printf("New block hash: %x\r\n", block.Hash)

	// 		return block
	// }

	kclique := bestPG.FindKClique(len(bestSol) + 1)
	newBlock := bc.MineBlock(txs, solHash, kclique, []byte{})
	fmt.Printf("Block mined with solution to %x\n", solHash)
	return newBlock
}
