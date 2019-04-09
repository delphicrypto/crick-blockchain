# Blockchain for https://arxiv.org/abs/1708.09419

## Modules

```
go get github.com/boltdb/bolt
go get github.com/soniakeys/graph
go get github.com/soniakeys/bits
go get golang.org/x/crypto/ripemd160
go get github.com/fatih/color
```


## Launch

From the main foler:
```
export NODE_ID=3000
go run main.go
```
To quickly generate a wallet and a blockchain use the command `qs`.

## TODO
add a lookup table for the difficulties, because calculateTarget now tales a lot of time

rethink diff update?

maybe add check that graph has no better solution? (if no k+1-clique with current best k-clique is found)

add send problem graphs to server. Or maybe add [ipfs](https://github.com/ipfs/go-ipfs-api) implementation of problemgraphs 

## BUGS
weird bug in genesis: txs hash changes after reload of blockchain, so that genesis pow check fails. the problem is the tx.serialize() that changes after closing the program and reopening it
the mining go routines in minepar don't stop if the block has already been mined in the other way. need to rewrite mining completely to implement stop signal in clique finder and nonce finder (terrible).