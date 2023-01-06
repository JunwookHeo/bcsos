package starks

import (
	"encoding/binary"
	"log"

	"github.com/holiman/uint256"
	"github.com/junwookheo/bcsos/common/blockchain"
	"github.com/junwookheo/bcsos/common/galois"
)

type fri struct {
	GFP *galois.GFP
}

func NewFri() *fri {
	f := fri{GFP: galois.NewGFP()}

	return &f
}

func (f *fri) Merklize(values []*uint256.Int) [][]byte {
	L := len(values)
	ms := make([][]byte, L*2)
	for i := 0; i < L; i++ {
		ms[L+i] = blockchain.CalHashSha256(values[i].Bytes())
	}

	for i := L - 1; i > 0; i-- {
		ms[i] = blockchain.CalMerkleNodeHash(ms[i*2], ms[i*2+1])
	}
	return ms
}

func (f *fri) makeMerkleBranch(tree [][]byte, index int) [][]byte {
	output := make([][]byte, 0)

	index += len(tree) >> 1
	o := tree[index]
	output = append(output, o)
	for index > 1 {
		o := tree[index^0x1]
		output = append(output, o)
		index = index >> 1
	}
	return output
}

func (f *fri) MakeMultiBranch(tree [][]byte, indices []uint32) [][][]byte {
	output := make([][][]byte, len(indices))
	for i := 0; i < len(indices); i++ {
		branch := f.makeMerkleBranch(tree, int(indices[i]))
		output[i] = branch
	}
	return output
}

func (f *fri) GetPseudorandomIndices(seed []byte, modulus uint32, count int) []uint32 {
	data := make([]byte, 4*count)
	r := seed
	size := 0
	for size < 4*count {
		for i := 0; i < len(r); i++ {
			if size >= 4*count {
				break
			}

			data[size] = r[i]
			size++
			r = blockchain.CalHashSha256(data)
		}
	}
	indices := make([]uint32, count)
	for i := 0; i < count; i++ {
		indices[i] = binary.BigEndian.Uint32(data[i*4:i*4+4]) % modulus
	}

	return indices
}

func (f *fri) ProveLowDegree(values []*uint256.Int, rou *uint256.Int, maxdeg int) []interface{} {
	if maxdeg <= 16 {
		log.Println("Produced FRI proof")
		ms := make([]interface{}, len(values))
		for i := 0; i < len(values); i++ {
			ms[i] = values[i].Bytes() //   blockchain.CalHashSha256(values[i].Bytes())
		}
		return ms
	}

	size, xs := f.GFP.ExtRootUnity(rou, false)
	if len(values) != size-1 {
		log.Panicf("Mismatch the size of values and xs : %v, %v", len(values), len(xs)-1)
		return nil
	}

	m1 := f.Merklize(values)
	special_x := f.GFP.IntFromBytes(m1[1])
	quarter_len := len(xs) >> 2

	colums := make([]*uint256.Int, quarter_len)
	for i := 0; i < quarter_len; i++ {
		xs := [4]*uint256.Int{xs[i], xs[i+quarter_len], xs[i+2*quarter_len], xs[i+3*quarter_len]}
		ys := [4]*uint256.Int{values[i], values[i+quarter_len], values[i+2*quarter_len], values[i+3*quarter_len]}
		x_poly := f.GFP.LagrangeInterp(xs[:], ys[:])
		colums[i] = f.GFP.EvalPolyAt(x_poly, special_x)
	}

	m2 := f.Merklize(colums)
	ys := f.GetPseudorandomIndices(m2[1], uint32(len(colums)), 10)

	poly_positions := make([]uint32, len(ys)*4)
	for i := 0; i < len(ys); i++ {
		for j := 0; j < 4; j++ {
			poly_positions[i*4+j] = ys[i] + uint32(quarter_len*j)
		}
	}

	output := make([]interface{}, 3)
	output[0] = m2[1]
	output[1] = f.MakeMultiBranch(m2, ys)
	output[2] = f.MakeMultiBranch(m1, poly_positions)

	f.ProveLowDegree(colums, f.GFP.Exp(rou, uint256.NewInt(4)), maxdeg>>2)

	return output
}
