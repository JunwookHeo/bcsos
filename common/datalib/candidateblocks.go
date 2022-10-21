package datalib

import (
	"encoding/hex"
	"log"
	"sync"

	"github.com/junwookheo/bcsos/common/blockchain"
	"github.com/junwookheo/bcsos/common/config"
)

type candblock struct {
	height int
	blocks []*blockchain.Block
}

type CandidateBlocks struct {
	mutex       sync.Mutex
	capacity    int
	maxheight   int
	savedheight int
	highest     *blockchain.Block
	cands       []candblock
}

type SaveBlock interface {
	AddNewBlock(b interface{}) int64
}

func (q *CandidateBlocks) PushAndSave(block *blockchain.Block, sb SaveBlock) bool {
	dif := block.Header.Height - q.maxheight
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if q.highest == nil {
		q.highest = block
	} else if q.highest.Header.Height < block.Header.Height {
		q.highest = block
	}
	// log.Printf("Highest(%v) : %v", q.highest.Header.Height, hex.EncodeToString(q.highest.Header.Hash))

	// Append new block to the end of list.
	for i := 0; i < dif; i++ {
		if len(q.cands) == q.capacity {
			q.cands = q.cands[1:len(q.cands)]
		}

		q.maxheight++
		cand := candblock{q.maxheight + i, nil}
		q.cands = append(q.cands, cand)
	}

	for i := len(q.cands); 0 < i; i-- {
		if q.cands[i-1].height == block.Header.Height {
			q.cands[i-1].blocks = append(q.cands[i-1].blocks, block)
			break
		}
	}

	prehash := hex.EncodeToString(block.Header.Hash)
	savedheight := q.savedheight
	for i := len(q.cands); 0 < i; i-- {
		for _, b := range q.cands[i-1].blocks {
			if hex.EncodeToString(b.Header.Hash) == prehash {
				prehash = hex.EncodeToString(b.Header.PrvHash)
				if q.savedheight < b.Header.Height && b.Header.Height <= q.maxheight-config.FINALITY {
					// log.Printf("Save block(%v) : %v", b.Header.Height, hex.EncodeToString(b.Header.Hash))
					sb.AddNewBlock(b)
					if savedheight < b.Header.Height {
						savedheight = b.Header.Height
					}
				}
				break
			}
		}
	}

	q.savedheight = savedheight

	return true
}

func (q *CandidateBlocks) GetHighestBlockHash() (int, string) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if q.highest == nil {
		return -1, ""
	}

	return q.highest.Header.Height, hex.EncodeToString(q.highest.Header.Hash)
}

func (q *CandidateBlocks) ShowAll() {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	for _, cand := range q.cands {
		for _, b := range cand.blocks {
			log.Printf("CandBlocks(%v) : %v - %v", b.Header.Height, hex.EncodeToString(b.Header.Hash), hex.EncodeToString(b.Header.PrvHash))
		}

	}

}

// This is to check whether Finality is long enough.
// If finality is short, panic will happen.
func (q *CandidateBlocks) CheckFinality() {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if len(q.cands) <= config.FINALITY {
		return
	}

	cand := q.cands[len(q.cands)-1]
	highests := cand.blocks
	saved := ""

	if len(highests) == 1 {
		return
	}

	for _, hblock := range highests {
		// log.Printf("highest hash(%v) : %v", hblock.Header.Height, hex.EncodeToString(hblock.Header.Hash))
		hash := hex.EncodeToString(hblock.Header.PrvHash)
		height := hblock.Header.Height
		for i := len(q.cands) - 1; len(q.cands)-config.FINALITY < i; i-- {
			for _, b := range q.cands[i-1].blocks {
				if hex.EncodeToString(b.Header.Hash) == hash {
					hash = hex.EncodeToString(b.Header.PrvHash)
					height = b.Header.Height
					break
				}
			}
		}
		// log.Printf("hash(%v) : %v", height-1, hash)
		if saved != "" && saved != hash {
			log.Panicf("Finality Error(%v-%v) : %v %v", hblock.Header.Height, height, saved, hash)
		}
		saved = hash
	}

}

func NewCandidateBlocks() *CandidateBlocks {
	capacity := config.FINALITY * 2

	cbs := CandidateBlocks{
		capacity:    capacity,
		maxheight:   -1,
		savedheight: -1,
		highest:     nil,
		cands:       make([]candblock, 0, capacity),
	}

	return &cbs
}
