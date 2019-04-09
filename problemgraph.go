package crickchain

import (
	"fmt"
	"log"
	"encoding/gob"
	"crypto/sha256"
	"bytes"
	"strconv"
	"encoding/json"
	"github.com/soniakeys/graph"
	"github.com/soniakeys/bits"
	//"github.com/boltdb/bolt"
)


type ProblemGraph struct {
	Hash 	[]byte
	Graph 	*graph.Undirected
}


func NewProblemGraph(nodes int, edges int) *ProblemGraph {
	// g := graph.Undirected{make(graph.AdjacencyList, 3)} // 3 nodes altogether
	// g.AddEdge(0, 1)
	// g.AddEdge(1, 2)
	// g.AddEdge(2, 0)
	g := graph.GnmUndirected(nodes, edges, nil)
	pg := ProblemGraph{[]byte{}, &g}
	pg.Hash = pg.GetHash()
	return &pg
}

// Hash the graph
func (pg *ProblemGraph) GetHash() []byte {
    arrBytes := []byte{}
    jsonBytes, _ := json.Marshal(pg)
    arrBytes = append(arrBytes, jsonBytes...)
    

	hash := sha256.Sum256(arrBytes)

	return hash[:]
}


//FindCliques finds all k-cliques and returns them
func (pg *ProblemGraph) FindAllKCliques(k int) [][]int {
	//we check that we have a siple (not loops nor parallels) graph
	simple, _ := pg.Graph.IsSimple()
	if !simple {
		return [][]int{}
	}

	kCliques := [][]int{}
	//pg.Graph.BKPivotMaxDegree, 
	pg.Graph.BronKerbosch1(func(c bits.Bits) bool {
		clique := c.Slice()
    	if len(clique) == k {
    		kCliques = append(kCliques, clique)
    	} else if len(clique) > k + 4 { //this +4 is arbitrary. The problem is that if the cliques is a sub-clique, is not found unless we return false here for the size of the parent clique
    		return false
    	}
    	return true
	})
	return kCliques
}

//FindClique finds one at least k-clique and returns it
func (pg *ProblemGraph) FindKClique(k int) []int {
	//we check that we have a siple (not loops nor parallels) graph
	simple, _ := pg.Graph.IsSimple()
	if !simple {
		return []int{}
	}

	kClique := []int{}
	//pg.Graph.BKPivotMaxDegree
	pg.Graph.BronKerbosch1(func(c bits.Bits) bool {
		clique := c.Slice()
    	if len(clique) >= k {
    		kClique = clique
    		return false
    	} else if len(clique) > k {
    		return false
    	}
    	return true
	})
	return kClique
}

//ValidateClique checks that the input is a clique of the graph
func (pg *ProblemGraph) ValidateClique(clique []int) bool {
	for _, n := range clique {
		for _, m := range clique {
    		if n != m {
    			if !pg.Graph.RemoveEdge(graph.NI(n), graph.NI(m)) {//return false if the edge is not present
    				return false
    			} else {//this is probably not important, since we dont care about this pg instance
    				pg.Graph.AddEdge(graph.NI(n), graph.NI(m))
    			}
    		}
		}	
	}
	return true
}

//FindClique finds all max-cliques and returns them. This scales exponentially (it's Np complete)
func (pg *ProblemGraph) FindMaxClique() [][]int {
	//we check that we have a siple (not loops nor parallels) graph
	simple, _ := pg.Graph.IsSimple()
	if !simple {
		return [][]int{}
	}

	var maxCliques [][]int
	m := 0
	pg.Graph.BronKerbosch3(pg.Graph.BKPivotMaxDegree, func(c bits.Bits) bool {
		clique := c.Slice()
    	if len(clique) > m {
    		maxCliques = maxCliques[:0]
    		maxCliques = append(maxCliques, clique)
    		m = len(clique)
    	} else if len(clique) == m {
    		maxCliques = append(maxCliques, clique)
    	}
    	return true
	})
	return maxCliques
}

//NicePrint print nicely the graph properties
func (pg *ProblemGraph) NicePrint(bc *Blockchain) {
	fmt.Printf("\n")
	printBlue(fmt.Sprintf("Hash: %x\n",pg.Hash))
	// for fr, to := range pg.Graph.AdjacencyList {
 //    	fmt.Println(fr, to)
	// }
	connected := pg.Graph.IsConnected()
	printGreen(fmt.Sprintf("Connected: %s\n", strconv.FormatBool(connected)))
	// for k := 3; k <= 8; k++ {
	// 	kcliques := pg.FindAllKCliques(k)
	// 	printYellow(fmt.Sprintf("%d %d-cliques:",len(kcliques), k))
	// 	for _, c := range kcliques {
	// 		fmt.Print(c)
	// 	}
	// 	fmt.Print("\n")
	// }
	bsol := bc.GetBestSolution(pg, bc.GetBestHeight())
	printYellow(fmt.Sprintf("Best solution: %d-clique:",len(bsol)))
	fmt.Println(bsol)
}

// Serialize serializes the graph
func (pg *ProblemGraph) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(pg)
	if err != nil {
		log.Panic(err)
	}

	return result.Bytes()
}

// DeserializeProblemGraph deserializes a problemgraph
func DeserializeProblemGraph(d []byte) *ProblemGraph {
	var g ProblemGraph

	decoder := gob.NewDecoder(bytes.NewReader(d))
	err := decoder.Decode(&g)
	if err != nil {
		log.Panic(err)
	}

	return &g
}

