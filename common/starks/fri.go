package starks

import (
	"encoding/binary"
	"log"

	"github.com/holiman/uint256"
	"github.com/junwookheo/bcsos/common/blockchain"
	"github.com/junwookheo/bcsos/common/galois"
)

type fri struct {
	GFP    *galois.GFP
	numIdx int
}

func NewFri() *fri {
	f := fri{GFP: galois.NewGFP(), numIdx: 40}

	return &f
}

func (f *fri) Merklize(values []*uint256.Int) [][]byte {
	L := len(values)
	ms := make([][]byte, L*2)
	for i := 0; i < L; i++ {
		// ms[L+i] = blockchain.CalHashSha256(values[i].Bytes())
		ms[L+i] = values[i].Bytes()
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

func (f *fri) verifyMerkleBranch(root []byte, index int, proof [][]byte) []byte {
	index += 1 << len(proof)
	v := proof[0]
	for i := 1; i < len(proof); i++ {
		if index%2 == 0 {
			v = blockchain.CalMerkleNodeHash(v, proof[i])
		} else {
			v = blockchain.CalMerkleNodeHash(proof[i], v)
		}
		index = index >> 1
	}

	if f.GFP.Cmp(f.GFP.IntFromBytes(v), f.GFP.IntFromBytes(root)) != 0 {
		log.Printf("Faile to verify : %v-%v", v, root)
		return nil
	}
	return proof[0]
}

func (f *fri) VerifyMultiBranch(root []byte, indices []uint32, proof [][][]byte) [][]byte {
	output := make([][]byte, len(proof))
	for i := 0; i < len(proof); i++ {
		output[i] = f.verifyMerkleBranch(root, int(indices[i]), proof[i])
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

func (f *fri) ProveLowDegree(values []*uint256.Int, rou *uint256.Int) []interface{} {
	L := len(values)
	if (L >> 2) <= 16 {
		log.Println("Produced FRI proof")
		ms := make([][]byte, L)
		for i := 0; i < L; i++ {
			ms[i] = values[i].Bytes() //   blockchain.CalHashSha256(values[i].Bytes())
			// ms[i] = t[:]
		}
		return []interface{}{ms}
	}

	log.Printf("Prove values with length : %v", L)
	size, xxs := f.GFP.ExtRootUnity(rou, false)
	if L != size-1 {
		log.Panicf("Mismatch the size of values and xs : %v, %v", L, len(xxs)-1)
		return nil
	}

	m1 := f.Merklize(values)
	special_x := f.GFP.IntFromBytes(m1[1])
	quarter_len := len(xxs) >> 2

	x_polys := make([][]*uint256.Int, quarter_len)
	dxs := make([][]*uint256.Int, quarter_len)
	dys := make([][]*uint256.Int, quarter_len)
	colums := make([]*uint256.Int, quarter_len)
	for i := 0; i < quarter_len; i++ {
		xs := []*uint256.Int{xxs[i], xxs[i+quarter_len], xxs[i+2*quarter_len], xxs[i+3*quarter_len]}
		ys := []*uint256.Int{values[i], values[i+quarter_len], values[i+2*quarter_len], values[i+3*quarter_len]}
		x_poly := f.GFP.LagrangeInterp(xs[:], ys[:])
		colums[i] = f.GFP.EvalPolyAt(x_poly, special_x)
		x_polys[i] = x_poly
		dxs[i] = xs
		dys[i] = ys
	}

	m2 := f.Merklize(colums)
	yys := f.GetPseudorandomIndices(m2[1], uint32(len(colums)), f.numIdx)

	poly_positions := make([]uint32, len(yys)*4)
	for i := 0; i < len(yys); i++ {
		for j := 0; j < 4; j++ {
			poly_positions[i*4+j] = yys[i] + uint32(quarter_len*j)
		}
	}

	current := make([]interface{}, 3)
	current[0] = m2[1]
	current[1] = f.MakeMultiBranch(m2, yys)
	current[2] = f.MakeMultiBranch(m1, poly_positions)

	next := f.ProveLowDegree(colums, f.GFP.Exp(rou, uint256.NewInt(4)))

	proof := make([]interface{}, 1)
	proof[0] = current
	proof = append(proof, next...)
	return proof
}

func (f *fri) VerifyLowDegreeProof(root []byte, proof []interface{}, rou *uint256.Int) bool {
	testval := rou.Clone()
	roudeg := uint64(1)

	for testval.Cmp(uint256.NewInt(1)) != 0 {
		roudeg *= 2
		testval = f.GFP.Mul(testval, testval)
	}

	qrou := make([]*uint256.Int, 4)
	for i := 0; i < 4; i++ {
		b := f.GFP.Mul(uint256.NewInt(uint64(i)), uint256.NewInt(roudeg>>2))
		qrou[i] = f.GFP.Exp(rou, b)
	}

	for i := 0; i < len(proof)-1; i++ {
		roudeg = roudeg / 4
		log.Printf("Verify values with length : %v", roudeg)
		p := proof[i].([]interface{})
		root2, _ := p[0].([]byte)
		cbranch, _ := p[1].([][][]byte)
		pbranch, _ := p[2].([][][]byte)

		ys := f.GetPseudorandomIndices(root2, uint32(roudeg), f.numIdx)
		poly_positions := make([]uint32, len(ys)*4)
		for j := 0; j < len(ys); j++ {
			for k := 0; k < 4; k++ {
				poly_positions[j*4+k] = ys[j] + uint32(int(roudeg)*k)
			}
		}

		column_values := f.VerifyMultiBranch(root2, ys, cbranch)
		poly_values := f.VerifyMultiBranch(root, poly_positions, pbranch)

		xs := make([][]*uint256.Int, len(ys))
		rows := make([][]*uint256.Int, len(ys))
		columnvals := make([]*uint256.Int, len(ys))
		for j := 0; j < len(ys); j++ {
			x1 := f.GFP.Exp(rou, uint256.NewInt(uint64(ys[j])))
			xs[j] = []*uint256.Int{f.GFP.Mul(qrou[0], x1), f.GFP.Mul(qrou[1], x1), f.GFP.Mul(qrou[2], x1), f.GFP.Mul(qrou[3], x1)}
			row := [][]byte{poly_values[j*4], poly_values[j*4+1], poly_values[j*4+2], poly_values[j*4+3]}
			rows[j] = []*uint256.Int{f.GFP.IntFromBytes(row[0]), f.GFP.IntFromBytes(row[1]), f.GFP.IntFromBytes(row[2]), f.GFP.IntFromBytes(row[3])}
			columnvals[j] = f.GFP.IntFromBytes(column_values[j])
		}

		special_x := f.GFP.IntFromBytes(root)
		for j := 0; j < len(ys); j++ {
			poly := f.GFP.LagrangeInterp(xs[j], rows[j])
			eval := f.GFP.EvalPolyAt(poly, special_x)
			if f.GFP.Cmp(eval, columnvals[j]) != 0 {
				log.Printf("Evaluation fail : %v-%v", eval, columnvals[j])
				return false
			}
		}

		root = root2
		rou = f.GFP.Exp(rou, uint256.NewInt(4))

	}

	log.Println("Evaluation success")
	return true
}
