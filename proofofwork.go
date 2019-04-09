package crickchain

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"strconv"
	"math"
	"math/big"
)

var (
	maxNonce = math.MaxInt64
)

// ProofOfWork represents a proof-of-work
type ProofOfWork struct {
	block  *Block
	target *big.Int
}

// NewProofOfWork builds and returns a ProofOfWork
func NewProofOfWork(b *Block) *ProofOfWork {
	pow := &ProofOfWork{b, b.Target}

	return pow
}

func (pow *ProofOfWork) prepareData(nonce int) []byte {
	hashedTxs := []byte{}
	if len(pow.block.Transactions) > 0 {
		hashedTxs = pow.block.HashTransactions()
	}
	solutionString := ""
	if len(pow.block.Solution) > 0 {
		for _, i := range pow.block.Solution {
    		solutionString += strconv.Itoa(i)
		}
	}
	byteSolution := []byte(solutionString)
	data := bytes.Join(
		[][]byte{
			pow.block.PrevBlockHash,
			hashedTxs,
			IntToHex(pow.block.Timestamp),
			[]byte(fmt.Sprintf("%x",pow.target)),
			IntToHex(int64(nonce)),
			pow.block.SolutionHash,
			byteSolution,
			pow.block.ProblemGraphHash,
		},
		[]byte{},
	)

	return data
}

// Run performs a proof-of-work
func (pow *ProofOfWork) Run() (int, []byte) {
	var hashInt big.Int
	var hash [32]byte
	nonce := 0

	for nonce < maxNonce {
		data := pow.prepareData(nonce)

		hash = sha256.Sum256(data)
		hashInt.SetBytes(hash[:])

		if hashInt.Cmp(pow.target) == -1 {
			break
		} else {
			nonce++
		}
	}
	fmt.Print("\n\n")

	return nonce, hash[:]
}

// Validate validates block's PoW
func (pow *ProofOfWork) Validate() bool {
	
	var hashInt big.Int

	data := pow.prepareData(pow.block.Nonce)
	hash := sha256.Sum256(data)
	hashInt.SetBytes(hash[:])
	isValid := hashInt.Cmp(pow.target) == -1
	return isValid
}
