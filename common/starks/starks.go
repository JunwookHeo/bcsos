package starks

import (
	"encoding/binary"
	"encoding/hex"
	"log"
	"time"

	"github.com/holiman/uint256"
	"github.com/junwookheo/bcsos/common/blockchain"
	"github.com/junwookheo/bcsos/common/dtype"
	"github.com/junwookheo/bcsos/common/galois"
	"github.com/junwookheo/bcsos/common/poscipher"
)

type starks struct {
	GFP       *galois.GFP
	numIdx    int
	steps     int
	extFactor int
}

const PSIZE = 31
const MAXDEG = 32

func NewStarks(steps int) *starks {
	if steps == 0 {
		steps = 2048 // 8192 / 4
	}
	f := starks{GFP: galois.NewGFP(), numIdx: 40, steps: steps, extFactor: 8 * 2}

	return &f
}

func (f *starks) GetSteps() int {
	return f.steps
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

func (f *starks) makeMerkleBranch(tree [][]byte, tree_cache map[int][]byte, index int) [][]byte {
	output := make([][]byte, 0)

	addCache := func(idx int, d []byte) {
		_, ok := tree_cache[idx]
		if ok {
			output = append(output, nil)
		} else {
			tree_cache[idx] = d
			output = append(output, d)
		}
	}

	index += len(tree) >> 1
	o := tree[index]
	addCache(index, o)

	for index > 1 {
		o = tree[index^0x1]
		addCache(index^0x1, o)
		index = index >> 1
	}
	return output
}

func (f *starks) MakeMultiBranch(tree [][]byte, indices []uint32) [][][]byte {
	tree_cache := make(map[int][]byte)
	output := make([][][]byte, len(indices))

	for i := 0; i < len(indices); i++ {
		branch := f.makeMerkleBranch(tree, tree_cache, int(indices[i]))
		output[i] = branch
	}
	return output
}

func (f *starks) verifyMerkleBranch(root []byte, tree_cache map[int][]byte, index int, proof [][]byte) []byte {
	cacheCache := func(idx int, d []byte) []byte {
		val, ok := tree_cache[idx]
		if ok {
			return val
		} else {
			if d == nil {
				log.Panicf("Failed %v", idx)
			}
			tree_cache[idx] = d
		}

		return d
	}

	index += 1 << len(proof)
	// v := proof[0]
	p0 := cacheCache(index, proof[0])
	v := p0

	for i := 1; i < len(proof); i++ {
		// val := proof[i]
		val := cacheCache(index^0x1, proof[i])

		if index%2 == 0 {
			v = blockchain.CalMerkleNodeHash(v, val)
		} else {
			v = blockchain.CalMerkleNodeHash(val, v)
		}
		index = index >> 1
	}

	if f.GFP.Cmp(f.GFP.IntFromBytes(v), f.GFP.IntFromBytes(root)) != 0 {
		log.Printf("Faile to verify : %v-%v", v, root)
		return nil
	}
	return p0
}

func (f *starks) VerifyMultiBranch(root []byte, indices []uint32, proof [][][]byte) [][]byte {
	tree_cache := make(map[int][]byte)
	output := make([][]byte, len(proof))
	for i := 0; i < len(proof); i++ {
		output[i] = f.verifyMerkleBranch(root, tree_cache, int(indices[i]), proof[i])
		if output[i] == nil {
			log.Printf("Faile to verify : index %v", int(indices[i]))
			return nil
		}
	}
	return output
}

func (f *starks) GetPseudorandomIndices(seed []byte, modulus uint32, count int, exclude uint32) []uint32 {
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
	real_modulus := modulus
	if exclude > 1 {
		real_modulus = modulus * (exclude - 1) / exclude
	}
	for i := 0; i < count; i++ {
		indices[i] = binary.BigEndian.Uint32(data[i*4:i*4+4]) % real_modulus
		indices[i] = indices[i] + 1 + indices[i]/(exclude-1)
	}

	// for i := 0; i < count; i++ {
	// 	if indices[i]%exclude == 0 {
	// 		log.Panicf("Wrong number : %v", indices[i])
	// 	}

	// }
	return indices
}

func (f *starks) ProveLowDegree(values []*uint256.Int, rou *uint256.Int) []*dtype.FriProofElement {
	L := len(values)
	log.Printf("Prove values with length : %v", L)

	if (L >> 2) <= MAXDEG {
		log.Println("Produced FRI proof")
		ms := make([][]byte, L)
		for i := 0; i < L; i++ {
			ms[i] = values[i].Bytes() //   blockchain.CalHashSha256(values[i].Bytes())
			// ms[i] = t[:]
		}
		proof := make([]*dtype.FriProofElement, 1)
		var current dtype.FriProofElement
		current.Root2 = ms
		proof[0] = &current
		return proof
	}

	size, xxs := f.GFP.ExtRootUnity(rou, false)
	if L != size-1 {
		log.Panicf("Mismatch the size of values and xs : %v, %v", L, len(xxs)-1)
		return nil
	}

	m1 := f.Merklize(values)
	special_x := f.GFP.IntFromBytes(m1[1])
	quarter_len := len(xxs) >> 2

	// start := time.Now().UnixNano()
	// log.Printf("ProveLowDegree 1: %v", time.Now().UnixNano()-start)

	// x_polys := make([][]*uint256.Int, quarter_len)
	// dxs := make([][]*uint256.Int, quarter_len)
	// dys := make([][]*uint256.Int, quarter_len)
	colums := make([]*uint256.Int, quarter_len)
	for i := 0; i < quarter_len; i++ {
		xs := []*uint256.Int{xxs[i], xxs[i+quarter_len], xxs[i+2*quarter_len], xxs[i+3*quarter_len]}
		ys := []*uint256.Int{values[i], values[i+quarter_len], values[i+2*quarter_len], values[i+3*quarter_len]}

		x_poly := f.GFP.LagrangeInterp_4(xs[:], ys[:])
		colums[i] = f.GFP.EvalPolyAt(x_poly, special_x)
		// x_polys[i] = x_poly
		// dxs[i] = xs
		// dys[i] = ys
	}
	// log.Printf("ProveLowDegree 2: %v", time.Now().UnixNano()-start)

	m2 := f.Merklize(colums)
	yys := f.GetPseudorandomIndices(m2[1], uint32(len(colums)), f.numIdx, uint32(f.extFactor))

	poly_positions := make([]uint32, len(yys)*4)
	for i := 0; i < len(yys); i++ {
		for j := 0; j < 4; j++ {
			poly_positions[i*4+j] = yys[i] + uint32(quarter_len*j)
		}
	}

	proof := make([]*dtype.FriProofElement, 1)
	var current dtype.FriProofElement
	current.Root2 = make([][]byte, 1)
	current.Root2[0] = m2[1]
	current.CBranch = f.MakeMultiBranch(m2, yys)
	current.PBranch = f.MakeMultiBranch(m1, poly_positions)
	next := f.ProveLowDegree(colums, f.GFP.Exp(rou, uint256.NewInt(4)))
	proof[0] = &current
	proof = append(proof, next...)

	return proof
}

func (f *starks) VerifyLowDegreeProof(root []byte, proof []*dtype.FriProofElement, rou *uint256.Int) bool {
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
		// log.Printf("Verify values with length : %v", roudeg)
		// p := proof[i].([]interface{})
		root2 := proof[i].Root2[0]
		cbranch := proof[i].CBranch
		pbranch := proof[i].PBranch

		ys := f.GetPseudorandomIndices(root2, uint32(roudeg), f.numIdx, uint32(f.extFactor))
		poly_positions := make([]uint32, len(ys)*4)
		for j := 0; j < len(ys); j++ {
			for k := 0; k < 4; k++ {
				poly_positions[j*4+k] = ys[j] + uint32(int(roudeg)*k)
			}
		}

		column_values := f.VerifyMultiBranch(root2, ys, cbranch)
		poly_values := f.VerifyMultiBranch(root, poly_positions, pbranch)
		if column_values == nil || poly_values == nil {
			log.Printf("Evaluation fail : %v-%v", column_values, poly_values)
			return false
		}

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
			poly := f.GFP.LagrangeInterp_4(xs[j], rows[j])
			eval := f.GFP.EvalPolyAt(poly, special_x)
			if f.GFP.Cmp(eval, columnvals[j]) != 0 {
				log.Printf("Evaluation fail : %v-%v", eval, columnvals[j])
				return false
			}
		}

		root = root2
		rou = f.GFP.Exp(rou, uint256.NewInt(4))

	}

	// roudeg = roudeg / 2

	data := proof[len(proof)-1]
	mdata := make([]*uint256.Int, len(data.Root2))
	for i := 0; i < len(data.Root2); i++ {
		mdata[i] = f.GFP.IntFromBytes(data.Root2[i])
	}

	roudeg = uint64(len(mdata)/f.extFactor) * 8
	log.Printf("rou deg-data deg : %v-%v", roudeg, len(mdata))

	mtree := f.Merklize(mdata)
	if f.GFP.Cmp(f.GFP.IntFromBytes(mtree[1]), f.GFP.IntFromBytes(root)) != 0 {
		log.Printf("Evaluation fail : %v-%v", mtree[1], root)
		return false
	}

	pts := make([]int, 0)
	for i := 0; i < len(mdata); i++ {
		if i%f.extFactor != 0 {
			pts = append(pts, i)
		}
	}

	_, powers := f.GFP.ExtRootUnity(rou, false)

	txs := make([]*uint256.Int, 0)
	for _, v := range pts[:roudeg] {
		txs = append(txs, powers[v])
	}

	tys := make([]*uint256.Int, 0)
	for _, v := range pts[:roudeg] {
		tys = append(tys, mdata[v])
	}

	poly_t := f.GFP.LagrangeInterp(txs, tys)
	for _, v := range pts[roudeg:] {
		e := f.GFP.EvalPolyAt(poly_t, powers[v])
		// log.Printf("%v, %v", v, e)
		if e.Cmp(mdata[v]) != 0 {
			log.Printf("Evaluation failed %v, %v", e, mdata[v])
			return false
		}
	}

	log.Println("Evaluation success")
	return true
}

// Vis : Input values
// Vos : Output values
// key : Adress of Prover
func (f *starks) GenerateStarksProof(vis []byte, vos []byte, key []byte) *dtype.StarksProof {
	pxu := f.GFP.LoadUint256FromStream31(vis)
	oxu := f.GFP.LoadUint256FromStream32(vos)
	kxu := f.GFP.LoadUint256FromStream32(key)

	log.Printf("Size of Inputs ks(%v), vi(%v), vo(%v)", len(kxu), len(pxu), len(oxu))

	klen := len(kxu)
	if len(kxu) < len(pxu) {
		for i := klen; i < len(pxu); i++ {
			kxu = append(kxu, kxu[i%klen])
		}
	} else {
		kxu = kxu[:len(pxu)]
	}

	for i := len(pxu); i < f.steps; i++ {
		pxu = append(pxu, uint256.NewInt(0))
	}
	for i := len(oxu); i < f.steps; i++ {
		oxu = append(oxu, uint256.NewInt(0))
	}
	for i := len(kxu); i < f.steps; i++ {
		kxu = append(kxu, uint256.NewInt(0))
	}

	precision := f.steps * f.extFactor
	skips := precision / f.steps
	log.Printf("precision : %v, skips : %v", precision, skips)

	g := f.GFP.Prime.Clone()
	g = g.Sub(g, uint256.NewInt(1))
	g = f.GFP.Div(g, uint256.NewInt(uint64(precision)))
	G2 := f.GFP.Exp(uint256.NewInt(7), g)
	// log.Printf("G2 : %v", G2)
	G1 := f.GFP.Exp(G2, uint256.NewInt(uint64(skips)))
	// log.Printf("G1 : %v", G1)

	// P(x)
	poly_pxu := f.GFP.IDFT(pxu[0:f.steps], G1)
	eval_pxu := f.GFP.DFT(poly_pxu, G2)
	// log.Printf("visu : %v, %v", len(poly_visu), len(eval_visu))
	// for i := 0; i < 10; i++ {
	// 	log.Printf("eval_visu : %v, %v", visu[i], eval_visu[i*f.extFactor])
	// }

	// O(x)
	poly_oxu := f.GFP.IDFT(oxu[0:f.steps], G1)
	eval_oxu := f.GFP.DFT(poly_oxu, G2)
	// log.Printf("vosu : %v, %v", len(poly_vosu), len(eval_vosu))
	// for i := 0; i < 10; i++ {
	// 	log.Printf("eval_vosu : %v, %v", vosu[i], eval_vosu[i*f.extFactor])
	// }

	// O(x/g1)
	poxu := make([]*uint256.Int, f.steps)
	poxu[0] = uint256.NewInt(0)
	for i := 1; i < f.steps; i++ {
		poxu[i] = oxu[i-1]
	}
	poly_poxu := f.GFP.IDFT(poxu, G1)
	eval_poxu := f.GFP.DFT(poly_poxu, G2)

	// skip2 := f.steps / len(kxu)
	// log.Printf("skip for key : %v", skip2)

	// K(x)
	poly_kxu := f.GFP.IDFT(kxu[:f.steps], G1)
	eval_kxu := f.GFP.DFT(poly_kxu, G2)
	// log.Printf("skip for key : %v", len(eval_key))

	// C(x) = C(P(x), O(x), K(x))
	// C(x) = P(x) - (((O(x)^3 - K(x))^3 - O(x/g1))
	eval_cp := make([]*uint256.Int, precision)
	for i := 0; i < precision; i++ {
		c := eval_oxu[i].Clone()
		c = f.GFP.Exp(c, uint256.NewInt(3))
		c = f.GFP.Add(c, eval_kxu[i])
		c = f.GFP.Exp(c, uint256.NewInt(3))
		c = f.GFP.Add(c, eval_poxu[i])

		eval_cp[i] = f.GFP.Sub(eval_pxu[i], c)
	}

	size, xs := f.GFP.ExtRootUnity(G2, false) // return from x^0 to x^n, x^n == x^0
	xs = xs[:size-1]
	// log.Printf("size : %v, xs[0] : %v, xs[-1] : %v", size, xs[0], xs[len(xs)-1])

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
	yos := []*uint256.Int{oxu[0], oxu[f.steps-1]}

	// Io(x)
	poly_io := f.GFP.LagrangeInterp(xos, yos)
	eval_io := make([]*uint256.Int, precision)
	for i := 0; i < precision; i++ {
		eval_io[i] = f.GFP.EvalPolyAt(poly_io, xs[i])
	}

	f1 := []*uint256.Int{f.GFP.Sub(f.GFP.Prime, uint256.NewInt(1)), uint256.NewInt(1)}           // (x - 1)
	f2 := []*uint256.Int{f.GFP.Sub(f.GFP.Prime, xs[(f.steps-1)*f.extFactor]), uint256.NewInt(1)} // (x - lastpoint)

	start := time.Now().UnixNano()
	log.Printf("GenerateStarksProof 1: %v", time.Now().UnixNano()-start)
	// Zo(x)
	poly_zo := f.GFP.MulPolys(f1, f2)
	// eval_inv_zo := make([]*uint256.Int, precision)
	// for i := 0; i < precision; i++ {
	// 	eval_inv_zo[i] = f.GFP.Inv(f.GFP.EvalPolyAt(poly_zo, xs[i]))
	// }
	eval_zo := make([]*uint256.Int, precision)
	for i := 0; i < precision; i++ {
		eval_zo[i] = f.GFP.EvalPolyAt(poly_zo, xs[i])
	}
	eval_inv_zo := f.GFP.MultInv(eval_zo)
	log.Printf("GenerateStarksProof 2: %v", time.Now().UnixNano()-start)

	// B(x) = (Vo(x) - Io(x)) / Zo(x)
	eval_b := make([]*uint256.Int, precision)
	for i := 0; i < precision; i++ {
		eval_b[i] = f.GFP.Sub(eval_oxu[i], eval_io[i])
		eval_b[i] = f.GFP.Mul(eval_b[i], eval_inv_zo[i])
	}

	// Compute Merkle Root
	tree_oxu := f.Merklize(eval_oxu)
	tree_poxu := f.Merklize(eval_poxu)
	tree_kxu := f.Merklize(eval_kxu)
	tree_d := f.Merklize(eval_d)
	tree_b := f.Merklize(eval_b)
	mr := tree_oxu[1]
	mr = append(mr, tree_poxu[1]...)
	mr = append(mr, tree_kxu[1]...)
	mr = append(mr, tree_d[1]...)
	mr = append(mr, tree_b[1]...)
	m_root := f.GetHashBytes(mr)

	// L(x) : Linear combination
	k1 := uint256.NewInt(0)
	kt := f.GetHashBytes(m_root)
	k1.SetBytes8(kt)
	k2 := uint256.NewInt(0)
	kt = f.GetHashBytes(k1.Bytes())
	k2.SetBytes8(kt)
	k3 := uint256.NewInt(0)
	kt = f.GetHashBytes(k2.Bytes())
	k3.SetBytes8(kt)
	k4 := uint256.NewInt(0)
	kt = f.GetHashBytes(k3.Bytes())
	k4.SetBytes8(kt)

	G2_steps := f.GFP.Exp(G2, uint256.NewInt(uint64(f.steps)))
	powers := make([]*uint256.Int, f.extFactor)
	powers[0] = uint256.NewInt(1)
	for i := 1; i < len(powers); i++ {
		powers[i] = f.GFP.Mul(powers[i-1], G2_steps)
	}

	eval_l := make([]*uint256.Int, precision)
	for i := 0; i < precision; i++ {
		eval_l[i] = uint256.NewInt(0)
		p := f.GFP.Add(f.GFP.Mul(eval_oxu[i], k1), f.GFP.Mul(eval_oxu[i], f.GFP.Mul(k2, powers[i%f.extFactor])))
		b := f.GFP.Add(f.GFP.Mul(eval_b[i], k3), f.GFP.Mul(eval_b[i], f.GFP.Mul(k4, powers[i%f.extFactor])))
		eval_l[i] = f.GFP.Add(eval_l[i], p)
		eval_l[i] = f.GFP.Add(eval_l[i], b)
		eval_l[i] = f.GFP.Add(eval_l[i], eval_d[i])
	}

	// Compute Merkle root of L(x)
	tree_l := f.Merklize(eval_l)

	positions := f.GetPseudorandomIndices(m_root, uint32(precision), 80, uint32(f.extFactor))

	var proof dtype.StarksProof
	// proof := make([]interface{}, 8)
	proof.MerkleRoot = m_root
	proof.TreeRoots = [][]byte{tree_oxu[1], tree_poxu[1], tree_kxu[1], tree_d[1], tree_b[1], tree_l[1]}
	proof.TreeOxu = f.MakeMultiBranch(tree_oxu, positions)
	proof.TreePOxu = f.MakeMultiBranch(tree_poxu, positions)
	proof.TreeKxu = f.MakeMultiBranch(tree_kxu, positions)
	proof.TreeD = f.MakeMultiBranch(tree_d, positions)
	proof.TreeB = f.MakeMultiBranch(tree_b, positions)
	proof.TreeL = f.MakeMultiBranch(tree_l, positions)
	proof.VosuFl = []*uint256.Int{oxu[0], oxu[f.steps-1]}
	proof.FriProof = f.ProveLowDegree(eval_l, G2)

	return &proof
}

func (f *starks) VerifyStarksProof(vis []byte, proof *dtype.StarksProof) bool {
	pxu := f.GFP.LoadUint256FromStream31(vis)
	// kxu := f.GFP.LoadUint256FromKey(key)
	// log.Printf("Length of Inputs : ks(%v), vi(%v)", len(kxu), len(pxu))

	for i := len(pxu); i < f.steps; i++ {
		pxu = append(pxu, uint256.NewInt(0))
	}

	precision := f.steps * f.extFactor
	skips := precision / f.steps
	log.Printf("precision : %v, skips : %v", precision, skips)

	g := f.GFP.Prime.Clone()
	g = g.Sub(g, uint256.NewInt(1))
	g = f.GFP.Div(g, uint256.NewInt(uint64(precision)))
	G2 := f.GFP.Exp(uint256.NewInt(7), g)
	// log.Printf("G2 : %v", G2)
	G1 := f.GFP.Exp(G2, uint256.NewInt(uint64(skips)))
	// log.Printf("G1 : %v", G1)

	// skip2 := f.steps / len(kxu)
	// log.Printf("skip for key : %v", skip2)

	// P(x)
	poly_pxu := f.GFP.IDFT(pxu[0:f.steps], G1)
	eval_pxu := f.GFP.DFT(poly_pxu, G2)
	// log.Printf("visu : %v, %v", len(poly_visu), len(eval_visu))

	// K(x)
	// poly_kxu := f.GFP.IDFT(kxu, f.GFP.Exp(G2, f.GFP.Mul(uint256.NewInt(uint64(f.extFactor)), uint256.NewInt(uint64(skip2)))))
	// log.Printf("poly_key %v", len(poly_key))

	// Verify root of trees
	// tree_roots = {tree_vosu[1], tree_d[1], tree_b[1], tree_l[1]}
	mr := proof.TreeRoots[0]
	mr = append(mr, proof.TreeRoots[1]...)
	mr = append(mr, proof.TreeRoots[2]...)
	mr = append(mr, proof.TreeRoots[3]...)
	mr = append(mr, proof.TreeRoots[4]...)
	h1 := hex.EncodeToString(f.GetHashBytes(mr))
	h2 := hex.EncodeToString(proof.MerkleRoot)
	if h1 != h2 {
		log.Printf("Verifying m_root hash : %v != %v", h1, h2)
		return false
	}

	if !f.VerifyLowDegreeProof(proof.TreeRoots[5], proof.FriProof, G2) {
		log.Printf("Low Degree Testing Fail")
		return false
	}

	// L(x) : Linear combination
	k1 := uint256.NewInt(0)
	kt := f.GetHashBytes(proof.MerkleRoot)
	k1.SetBytes8(kt)
	k2 := uint256.NewInt(0)
	kt = f.GetHashBytes(k1.Bytes())
	k2.SetBytes8(kt)
	k3 := uint256.NewInt(0)
	kt = f.GetHashBytes(k2.Bytes())
	k3.SetBytes8(kt)
	k4 := uint256.NewInt(0)
	kt = f.GetHashBytes(k3.Bytes())
	k4.SetBytes8(kt)

	positions := f.GetPseudorandomIndices(proof.MerkleRoot, uint32(precision), 80, uint32(f.extFactor))

	// Compute Io(x)
	lp := uint256.NewInt(uint64((f.steps - 1) * f.extFactor))
	last_step_position := f.GFP.Exp(G2, lp)

	// Io(x)
	xos := []*uint256.Int{uint256.NewInt(1), last_step_position}
	yos := []*uint256.Int{proof.VosuFl[0], proof.VosuFl[1]}
	poly_io := f.GFP.LagrangeInterp(xos, yos)

	// Zo(x)
	f1 := []*uint256.Int{f.GFP.Sub(f.GFP.Prime, uint256.NewInt(1)), uint256.NewInt(1)}  // (x - 1)
	f2 := []*uint256.Int{f.GFP.Sub(f.GFP.Prime, last_step_position), uint256.NewInt(1)} // (x - lastpoint)
	poly_zo := f.GFP.MulPolys(f1, f2)

	// log.Printf("io : %v, zo : %v", poly_io, poly_zo)

	leaves_oxu := f.VerifyMultiBranch(proof.TreeRoots[0], positions, proof.TreeOxu)
	leaves_poxu := f.VerifyMultiBranch(proof.TreeRoots[1], positions, proof.TreePOxu)
	leaves_kxu := f.VerifyMultiBranch(proof.TreeRoots[2], positions, proof.TreeKxu)
	leaves_d := f.VerifyMultiBranch(proof.TreeRoots[3], positions, proof.TreeD)
	leaves_b := f.VerifyMultiBranch(proof.TreeRoots[4], positions, proof.TreeB)
	leaves_l := f.VerifyMultiBranch(proof.TreeRoots[5], positions, proof.TreeL)

	// log.Printf("leaves %v, %v, %v, %v", len(leaves_vosu), len(leaves_d), len(leaves_b), len(leaves_l))

	for i := 0; i < len(positions); i++ {
		power := uint256.NewInt(uint64(positions[i]))
		x := f.GFP.Exp(G2, power)
		x_to_steps := f.GFP.Exp(x, uint256.NewInt(uint64(f.steps)))

		p_x := eval_pxu[positions[i]]
		o_x := f.GFP.IntFromBytes(leaves_oxu[i])
		po_x := f.GFP.IntFromBytes(leaves_poxu[i])
		k_x := f.GFP.IntFromBytes(leaves_kxu[i])
		d_x := f.GFP.IntFromBytes(leaves_d[i])
		b_x := f.GFP.IntFromBytes(leaves_b[i])
		l_x := f.GFP.IntFromBytes(leaves_l[i])

		z := f.GFP.Exp(x, uint256.NewInt(uint64(f.steps)))
		z = f.GFP.Sub(z, uint256.NewInt(1))

		// log.Printf("vi : %v, vo : %v, vo_pre : %v, kx : %v", vi_x, vo_x, vo_xpre, k_x)

		// C(x) = P(x) - (((O(x)^3 + K(x))^3 + O(x/g1))
		// C(x) = Z(x)*D(x)
		c := f.GFP.Exp(o_x, uint256.NewInt(3))
		c = f.GFP.Add(c, k_x)
		c = f.GFP.Exp(c, uint256.NewInt(3))
		c = f.GFP.Add(c, po_x)

		c = f.GFP.Sub(p_x, c)
		q := f.GFP.Mul(d_x, z)

		if c.Cmp(q) != 0 {
			log.Printf("Verification fail : C(x) != Z(x)*D(x), %v != %v", c, q)
			return false
		}

		// Check boundary constraints  B(x) = (Vo(x) - Io(x)) / Zo(x)
		// B(x) * Zo(x) + Io(x) = Vo(x)
		io := f.GFP.EvalPolyAt(poly_io, x)
		zo := f.GFP.EvalPolyAt(poly_zo, x)
		vb := f.GFP.Add(f.GFP.Mul(b_x, zo), io)
		if vb.Cmp(o_x) != 0 {
			log.Printf("Verification fail : B(x) * Zo(x) + Io(x) != Vo(x), %v, %v", vb, o_x)
			return false
		}

		// Check correctness of the linear combination
		p := f.GFP.Add(f.GFP.Mul(o_x, k1), f.GFP.Mul(o_x, f.GFP.Mul(k2, x_to_steps)))
		b := f.GFP.Add(f.GFP.Mul(b_x, k3), f.GFP.Mul(b_x, f.GFP.Mul(k4, x_to_steps)))
		vl := f.GFP.Add(d_x, p)
		vl = f.GFP.Add(vl, b)
		if vl.Cmp(l_x) != 0 {
			log.Printf("Verification fail : Linear combination, %v, %v, %v", positions[i], vl, l_x)
			return false
		}
	}

	log.Println("Verification Success!!!")
	return true
}

// hash : random seed to calculate starting point
// px : Input values
// Vos : Output values
// key : Adress of Prover
func (f *starks) GenerateStarksProofPreKey(hash string, vis []byte, vos []byte, key []byte) *dtype.StarksProof {
	si := poscipher.GetRandIntFromHash(hash)

	pxu := f.GFP.LoadUint256FromStream31(vis)
	oxu := f.GFP.LoadUint256FromStream32(vos)
	kxu := f.GFP.LoadUint256FromStream32(key)
	log.Printf("Size of Inputs ks(%v), vi(%v), vo(%v)", len(kxu), len(pxu), len(oxu))

	klen := len(kxu)
	if len(kxu) < len(pxu) {
		for i := 0; len(kxu) < len(pxu); i++ {
			kxu = append(kxu, kxu[i%klen])
		}
	} else {
		kxu = kxu[:len(pxu)]
	}

	for i := len(pxu); i < f.steps; i++ {
		pxu = append(pxu, uint256.NewInt(0))
	}
	for i := len(oxu); i < f.steps; i++ {
		oxu = append(oxu, uint256.NewInt(0))
	}
	for i := len(kxu); i < f.steps; i++ {
		kxu = append(kxu, uint256.NewInt(0))
	}

	if len(pxu) > f.steps {
		si = si % (len(pxu) - f.steps)
	} else {
		si = 0
	}

	log.Printf("Starting position : %v", si)

	precision := f.steps * f.extFactor
	skips := precision / f.steps
	log.Printf("precision : %v, skips : %v", precision, skips)

	g := f.GFP.Prime.Clone()
	g = g.Sub(g, uint256.NewInt(1))
	g = f.GFP.Div(g, uint256.NewInt(uint64(precision)))
	G2 := f.GFP.Exp(uint256.NewInt(7), g)
	// log.Printf("G2 : %v", G2)
	G1 := f.GFP.Exp(G2, uint256.NewInt(uint64(skips)))
	// log.Printf("G1 : %v", G1)

	// P(x)
	poly_pxu := f.GFP.IDFT(pxu[si:f.steps+si], G1)
	eval_pxu := f.GFP.DFT(poly_pxu, G2)

	// Vo(x)
	poly_oxu := f.GFP.IDFT(oxu[si:f.steps+si], G1)
	eval_oxu := f.GFP.DFT(poly_oxu, G2)

	// K(x)
	poly_kxu := f.GFP.IDFT(kxu[si:f.steps+si], G1)
	eval_kxu := f.GFP.DFT(poly_kxu, G2)

	// Polynomial of O(x/g1)
	poxu := make([]*uint256.Int, f.steps)
	if si == 0 {
		poxu[0] = uint256.NewInt(0)
	} else {
		poxu[0] = oxu[si-1]
	}
	for i := 1; i < f.steps; i++ {
		poxu[i] = oxu[si+i-1]
	}
	poly_poxu := f.GFP.IDFT(poxu, G1)
	eval_poxu := f.GFP.DFT(poly_poxu, G2)

	// C(x) = C(P(x), O(x), K(x))
	// O[i]^3 * O[i-1] - K[i] = P[i]
	// C(x) = P(x) - (O(x)^3 * O(x/g1) - K(x))
	eval_cp := make([]*uint256.Int, precision)
	for i := 0; i < precision; i++ {
		c := eval_oxu[i].Clone()
		c = f.GFP.Exp(c, uint256.NewInt(3))
		c = f.GFP.Add(c, eval_kxu[i])
		c = f.GFP.Exp(c, uint256.NewInt(3))
		c = f.GFP.Add(c, eval_poxu[i])

		eval_cp[i] = f.GFP.Sub(eval_pxu[i], c)

	}

	size, xs := f.GFP.ExtRootUnity(G2, false) // return from x^0 to x^n, x^n == x^0
	xs = xs[:size-1]
	// log.Printf("size : %v, xs[0] : %v, xs[-1] : %v", size, xs[0], xs[len(xs)-1])

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
	yos := []*uint256.Int{oxu[si], oxu[si+f.steps-1]}

	// Io(x)
	poly_io := f.GFP.LagrangeInterp(xos, yos)
	eval_io := make([]*uint256.Int, precision)
	for i := 0; i < precision; i++ {
		eval_io[i] = f.GFP.EvalPolyAt(poly_io, xs[i])
	}

	f1 := []*uint256.Int{f.GFP.Sub(f.GFP.Prime, uint256.NewInt(1)), uint256.NewInt(1)}           // (x - 1)
	f2 := []*uint256.Int{f.GFP.Sub(f.GFP.Prime, xs[(f.steps-1)*f.extFactor]), uint256.NewInt(1)} // (x - lastpoint)

	start := time.Now().UnixNano()
	log.Printf("GenerateStarksProof 1: %v", time.Now().UnixNano()-start)
	// Zo(x)
	poly_zo := f.GFP.MulPolys(f1, f2)
	// eval_inv_zo := make([]*uint256.Int, precision)
	// for i := 0; i < precision; i++ {
	// 	eval_inv_zo[i] = f.GFP.Inv(f.GFP.EvalPolyAt(poly_zo, xs[i]))
	// }
	eval_zo := make([]*uint256.Int, precision)
	for i := 0; i < precision; i++ {
		eval_zo[i] = f.GFP.EvalPolyAt(poly_zo, xs[i])
	}
	eval_inv_zo := f.GFP.MultInv(eval_zo)
	log.Printf("GenerateStarksProof 2: %v", time.Now().UnixNano()-start)

	// B(x) = (Vo(x) - Io(x)) / Zo(x)
	eval_b := make([]*uint256.Int, precision)
	for i := 0; i < precision; i++ {
		eval_b[i] = f.GFP.Sub(eval_oxu[i], eval_io[i])
		eval_b[i] = f.GFP.Mul(eval_b[i], eval_inv_zo[i])
	}

	// Compute Merkle Root
	tree_oxu := f.Merklize(eval_oxu)
	tree_poxu := f.Merklize(eval_poxu)
	tree_kxu := f.Merklize(eval_kxu)
	tree_d := f.Merklize(eval_d)
	tree_b := f.Merklize(eval_b)
	mr := tree_oxu[1]
	mr = append(mr, tree_poxu[1]...)
	mr = append(mr, tree_kxu[1]...)
	mr = append(mr, tree_d[1]...)
	mr = append(mr, tree_b[1]...)
	m_root := f.GetHashBytes(mr)

	k1 := uint256.NewInt(0)
	kt := f.GetHashBytes(m_root)
	k1.SetBytes8(kt)
	k2 := uint256.NewInt(0)
	kt = f.GetHashBytes(k1.Bytes())
	k2.SetBytes8(kt)
	k3 := uint256.NewInt(0)
	kt = f.GetHashBytes(k2.Bytes())
	k3.SetBytes8(kt)
	k4 := uint256.NewInt(0)
	kt = f.GetHashBytes(k3.Bytes())
	k4.SetBytes8(kt)

	G2_steps := f.GFP.Exp(G2, uint256.NewInt(uint64(f.steps)))
	powers := make([]*uint256.Int, f.extFactor)
	powers[0] = uint256.NewInt(1)
	for i := 1; i < len(powers); i++ {
		powers[i] = f.GFP.Mul(powers[i-1], G2_steps)
	}

	eval_l := make([]*uint256.Int, precision)
	for i := 0; i < precision; i++ {
		eval_l[i] = uint256.NewInt(0)
		p := f.GFP.Add(f.GFP.Mul(eval_oxu[i], k1), f.GFP.Mul(eval_oxu[i], f.GFP.Mul(k2, powers[i%f.extFactor])))
		b := f.GFP.Add(f.GFP.Mul(eval_b[i], k3), f.GFP.Mul(eval_b[i], f.GFP.Mul(k4, powers[i%f.extFactor])))
		eval_l[i] = f.GFP.Add(eval_l[i], p)
		eval_l[i] = f.GFP.Add(eval_l[i], b)
		eval_l[i] = f.GFP.Add(eval_l[i], eval_d[i])
	}

	// Compute Merkle root of L(x)
	tree_l := f.Merklize(eval_l)

	positions := f.GetPseudorandomIndices(m_root, uint32(precision), 80, uint32(f.extFactor))

	// proof := make([]interface{}, 9)
	var proof dtype.StarksProof
	proof.RandomHash = hash
	proof.MerkleRoot = m_root
	proof.TreeRoots = [][]byte{tree_oxu[1], tree_poxu[1], tree_kxu[1], tree_d[1], tree_b[1], tree_l[1]}
	proof.TreeOxu = f.MakeMultiBranch(tree_oxu, positions)
	proof.TreePOxu = f.MakeMultiBranch(tree_poxu, positions)
	proof.TreeKxu = f.MakeMultiBranch(tree_kxu, positions)
	proof.TreeD = f.MakeMultiBranch(tree_d, positions)
	proof.TreeB = f.MakeMultiBranch(tree_b, positions)
	proof.TreeL = f.MakeMultiBranch(tree_l, positions)
	proof.VosuFl = []*uint256.Int{oxu[si], oxu[si+f.steps-1]}
	proof.FriProof = f.ProveLowDegree(eval_l, G2)

	return &proof
}

func (f *starks) VerifyStarksProofPreKey(vis []byte, proof *dtype.StarksProof) bool {
	pxu := f.GFP.LoadUint256FromStream31(vis)
	log.Printf("Length of Inputs : vi(%v)", len(pxu))

	for i := len(pxu); i < f.steps; i++ {
		pxu = append(pxu, uint256.NewInt(0))
	}

	precision := f.steps * f.extFactor
	skips := precision / f.steps
	log.Printf("precision : %v, skips : %v", precision, skips)

	g := f.GFP.Prime.Clone()
	g = g.Sub(g, uint256.NewInt(1))
	g = f.GFP.Div(g, uint256.NewInt(uint64(precision)))
	G2 := f.GFP.Exp(uint256.NewInt(7), g)
	// log.Printf("G2 : %v", G2)
	G1 := f.GFP.Exp(G2, uint256.NewInt(uint64(skips)))
	// log.Printf("G1 : %v", G1)

	si := poscipher.GetRandIntFromHash(proof.RandomHash)
	if len(pxu) > f.steps {
		si = si % (len(pxu) - f.steps)
	} else {
		si = 0
	}

	log.Printf("Starting position : %v", si)

	// P(x)
	poly_pxu := f.GFP.IDFT(pxu[si:si+f.steps], G1)
	eval_pxu := f.GFP.DFT(poly_pxu, G2)

	// Verify root of trees
	// tree_roots = {tree_oxu[1], tree_poxu[1], tree_kxu[1], tree_d[1], tree_b[1], tree_l[1]}
	mr := proof.TreeRoots[0]
	mr = append(mr, proof.TreeRoots[1]...)
	mr = append(mr, proof.TreeRoots[2]...)
	mr = append(mr, proof.TreeRoots[3]...)
	mr = append(mr, proof.TreeRoots[4]...)
	h1 := hex.EncodeToString(f.GetHashBytes(mr))
	h2 := hex.EncodeToString(proof.MerkleRoot)
	if h1 != h2 {
		log.Printf("Verifying m_root hash : %v != %v", h1, h2)
		return false
	}

	if !f.VerifyLowDegreeProof(proof.TreeRoots[5], proof.FriProof, G2) {
		log.Printf("Low Degree Testing Fail")
		return false
	}

	k1 := uint256.NewInt(0)
	kt := f.GetHashBytes(proof.MerkleRoot)
	k1.SetBytes8(kt)
	k2 := uint256.NewInt(0)
	kt = f.GetHashBytes(k1.Bytes())
	k2.SetBytes8(kt)
	k3 := uint256.NewInt(0)
	kt = f.GetHashBytes(k2.Bytes())
	k3.SetBytes8(kt)
	k4 := uint256.NewInt(0)
	kt = f.GetHashBytes(k3.Bytes())
	k4.SetBytes8(kt)

	positions := f.GetPseudorandomIndices(proof.MerkleRoot, uint32(precision), 80, uint32(f.extFactor))

	// Compute Io(x)
	lp := uint256.NewInt(uint64((f.steps - 1) * f.extFactor))
	last_step_position := f.GFP.Exp(G2, lp)

	// Io(x)
	xos := []*uint256.Int{uint256.NewInt(1), last_step_position}
	yos := []*uint256.Int{proof.VosuFl[0], proof.VosuFl[1]}
	poly_io := f.GFP.LagrangeInterp(xos, yos)

	// Zo(x)
	f1 := []*uint256.Int{f.GFP.Sub(f.GFP.Prime, uint256.NewInt(1)), uint256.NewInt(1)}  // (x - 1)
	f2 := []*uint256.Int{f.GFP.Sub(f.GFP.Prime, last_step_position), uint256.NewInt(1)} // (x - lastpoint)
	poly_zo := f.GFP.MulPolys(f1, f2)

	// log.Printf("io : %v, zo : %v", poly_io, poly_zo)

	leaves_oxu := f.VerifyMultiBranch(proof.TreeRoots[0], positions, proof.TreeOxu)
	leaves_poxu := f.VerifyMultiBranch(proof.TreeRoots[1], positions, proof.TreePOxu)
	leaves_kxu := f.VerifyMultiBranch(proof.TreeRoots[2], positions, proof.TreeKxu)
	leaves_d := f.VerifyMultiBranch(proof.TreeRoots[3], positions, proof.TreeD)
	leaves_b := f.VerifyMultiBranch(proof.TreeRoots[4], positions, proof.TreeB)
	leaves_l := f.VerifyMultiBranch(proof.TreeRoots[5], positions, proof.TreeL)

	// log.Printf("leaves %v, %v, %v, %v", len(leaves_vosu), len(leaves_d), len(leaves_b), len(leaves_l))

	for i := 0; i < len(positions); i++ {
		power := uint256.NewInt(uint64(positions[i]))
		x := f.GFP.Exp(G2, power)
		x_to_steps := f.GFP.Exp(x, uint256.NewInt(uint64(f.steps)))

		vi_x := eval_pxu[positions[i]]
		vo_x := f.GFP.IntFromBytes(leaves_oxu[i])
		vo_xpre := f.GFP.IntFromBytes(leaves_poxu[i])
		k_x := f.GFP.IntFromBytes(leaves_kxu[i])
		d_x := f.GFP.IntFromBytes(leaves_d[i])
		b_x := f.GFP.IntFromBytes(leaves_b[i])
		l_x := f.GFP.IntFromBytes(leaves_l[i])

		z := f.GFP.Exp(x, uint256.NewInt(uint64(f.steps)))
		z = f.GFP.Sub(z, uint256.NewInt(1))
		// k_x := f.GFP.EvalPolyAt(poly_key, f.GFP.Exp(x, uint256.NewInt(uint64(skip2))))

		// log.Printf("vi : %v, vo : %v, vo_pre : %v, kx : %v", vi_x, vo_x, vo_xpre, k_x)

		// C(x) = P(x) - (((O(x)^3 + K(x))^3 + O(x/g1))
		// C(x) = Z(x)*D(x)
		c := vo_x.Clone()
		c = f.GFP.Exp(c, uint256.NewInt(3))
		c = f.GFP.Add(c, k_x)
		c = f.GFP.Exp(c, uint256.NewInt(3))
		c = f.GFP.Add(c, vo_xpre)

		c = f.GFP.Sub(vi_x, c)
		q := f.GFP.Mul(d_x, z)

		if c.Cmp(q) != 0 {
			log.Printf("Verification fail : C(x) != Z(x)*D(x), %v != %v", c, q)
			return false
		}

		// Check boundary constraints  B(x) = (Vo(x) - Io(x)) / Zo(x)
		// B(x) * Zo(x) + Io(x) = Vo(x)
		io := f.GFP.EvalPolyAt(poly_io, x)
		zo := f.GFP.EvalPolyAt(poly_zo, x)
		vb := f.GFP.Add(f.GFP.Mul(b_x, zo), io)
		if vb.Cmp(vo_x) != 0 {
			log.Printf("Verification fail : B(x) * Zo(x) + Io(x) != Vo(x), %v, %v", vb, vo_x)
			return false
		}

		// Check correctness of the linear combination
		p := f.GFP.Add(f.GFP.Mul(vo_x, k1), f.GFP.Mul(vo_x, f.GFP.Mul(k2, x_to_steps)))
		b := f.GFP.Add(f.GFP.Mul(b_x, k3), f.GFP.Mul(b_x, f.GFP.Mul(k4, x_to_steps)))
		vl := uint256.NewInt(0)
		vl = f.GFP.Add(vl, d_x)
		vl = f.GFP.Add(vl, p)
		vl = f.GFP.Add(vl, b)
		if vl.Cmp(l_x) != 0 {
			log.Printf("Verification fail : Linear combination, %v, %v, %v", positions[i], vl, l_x)
			return false
		}
	}

	log.Println("Verification Success!!!")
	return true
}

func (f *starks) GetSizeStarksProofPreKey(proof *dtype.StarksProof) int {
	size := 0

	size += len(proof.MerkleRoot)

	for i := 0; i < len(proof.TreeRoots); i++ {
		size += len(proof.TreeRoots[i])
	}
	// log.Printf("sizof buf : %v", size)

	for i := 0; i < len(proof.TreeOxu); i++ {
		for j := 0; j < len(proof.TreeOxu[i]); j++ {
			size += len(proof.TreeOxu[i][j])
		}
	}

	for i := 0; i < len(proof.TreePOxu); i++ {
		for j := 0; j < len(proof.TreePOxu[i]); j++ {
			size += len(proof.TreePOxu[i][j])
		}
	}

	// log.Printf("sizof buf : %v", size)
	for i := 0; i < len(proof.TreeKxu); i++ {
		for j := 0; j < len(proof.TreeKxu[i]); j++ {
			size += len(proof.TreeKxu[i][j])
		}
	}

	// log.Printf("sizof buf : %v", size)
	for i := 0; i < len(proof.TreeD); i++ {
		for j := 0; j < len(proof.TreeD[i]); j++ {
			size += len(proof.TreeD[i][j])
		}
	}

	// log.Printf("sizof buf : %v", size)
	for i := 0; i < len(proof.TreeB); i++ {
		for j := 0; j < len(proof.TreeB[i]); j++ {
			size += len(proof.TreeB[i][j])
		}
	}

	// log.Printf("sizof buf : %v", size)
	for i := 0; i < len(proof.TreeL); i++ {
		for j := 0; j < len(proof.TreeL[i]); j++ {
			size += len(proof.TreeL[i][j])
		}
	}

	// log.Printf("sizof buf : %v", size)
	size += len(proof.VosuFl) * 4 * 8

	// log.Printf("sizof buf : %v", size)
	for i := 0; i < len(proof.FriProof)-1; i++ {
		p := proof.FriProof[i]
		root2 := p.Root2[0]
		size += len(root2)

		cbranch := p.CBranch
		for i := 0; i < len(cbranch); i++ {
			for j := 0; j < len(cbranch[i]); j++ {
				size += len(cbranch[i][j])
			}
		}

		pbranch := p.PBranch
		for i := 0; i < len(pbranch); i++ {
			for j := 0; j < len(pbranch[i]); j++ {
				size += len(pbranch[i][j])
			}
		}
	}

	log.Printf("sizof buf : %v", size)
	rootl := proof.FriProof[len(proof.FriProof)-1].Root2
	for i := 0; i < len(rootl); i++ {
		size += len(rootl[i])
	}

	// buf_proof, err := json.Marshal(proof)
	// if err != nil {
	// 	log.Printf("json.Marshel error : %v", err)
	// }
	// log.Printf("sizof buf : %v", len(buf_proof))

	return size
}
