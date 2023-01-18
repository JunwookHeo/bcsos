package starks

import (
	"encoding/binary"
	"log"

	"github.com/holiman/uint256"
	"github.com/junwookheo/bcsos/common/blockchain"
	"github.com/junwookheo/bcsos/common/galois"
)

type starks struct {
	GFP       *galois.GFP
	numIdx    int
	steps     int
	extFactor int
}

const PSIZE = 31

func NewStarks() *starks {
	f := starks{GFP: galois.NewGFP(), numIdx: 40, steps: 8192, extFactor: 8}

	return &f
}

func (f *starks) GetHashBytes(b []byte) []byte {
	return blockchain.CalHashSha256(b)
}

func (f *starks) Merklize(values []*uint256.Int) [][]byte {
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

func (f *starks) makeMerkleBranch(tree [][]byte, index int) [][]byte {
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

func (f *starks) MakeMultiBranch(tree [][]byte, indices []uint32) [][][]byte {
	output := make([][][]byte, len(indices))
	for i := 0; i < len(indices); i++ {
		branch := f.makeMerkleBranch(tree, int(indices[i]))
		output[i] = branch
	}
	return output
}

func (f *starks) verifyMerkleBranch(root []byte, index int, proof [][]byte) []byte {
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

func (f *starks) VerifyMultiBranch(root []byte, indices []uint32, proof [][][]byte) [][]byte {
	output := make([][]byte, len(proof))
	for i := 0; i < len(proof); i++ {
		output[i] = f.verifyMerkleBranch(root, int(indices[i]), proof[i])
	}
	return output
}

func (f *starks) GetPseudorandomIndices(seed []byte, modulus uint32, count int) []uint32 {
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

func (f *starks) ProveLowDegree(values []*uint256.Int, rou *uint256.Int) []interface{} {
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

func (f *starks) VerifyLowDegreeProof(root []byte, proof []interface{}, rou *uint256.Int) bool {
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

// Vis : Input values
// Vos : Output values
// key : Adress of Prover
func (f *starks) GenerateStarksProof(vis []byte, vos []byte, key []byte) []interface{} {
	visu := f.GFP.LoadUint256FromStream31(vis)
	vosu := f.GFP.LoadUint256FromStream32(vos)
	ks := f.GFP.LoadUint256FromKey(key)
	log.Printf("%v, %v, %v", len(ks), len(visu), len(vosu))

	precision := f.steps * f.extFactor
	skips := precision / f.steps
	log.Printf("precision : %v, skips : %v", precision, skips)

	g := f.GFP.Prime.Clone()
	g = g.Sub(g, uint256.NewInt(1))
	g = f.GFP.Div(g, uint256.NewInt(uint64(precision)))
	G2 := f.GFP.Exp(uint256.NewInt(7), g)
	log.Printf("G2 : %v", G2)
	G1 := f.GFP.Exp(G2, uint256.NewInt(uint64(skips)))
	log.Printf("G1 : %v", G1)

	// Vi(x)
	poly_visu := f.GFP.IDFT(visu[0:f.steps], G1)
	eval_visu := f.GFP.DFT(poly_visu, G2)
	log.Printf("visu : %v, %v", len(poly_visu), len(eval_visu))
	// for i := 0; i < 10; i++ {
	// 	log.Printf("eval_visu : %v, %v", visu[i], eval_visu[i*f.extFactor])
	// }

	// Vo(x)
	poly_vosu := f.GFP.IDFT(vosu[0:f.steps], G1)
	eval_vosu := f.GFP.DFT(poly_vosu, G2)
	log.Printf("vosu : %v, %v", len(poly_vosu), len(eval_vosu))
	// for i := 0; i < 10; i++ {
	// 	log.Printf("eval_vosu : %v, %v", vosu[i], eval_vosu[i*f.extFactor])
	// }

	skip2 := f.steps / len(ks)
	log.Printf("skip for key : %v", skip2)

	// K(x)
	poly_key := f.GFP.IDFT(ks, f.GFP.Exp(G1, uint256.NewInt(uint64(skip2))))
	eval_key := f.GFP.DFT(poly_key, f.GFP.Exp(G2, uint256.NewInt(uint64(skip2))))
	log.Printf("skip for key : %v", len(eval_key))

	// C(x) = C(Vi(x), Vo(x), K(x))
	// Vo[i]^3 * Vo[i-1] - K[i] = Vi[i]
	// C(x) = Vi(x) - (Vo(x)^3 * Vo(x/g1) - K(x))
	eval_cp := make([]*uint256.Int, precision)
	pre := uint256.NewInt(1)
	for i := 0; i < precision; i++ {
		c := f.GFP.Exp(eval_vosu[i], uint256.NewInt(3))
		c = f.GFP.Mul(c, pre)
		c = f.GFP.Sub(c, eval_key[i%len(eval_key)])
		eval_cp[i] = f.GFP.Sub(eval_visu[i], c)
		if i+1-f.extFactor < 0 {
			pre = eval_vosu[(i+1-f.extFactor)+precision]
		} else {
			pre = eval_vosu[i+1-f.extFactor]
		}
		if pre.IsZero() {
			pre = uint256.NewInt(1)
		}
	}

	size, xs := f.GFP.ExtRootUnity(G2, false) // return from x^0 to x^n, x^n == x^0
	size -= 1
	xs = xs[0:size]
	log.Printf("size : %v, xs[0] : %v, xs[-1] : %v", size, xs[0], xs[len(xs)-1])

	// Z(x)
	eval_inv_z := make([]*uint256.Int, precision)
	for i := 0; i < skips; i++ {
		eval_inv_z[i] = f.GFP.Inv(f.GFP.Sub(xs[(i*f.steps)%precision], uint256.NewInt(1)))
	}
	for i := skips; i < precision; i++ {
		eval_inv_z[i] = eval_inv_z[i%skips]
	}

	// D(x)
	eval_d := make([]*uint256.Int, precision)
	for i := 0; i < precision; i++ {
		eval_d[i] = f.GFP.Mul(eval_cp[i], eval_inv_z[i])
	}

	// Compute polynomial for Vos
	// Compute interpolant of ((1, input), (x_atlast_step, output))
	xos := []*uint256.Int{xs[0], xs[(f.steps-1)*f.extFactor]}
	yos := []*uint256.Int{vosu[0], vosu[f.steps-1]}

	// Io(x)
	poly_io := f.GFP.LagrangeInterp(xos, yos)
	eval_io := make([]*uint256.Int, precision)
	for i := 0; i < precision; i++ {
		eval_io[i] = f.GFP.EvalPolyAt(poly_io, xs[i])
	}

	f1 := []*uint256.Int{f.GFP.Sub(f.GFP.Prime, uint256.NewInt(1)), uint256.NewInt(1)}           // (x - 1)
	f2 := []*uint256.Int{f.GFP.Sub(f.GFP.Prime, xs[(f.steps-1)*f.extFactor]), uint256.NewInt(1)} // (x - lastpoint)

	// Zo(x)
	poly_zo := f.GFP.MulPolys(f1, f2)
	eval_inv_zo := make([]*uint256.Int, precision)
	for i := 0; i < precision; i++ {
		eval_inv_zo[i] = f.GFP.Inv(f.GFP.EvalPolyAt(poly_zo, xs[i]))
	}

	// B(x) = (Vo(x) - Io(x)) / Zo(x)
	eval_b := make([]*uint256.Int, precision)
	for i := 0; i < precision; i++ {
		eval_b[i] = f.GFP.Sub(eval_vosu[i], eval_io[i])
		eval_b[i] = f.GFP.Mul(eval_b[i], eval_inv_zo[i])
	}

	// Compute Merkle Root
	tree_vosu := f.Merklize(eval_vosu)
	tree_d := f.Merklize(eval_d)
	tree_b := f.Merklize(eval_b)
	mr := tree_vosu[1]
	mr = append(mr, tree_d[1]...)
	mr = append(mr, tree_b[1]...)
	m_root := f.GetHashBytes(mr)

	// L(x) : Linear combination
	k1 := uint256.NewInt(0)
	k1.SetBytes8(m_root[0:])
	k2 := uint256.NewInt(0)
	k2.SetBytes8(m_root[8:])
	k3 := uint256.NewInt(0)
	k3.SetBytes8(m_root[16:])
	k4 := uint256.NewInt(0)
	k4.SetBytes8(m_root[24:])

	G2_steps := f.GFP.Exp(G2, uint256.NewInt(uint64(f.steps)))
	powers := make([]*uint256.Int, f.extFactor)
	powers[0] = uint256.NewInt(1)
	for i := 1; i < len(powers); i++ {
		powers[i] = f.GFP.Mul(powers[i-1], G2_steps)
	}

	eval_l := make([]*uint256.Int, precision)
	for i := 0; i < precision; i++ {
		p := f.GFP.Add(f.GFP.Mul(eval_vosu[i], k1), f.GFP.Mul(eval_vosu[i], f.GFP.Mul(k2, powers[i%f.extFactor])))
		b := f.GFP.Add(f.GFP.Mul(eval_b[i], k3), f.GFP.Mul(eval_b[i], f.GFP.Mul(k4, powers[i%f.extFactor])))
		eval_l[i] = f.GFP.Add(eval_d[i], p)
		eval_l[i] = f.GFP.Add(eval_l[i], b)
	}

	// Compute Merkle root of L(x)
	tree_l := f.Merklize(eval_l)

	log.Printf("mtree : %v", m_root)
	positions := f.GetPseudorandomIndices(m_root, uint32(precision), 80)
	augmented_positions := make([]uint32, len(positions)*4)
	for i := 0; i < len(positions); i++ {
		for j := 0; j < 4; j++ {
			augmented_positions[i*4+j] = positions[i] + uint32(skips*j)
		}
	}
	// log.Printf("positions : %v", positions)
	// log.Printf("augmented_positions : %v", augmented_positions)
	// log.Printf("tree_l : %v", tree_l[1])

	proof := make([]interface{}, 7)
	proof[0] = m_root
	proof[1] = tree_l[1]
	proof[2] = f.MakeMultiBranch(tree_vosu, augmented_positions)
	proof[3] = f.MakeMultiBranch(tree_d, augmented_positions)
	proof[4] = f.MakeMultiBranch(tree_b, augmented_positions)
	proof[5] = f.MakeMultiBranch(tree_l, positions)
	proof[6] = f.ProveLowDegree(eval_l, G2)

	return proof
}