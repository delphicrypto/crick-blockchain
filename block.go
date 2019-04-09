package crickchain

import (
	"bytes"
	"encoding/gob"
	"log"
	"fmt"
	"strconv"
	"time"
	"math/big"
)



// Block represents a block in the blockchain
type Block struct {
	Timestamp     int64
	Transactions  []*Transaction
	PrevBlockHash []byte
	Hash          []byte
	Nonce         int
	Height        int
	Target	  	  *big.Int
	SolutionHash []byte
	Solution	  []int
	ProblemGraphHash []byte
}

// NewBlock creates and returns Block
func NewBlock(transactions []*Transaction, prevBlockHash []byte, height int, target *big.Int, solHash []byte, solution []int, pgHash []byte) *Block {
	block := &Block{time.Now().UnixNano(), transactions, prevBlockHash, []byte{}, 0, height, target, solHash, solution, pgHash}
	pow := NewProofOfWork(block)
	nonce, hash := pow.Run()

	block.Hash = hash[:]
	block.Nonce = nonce

	return block
}

// NewGenesisBlock creates and returns genesis Block
func NewGenesisBlock(coinbase *Transaction) *Block {
	target := targetFromTargetBits(initialTargetBits)
	
	return NewBlock([]*Transaction{coinbase}, []byte{}, 0, target, []byte{}, []int{}, []byte{})
}

// HashTransactions returns a hash of the transactions in the block
func (b *Block) HashTransactions() []byte {
	var transactions [][]byte

	for _, tx := range b.Transactions {
		transactions = append(transactions, tx.Serialize())
		//fmt.Println(tx, tx.Serialize(), string(tx.Serialize()), DeserializeTransaction(tx.Serialize()))
	}
	mTree := NewMerkleTree(transactions)

	return mTree.RootNode.Data
}

func (b *Block) HasValidSolution(bc *Blockchain) bool {
	if len(b.SolutionHash) == 0 {
		return false
	}
	//check that is not the initial solution posted with the problem
	if Equal(b.ProblemGraphHash, b.SolutionHash) {
		return false
	}
	pg, err := bc.GetProblemGraphFromHash(b.SolutionHash)
	if err != nil {
		return false
	}
	bestSolution := bc.GetBestSolution(&pg, b.Height - 1)
	if len(b.Solution) <= len(bestSolution) {
		return false
	}

	//verify that solution is valid
	return pg.ValidateClique(b.Solution)
}

// Serialize serializes the block
func (b *Block) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(b)
	if err != nil {
		log.Panic(err)
	}

	return result.Bytes()
}

func (b *Block) Validate(bc *Blockchain) bool {
	var chainTarget *big.Int
	if b.HasValidSolution(bc) {
		chainTarget = bc.CalculateTarget(b.Height, true)
	} else {
		chainTarget = bc.CalculateTarget(b.Height, false)
	}
	
	//check that the targetBits is correct
	if b.Target.Cmp(chainTarget) != 0 {
		return false
	}
	pow := NewProofOfWork(b)
	return pow.Validate()
}

//NicePrint print nicely the block properties
func (b *Block) NicePrint(bc *Blockchain) {
	fmt.Printf("\n")
	printGreen(fmt.Sprintf("============ Block %d ============\n", b.Height))
	printBlue(fmt.Sprintf("Hash:   %064x\n", b.Hash))
	fmt.Printf("Prev:   %064x\n", b.PrevBlockHash)
	fmt.Printf("Target: %064x\n", b.Target)
	fmt.Printf("Difficulty: %d\n", targetToDifficulty(b.Target))
	prevBlock, _ := bc.GetBlockFromHash(b.PrevBlockHash)
	time := (b.Timestamp - prevBlock.Timestamp) / 1e9
	fmt.Printf("Time: %d seconds\n", time)
	validBlock := b.Validate(bc)
	if validBlock {
		printGreen(fmt.Sprintf("PoW: %s\n", strconv.FormatBool(validBlock)))
	} else {
		printRed(fmt.Sprintf("PoW: %s\n", strconv.FormatBool(validBlock)))
	}

	if len(b.SolutionHash) > 0 {
		printGreen(fmt.Sprintf("Solution to %x: ", b.SolutionHash))
		fmt.Println(b.Solution)
		validSol := b.HasValidSolution(bc)
		if validSol {
			printGreen("Valid Solution\n")
		} else {
			printRed("Not Valid\n")
		}
	} else {
		printRed("No solution\n")
	}

	if len(b.ProblemGraphHash) > 0 {
		printGreen(fmt.Sprintf("New Problem %x \n", b.ProblemGraphHash))
		// pg, err := bc.GetProblemGraphFromHash(b.ProblemGraphHash)
		// if err == nil {
		// 	pg.NicePrint()
		// }
		
	} else {
		printRed("No problem ;)\n")
	}

	for _, tx := range b.Transactions {
		printYellow(fmt.Sprintln(tx))
	}
	fmt.Printf("\n")
}

// DeserializeBlock deserializes a block
func DeserializeBlock(d []byte) *Block {
	var block Block

	decoder := gob.NewDecoder(bytes.NewReader(d))
	err := decoder.Decode(&block)
	if err != nil {
		log.Panic(err)
	}

	return &block
}

