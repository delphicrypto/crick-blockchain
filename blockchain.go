package crickchain

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"os"
	"math/big"
	"github.com/boltdb/bolt"
)

const (
	dbFile = "blockchain_%s.db"
	blocksBucket = "blocks"
	problemsBucket = "problems"
	genesisCoinbaseData = "The Times 03/Jan/2009 Chancellor on brink of second bailout for banks"
	blocksPerTargetUpdate = 64
	initialTargetBits = 16
	initialReducedTargetBits = 12
	maxTargetLength = 78
	targetBlocksPerMinute = 6
	nanosecondsPerMinute = 60 * 1e9
	maxTargetChange = 4.0
	eta = 0.25
)

var targetTable = map[int]map[string]*big.Int{
	0 : map[string]*big.Int{
		"normal" : targetFromTargetBits(initialTargetBits),
		"reduced": targetFromTargetBits(initialReducedTargetBits),
	},
}

// Blockchain implements interactions with a DB
type Blockchain struct {
	tip []byte
	db  *bolt.DB
}

// CreateBlockchain creates a new blockchain DB
func CreateBlockchain(address, filename string) *Blockchain {
	if dbExists(filename) {
		fmt.Println("Blockchain already exists.")
		bc := NewBlockchain(filename)
		return bc
	}

	var tip []byte

	cbtx := NewCoinbaseTX(address, genesisCoinbaseData)
	//pg := NewProblemGraph(20, 85)//remember to add it to the blockchain db at the end!
	genesis := NewGenesisBlock(cbtx)
	
	db, err := bolt.Open(filename, 0600, nil)
	if err != nil {
		log.Panic(err)
	}
	err = db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucket([]byte(blocksBucket))
		if err != nil {
			log.Panic(err)
		}
		err = b.Put(genesis.Hash, genesis.Serialize())
		if err != nil {
			log.Panic(err)
		}
		err = b.Put([]byte("l"), genesis.Hash)
		if err != nil {
			log.Panic(err)
		}
		tip = genesis.Hash
		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket([]byte(problemsBucket))
		if err != nil {
			log.Panic(err)
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	bc := Blockchain{tip, db}
	//bc.AddProblemGraph(pg)

	return &bc
}

// NewBlockchain creates a new Blockchain with genesis Block
func NewBlockchain(filename string) *Blockchain {
	if dbExists(filename) == false {
		fmt.Println("No existing blockchain found. Create one first.")
		os.Exit(1)
	}

	var tip []byte
	db, err := bolt.Open(filename, 0600, nil)
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		tip = b.Get([]byte("l"))

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	bc := Blockchain{tip, db}

	return &bc
}

//CloseDB exposes the close database function
func (bc *Blockchain) CloseDB() {
	bc.db.Close()
}

// AddBlock saves the block into the blockchain
func (bc *Blockchain) AddBlock(block *Block) {
	if block.Validate(bc) {
		err := bc.db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(blocksBucket))
			blockInDb := b.Get(block.Hash)

			if blockInDb != nil {
				return nil
			}

			blockData := block.Serialize()
			err := b.Put(block.Hash, blockData)
			if err != nil {
				log.Panic(err)
			}

			lastHash := b.Get([]byte("l"))
			lastBlockData := b.Get(lastHash)
			lastBlock := DeserializeBlock(lastBlockData)

			if block.Height > lastBlock.Height {
				err = b.Put([]byte("l"), block.Hash)
				if err != nil {
					log.Panic(err)
				}
				bc.tip = block.Hash
			}

			return nil
		})
		if err != nil {
			log.Panic(err)
		}
	}	
}

// FindTransaction finds a transaction by its ID
func (bc *Blockchain) FindTransaction(ID []byte) (Transaction, error) {
	bci := bc.Iterator()

	for {
		block := bci.Next()

		for _, tx := range block.Transactions {
			if bytes.Compare(tx.ID, ID) == 0 {
				return *tx, nil
			}
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return Transaction{}, errors.New("Transaction is not found")
}

func (bc *Blockchain) GetBlocksPerTargetUpdate() int {
	return blocksPerTargetUpdate
}



// FindUTXO finds all unspent transaction outputs and returns transactions with spent outputs removed
func (bc *Blockchain) FindUTXO() map[string]TXOutputs {
	UTXO := make(map[string]TXOutputs)
	spentTXOs := make(map[string][]int)
	bci := bc.Iterator()

	for {
		block := bci.Next()

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

		Outputs:
			for outIdx, out := range tx.Vout {
				// Was the output spent?
				if spentTXOs[txID] != nil {
					for _, spentOutIdx := range spentTXOs[txID] {
						if spentOutIdx == outIdx {
							continue Outputs
						}
					}
				}

				outs := UTXO[txID]
				outs.Outputs = append(outs.Outputs, out)
				UTXO[txID] = outs
			}

			if tx.IsCoinbase() == false {
				for _, in := range tx.Vin {
					inTxID := hex.EncodeToString(in.Txid)
					spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout)
				}
			}
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return UTXO
}

// Iterator returns a BlockchainIterat
func (bc *Blockchain) Iterator() *BlockchainIterator {
	bci := &BlockchainIterator{bc.tip, bc.db}

	return bci
}

// GetBestHeight returns the height of the latest block
func (bc *Blockchain) GetBestHeight() int {
	var lastBlock Block

	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash := b.Get([]byte("l"))
		blockData := b.Get(lastHash)
		lastBlock = *DeserializeBlock(blockData)

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	return lastBlock.Height
}

// GetBlockFromHash finds a block by its hash and returns it
func (bc *Blockchain) GetBlockFromHash(blockHash []byte) (Block, error) {
	var block Block

	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))

		blockData := b.Get(blockHash)

		if blockData == nil {
			return errors.New("Block is not found.")
		}

		block = *DeserializeBlock(blockData)

		return nil
	})
	if err != nil {
		return block, err
	}

	return block, nil
}

// GetBlockFromHeight finds a block by its height and returns it
func (bc *Blockchain) GetBlockFromHeight(height int) (Block, error) {
	var block Block

	hashes := bc.GetBlockHashes()
	if height > len(hashes) - 1 {
		return block, errors.New("Block is not found.")
	}
	blockHash := hashes[len(hashes) - 1 - height]
	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))

		blockData := b.Get(blockHash)

		if blockData == nil {
			return errors.New("Block is not found.")
		}

		block = *DeserializeBlock(blockData)

		return nil
	})
	if err != nil {
		return block, err
	}

	return block, nil
}

// GetBlockHashes returns a list of hashes of all the blocks in the chain
func (bc *Blockchain) GetBlockHashes() [][]byte {
	var blocks [][]byte
	bci := bc.Iterator()

	for {
		block := bci.Next()
		blocks = append(blocks, block.Hash)

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return blocks
}

// GetProlemGraphFromHash finds a Problemgraph by its hash and returns it
func (bc *Blockchain) GetProblemGraphFromHash(pgHash []byte) (ProblemGraph, error) {
	var pg ProblemGraph

	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(problemsBucket))

		pgData := b.Get(pgHash)

		if pgData == nil {
			return errors.New("Problem is not found.")
		}

		pg = *DeserializeProblemGraph(pgData)

		return nil
	})
	if err != nil {
		return pg, err
	}

	return pg, nil
}

// GetProblemGraphHashes returns a list of hashes of all the problems in the chain
func (bc *Blockchain) GetProblemGraphHashes() [][]byte {
	var problems [][]byte
	bci := bc.Iterator()

	for {
		block := bci.Next()
		if len(block.ProblemGraphHash) > 0 {
			problems = append(problems, block.ProblemGraphHash)
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return problems
}




// GetAllSolutions returns the all solutions found in the blockchain for the given problemgraph
func (bc *Blockchain) GetAllSolutions(pg *ProblemGraph) [][]int {
	bci := bc.Iterator()
	allSolutions := [][]int{}
	for {
		block := bci.Next()
		if Equal(block.SolutionHash, pg.Hash) {
			allSolutions = append(allSolutions, block.Solution)
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return allSolutions
}

// GetBestSolution returns the best solution found in the blockchain for the given problemgraph
func (bc *Blockchain) GetBestSolution(pg *ProblemGraph, height int) []int {
	bci := bc.Iterator()
	bestSol := []int{}
	for {
		block := bci.Next()
		if Equal(block.SolutionHash, pg.Hash) && (len(block.Solution) > len(bestSol) && (block.Height <= height)) {
			bestSol = block.Solution
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return bestSol
}



//GetNumberOfBlocks returns the number of blocks without solution. If reduced is true, returns the number of blocks with solution.
func (bc *Blockchain) GetNumberOfBlocks(from int, to int, reduced bool) int {
	n := 0
	if (from > to) || (to > bc.GetBestHeight()) || (from < 0) {
		fmt.Println("Invalid parameters.", from, to, bc.GetBestHeight())
		os.Exit(1)
	} 
	for h := to; h >= from; h-- {
		block, err := bc.GetBlockFromHeight(h)
		if err != nil {
			log.Panic(err)
		}
		if (block.HasValidSolution(bc) && reduced) ||  (!reduced && !block.HasValidSolution(bc)) {
			n += 1
		} 		
	}
	return n	
}

//TimeForBlocks returns the time spent mining block. If reduced is true, returns time sent for blocks at reduced difficulty
func (bc *Blockchain) TimeForBlocks(from int, to int, reduced bool) int64 {
	t := int64(0)
	if (from > to) || (to > bc.GetBestHeight()) || (from < 0) {
		fmt.Println("Invalid parameters.", from, to, bc.GetBestHeight())
		os.Exit(1)
	} 
	for h := to; h > from; h-- {
		block, err := bc.GetBlockFromHeight(h)
		if err != nil {
			log.Panic(err)
		}
		if (block.HasValidSolution(bc) && reduced) ||  (!reduced && !block.HasValidSolution(bc)) {
			prevBlock, err := bc.GetBlockFromHeight(h-1)
			if err != nil {
				log.Panic(err)
		}
			t += block.Timestamp - prevBlock.Timestamp
		} 		
	}
	return t	
}


//CalculateTarget return the new target. If reduced is true, returns the reduced target
func (bc *Blockchain) CalculateTarget(height int, reduced bool) *big.Int {
	var prevTarget *big.Int
	var newTarget *big.Int
	//var tBits int

	base := height/blocksPerTargetUpdate
	if val, ok := targetTable[base]; ok {
		if reduced {
			return val["reduced"] 
		}
	    return val["normal"] 
	}

	// if height < blocksPerTargetUpdate {
	// 	if reduced {
	// 		tBits = initialReducedTargetBits
	// 	} else {
	// 		tBits = initialTargetBits
	// 	}
	// 	initialTarget := targetFromTargetBits(tBits)
	// 	return initialTarget
	// }

	hashes := bc.GetBlockHashes()
	total := len(hashes)
	index := ((height-1)/blocksPerTargetUpdate) * blocksPerTargetUpdate //this return only integer part of ratio since i'm divindg two integers
	baseBlock, _ := bc.GetBlockFromHash(hashes[total -1 - index])//this block is the first block in the batch of blocks we need to calculate difficulty
	//This iw rong now
	rest := height%blocksPerTargetUpdate
	if rest != 0 {
		return bc.CalculateTarget(height - rest, reduced)
	}
	
	lastBlock, _ := bc.GetBlockFromHash(hashes[total - 1 - (blocksPerTargetUpdate + index - 1)])//this block is the last block in the batch of blocks we need to calculate difficulty
	tReduced := bc.TimeForBlocks(baseBlock.Height, lastBlock.Height, true)
	tNormal := bc.TimeForBlocks(baseBlock.Height, lastBlock.Height, false)
	etaStar := float64(tReduced)/float64(tNormal)
	
	r := float64(bc.GetNumberOfBlocks(baseBlock.Height, lastBlock.Height, false))/blocksPerTargetUpdate //this is b in the paper

	t := lastBlock.Timestamp - baseBlock.Timestamp
	
	prevTarget = bc.CalculateTarget(height - blocksPerTargetUpdate, false)
	prevDiff := targetToDifficulty(prevTarget)

	timeTarget := nanosecondsPerMinute * blocksPerTargetUpdate / targetBlocksPerMinute


	retarget := (r + (1 - r) * etaStar)/(r + (1 - r) * eta) * (float64(timeTarget) / float64(t))
	if retarget > maxTargetChange {
		retarget = maxTargetChange
	} else if retarget < 1.0/maxTargetChange {
		retarget = 1.0/maxTargetChange
	}

	floatRetarget := new(big.Float).SetFloat64(retarget)
	floatDiff := bigIntToBigFloat(prevDiff)
	
	newFloatDiff := new(big.Float).Mul(floatDiff, floatRetarget)
	newDiff := bigFloatToBigInt(newFloatDiff)
	newTarget = difficultyToTarget(newDiff)
	

	
	//calculate reduced newtarget
	prevTargetReduced := bc.CalculateTarget(height - blocksPerTargetUpdate, true)
	prevDiffReduced := targetToDifficulty(prevTargetReduced)
	
	retargetReduced := eta * retarget - etaStar
	if retargetReduced > maxTargetChange {
		retargetReduced = maxTargetChange
	} else if retargetReduced < 1.0/maxTargetChange {
		retargetReduced = 1.0/maxTargetChange
	}


	newDiffReduced := new(big.Int)

	newDiffReduced.Add(prevDiffReduced, bigFloatToBigInt(new(big.Float).Mul(floatDiff, new(big.Float).SetFloat64(retargetReduced))))
	newTargetReduced := difficultyToTarget(newDiffReduced)
	
	verbose := false
	if verbose {
		fmt.Printf("eta* = %4f\n", etaStar)
		fmt.Printf("b = %4f\n", r)
		fmt.Printf("Rescaling factor: %5f\n", retarget)
		fmt.Printf("new diff %d\n", newDiff)
		fmt.Printf("Rescaling factor reduced: %5f\n", retargetReduced)
		fmt.Printf("new diff reduced %d\n", newDiffReduced)
	}
	targetTable[base] = map[string]*big.Int{}
	targetTable[base]["normal"] = newTarget
	targetTable[base]["reduced"] = newTargetReduced
	if !reduced {
		return newTarget
	}
	return newTargetReduced		
}

//Calculate the new target bits
func (bc *Blockchain) CurrentTarget(reduced bool) *big.Int {
	height := bc.GetBestHeight() + 1
	return bc.CalculateTarget(height, reduced)
}

//GetBlockTarget returns the target for a block
func (bc *Blockchain) GetBlockTarget(height int, solHash []byte, solution []int, pgHash []byte) *big.Int {
	//if solution is valid, use reduced difficulty
	if len(solHash) > 0 && !Equal(pgHash, solHash) {
		pg, err := bc.GetProblemGraphFromHash(solHash)
		if err == nil {
			bestSol := bc.GetBestSolution(&pg, height)
			if (len(solution) > len(bestSol)) && pg.ValidateClique(solution) {
				return bc.CalculateTarget(height+1, true)
			}
		}
	}
	return bc.CalculateTarget(height+1, false)
}

func (bc *Blockchain) GetVerifiedTransactions(transactions []*Transaction) []*Transaction {
	for i := len(transactions) - 1; i >= 0; i-- {
	    tx := transactions[i]
	    if bc.VerifyTransaction(tx) != true {
			fmt.Println("ERROR: Invalid transaction\n", tx)
			transactions = append(transactions[:i],
	                transactions[i+1:]...)
		}
	}

	return transactions
}


// MineBlock mines a new block with the provided transactions
func (bc *Blockchain) MineBlock(transactions []*Transaction, solHash []byte, solution []int,  pgHash []byte) *Block {
	// bci := bc.Iterator()
	// lastBlock := bci.Next()
	// lastHash := lastBlock.Hash
	// lastHeight := lastBlock.Height
	var lastHash []byte
	var lastHeight int

	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l"))

		blockData := b.Get(lastHash)
		block := DeserializeBlock(blockData)

		lastHeight = block.Height

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	verifiedTxs := bc.GetVerifiedTransactions(transactions)
	//fmt.Println(verifiedTxs)
	target := bc.GetBlockTarget(lastHeight, solHash, solution, pgHash)	

	newBlock := NewBlock(verifiedTxs, lastHash, lastHeight+1, target, solHash, solution, pgHash)

	return newBlock
}


// SignTransaction signs inputs of a Transaction
func (bc *Blockchain) SignTransaction(tx *Transaction, privKey ecdsa.PrivateKey) {
	prevTXs := make(map[string]Transaction)

	for _, vin := range tx.Vin {
		prevTX, err := bc.FindTransaction(vin.Txid)
		if err != nil {
			log.Panic(err)
		}
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	tx.Sign(privKey, prevTXs)
}

// VerifyTransaction verifies transaction input signatures
func (bc *Blockchain) VerifyTransaction(tx *Transaction) bool {
	if tx.IsCoinbase() {
		return true
	}

	prevTXs := make(map[string]Transaction)

	for _, vin := range tx.Vin {
		prevTX, err := bc.FindTransaction(vin.Txid)
		if err != nil {
			log.Panic(err)
		}
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	return tx.Verify(prevTXs)
}

// AddProblemGraph add a problem to the database
func (bc *Blockchain) AddProblemGraph(pg *ProblemGraph) {
	err := bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(problemsBucket))
		problemInDb := b.Get(pg.Hash)
		if problemInDb != nil {
			return nil
		}

		pgData := pg.Serialize()
		err := b.Put(pg.Hash, pgData)
		if err != nil {
			log.Panic(err)
		}

		return nil
	})
	if err != nil {
		log.Panic(err)
	}
}

func dbExists(dbFile string) bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}

	return true
}
