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
	cands       []candblock
}

type SaveBlock interface {
	AddBlock(b *blockchain.Block) int64
}

func (q *CandidateBlocks) PushAndSave(block *blockchain.Block, sb SaveBlock) bool {
	dif := block.Header.Height - q.maxheight
	q.mutex.Lock()
	defer q.mutex.Unlock()

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
					log.Printf("Save block(%v) : %v", b.Header.Height, hex.EncodeToString(b.Header.Hash))
					sb.AddBlock(b)
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

func (q *CandidateBlocks) ShowAll() {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	for _, cand := range q.cands {
		for _, b := range cand.blocks {
			log.Printf("CandBlocks(%v) : %v - %v", b.Header.Height, hex.EncodeToString(b.Header.Hash), hex.EncodeToString(b.Header.PrvHash))
		}

	}

}

func NewCandidateBlocks() *CandidateBlocks {
	capacity := config.FINALITY * 2

	cbs := CandidateBlocks{
		capacity:    capacity,
		maxheight:   -1,
		savedheight: -1,
		cands:       make([]candblock, 0, capacity),
	}

	return &cbs
}
