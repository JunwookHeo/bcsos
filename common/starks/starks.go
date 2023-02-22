package starks

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"log"
	"time"

	"github.com/holiman/uint256"
	"github.com/junwookheo/bcsos/common/blockchain"
	"github.com/junwookheo/bcsos/common/dtype"
	"github.com/junwookheo/bcsos/common/galois"
)

type starks struct {
	GFP       *galois.GFP
	numIdx    int
	steps     int
	extFactor int
}

const PSIZE = 31

func NewStarks(steps int) *starks {
	if steps == 0 {
		steps = 2048 // 8192 / 4
	}
	f := starks{GFP: galois.NewGFP(), numIdx: 40, steps: steps, extFactor: 8}

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
	if (L >> 2) <= 16 {
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

	log.Printf("Prove values with length : %v", L)
	size, xxs := f.GFP.ExtRootUnity(rou, false)
	if L != size-1 {
		log.Panicf("Mismatch the size of values and xs : %v, %v", L, len(xxs)-1)
		return nil
	}

	m1 := f.Merklize(values)
	special_x := f.GFP.IntFromBytes(m1[1])
	quarter_len := len(xxs) >> 2

	start := time.Now().UnixNano()
	log.Printf("ProveLowDegree 1: %v", time.Now().UnixNano()-start)

	x_polys := make([][]*uint256.Int, quarter_len)
	dxs := make([][]*uint256.Int, quarter_len)
	dys := make([][]*uint256.Int, quarter_len)
	colums := make([]*uint256.Int, quarter_len)
	for i := 0; i < quarter_len; i++ {
		xs := []*uint256.Int{xxs[i], xxs[i+quarter_len], xxs[i+2*quarter_len], xxs[i+3*quarter_len]}
		ys := []*uint256.Int{values[i], values[i+quarter_len], values[i+2*quarter_len], values[i+3*quarter_len]}
		// x_poly := f.GFP.LagrangeInterp(xs[:], ys[:])
		x_poly := f.GFP.LagrangeInterp_4(xs[:], ys[:])
		colums[i] = f.GFP.EvalPolyAt(x_poly, special_x)
		x_polys[i] = x_poly
		dxs[i] = xs
		dys[i] = ys
	}
	log.Printf("ProveLowDegree 2: %v", time.Now().UnixNano()-start)
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
		log.Printf("Verify values with length : %v", roudeg)
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

	log.Println("Evaluation success")
	return true
}

// Vis : Input values
// Vos : Output values
// key : Adress of Prover
func (f *starks) GenerateStarksProof(vis []byte, vos []byte, key []byte) *dtype.StarksProof {
	visu := f.GFP.LoadUint256FromStream31(vis)
	vosu := f.GFP.LoadUint256FromStream32(vos)
	ks := f.GFP.LoadUint256FromKey(key)
	log.Printf("Size of Inputs ks(%v), vi(%v), vo(%v)", len(ks), len(visu), len(vosu))

	for i := len(visu); i < f.steps; i++ {
		visu = append(visu, uint256.NewInt(0))
	}
	for i := len(vosu); i < f.steps; i++ {
		vosu = append(vosu, uint256.NewInt(0))
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

	// Vi(x)
	poly_visu := f.GFP.IDFT(visu[0:f.steps], G1)
	eval_visu := f.GFP.DFT(poly_visu, G2)
	// log.Printf("visu : %v, %v", len(poly_visu), len(eval_visu))
	// for i := 0; i < 10; i++ {
	// 	log.Printf("eval_visu : %v, %v", visu[i], eval_visu[i*f.extFactor])
	// }

	// Vo(x)
	poly_vosu := f.GFP.IDFT(vosu[0:f.steps], G1)
	eval_vosu := f.GFP.DFT(poly_vosu, G2)
	// log.Printf("vosu : %v, %v", len(poly_vosu), len(eval_vosu))
	// for i := 0; i < 10; i++ {
	// 	log.Printf("eval_vosu : %v, %v", vosu[i], eval_vosu[i*f.extFactor])
	// }

	skip2 := f.steps / len(ks)
	// log.Printf("skip for key : %v", skip2)

	// K(x)
	poly_key := f.GFP.IDFT(ks, f.GFP.Exp(G1, uint256.NewInt(uint64(skip2))))
	eval_key := f.GFP.DFT(poly_key, f.GFP.Exp(G2, uint256.NewInt(uint64(skip2))))
	// log.Printf("skip for key : %v", len(eval_key))

	// C(x) = C(Vi(x), Vo(x), K(x))
	// Vo[i]^3 * Vo[i-1] - K[i] = Vi[i]
	// C(x) = Vi(x) - (Vo(x)^3 * Vo(x/g1) - K(x))
	eval_cp := make([]*uint256.Int, precision)
	pre := uint256.NewInt(1)
	for i := 0; i < precision; i++ {
		c := f.GFP.Exp(eval_vosu[i], uint256.NewInt(3))
		c = f.GFP.Mul(c, pre)
		// c = f.GFP.Sub(c, eval_key[i%len(eval_key)])
		if !eval_key[i%len(eval_key)].IsZero() {
			c = f.GFP.Mul(c, eval_key[i%len(eval_key)])
		}
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
	yos := []*uint256.Int{vosu[0], vosu[f.steps-1]}

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

	positions := f.GetPseudorandomIndices(m_root, uint32(precision), 80, uint32(f.extFactor))
	augmented_positions := make([]uint32, len(positions)*2)
	for i := 0; i < len(positions); i++ {
		// In this case, we need previous Vo(x) for encryption/decryption
		augmented_positions[i*2] = positions[i]
		if positions[i] >= uint32(skips) {
			augmented_positions[i*2+1] = positions[i] - uint32(skips)
		} else {
			augmented_positions[i*2+1] = uint32(precision) + positions[i] - uint32(skips)
		}
	}
	// log.Printf("positions : %v", positions)
	// log.Printf("augmented_positions : %v", augmented_positions)
	// log.Printf("tree_l : %v", tree_l[1])

	var proof dtype.StarksProof
	// proof := make([]interface{}, 8)
	proof.MerkleRoot = m_root
	proof.TreeRoots = [][]byte{tree_vosu[1], tree_d[1], tree_b[1], tree_l[1]}
	proof.TreeVosu = f.MakeMultiBranch(tree_vosu, augmented_positions)
	proof.TreeD = f.MakeMultiBranch(tree_d, augmented_positions)
	proof.TreeB = f.MakeMultiBranch(tree_b, augmented_positions)
	proof.TreeL = f.MakeMultiBranch(tree_l, positions)
	proof.VosuFl = []*uint256.Int{vosu[0], vosu[f.steps-1]}
	proof.FriProof = f.ProveLowDegree(eval_l, G2)

	return &proof
}

func (f *starks) VerifyStarksProof(vis []byte, key []byte, proof *dtype.StarksProof) bool {
	visu := f.GFP.LoadUint256FromStream31(vis)
	ks := f.GFP.LoadUint256FromKey(key)
	log.Printf("Length of Inputs : ks(%v), vi(%v)", len(ks), len(visu))

	for i := len(visu); i < f.steps; i++ {
		visu = append(visu, uint256.NewInt(0))
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

	skip2 := f.steps / len(ks)
	// log.Printf("skip for key : %v", skip2)

	// Vi(x)
	poly_visu := f.GFP.IDFT(visu[0:f.steps], G1)
	eval_visu := f.GFP.DFT(poly_visu, G2)
	// log.Printf("visu : %v, %v", len(poly_visu), len(eval_visu))

	// K(x)
	poly_key := f.GFP.IDFT(ks, f.GFP.Exp(G2, f.GFP.Mul(uint256.NewInt(uint64(f.extFactor)), uint256.NewInt(uint64(skip2)))))
	// log.Printf("poly_key %v", len(poly_key))

	m_root := proof.MerkleRoot
	tree_roots := proof.TreeRoots
	tree_vosu := proof.TreeVosu
	tree_d := proof.TreeD
	tree_b := proof.TreeB
	tree_l := proof.TreeL
	vosu_fl := proof.VosuFl
	fri_proof := proof.FriProof

	// Verify root of trees
	// tree_roots = {tree_vosu[1], tree_d[1], tree_b[1], tree_l[1]}
	mr := tree_roots[0]
	mr = append(mr, tree_roots[1]...)
	mr = append(mr, tree_roots[2]...)
	h1 := hex.EncodeToString(f.GetHashBytes(mr))
	h2 := hex.EncodeToString(m_root)
	if h1 != h2 {
		log.Printf("Verifying m_root hash : %v != %v", h1, h2)
		return false
	}

	if !f.VerifyLowDegreeProof(tree_roots[3], fri_proof, G2) {
		log.Printf("Low Degree Testing Fail")
		return false
	}

	// L(x) : Linear combination
	k1 := uint256.NewInt(0)
	k1.SetBytes8(m_root[0:])
	k2 := uint256.NewInt(0)
	k2.SetBytes8(m_root[8:])
	k3 := uint256.NewInt(0)
	k3.SetBytes8(m_root[16:])
	k4 := uint256.NewInt(0)
	k4.SetBytes8(m_root[24:])

	positions := f.GetPseudorandomIndices(m_root, uint32(precision), 80, uint32(f.extFactor))
	augmented_positions := make([]uint32, len(positions)*2)

	for i := 0; i < len(positions); i++ {
		// In this case, we need previous Vo(x) for encryption/decryption
		augmented_positions[i*2] = positions[i]
		if positions[i] >= uint32(skips) {
			augmented_positions[i*2+1] = positions[i] - uint32(skips)
		} else {
			augmented_positions[i*2+1] = uint32(precision) + positions[i] - uint32(skips)
		}
	}

	// Compute Io(x)
	lp := uint256.NewInt(uint64((f.steps - 1) * f.extFactor))
	last_step_position := f.GFP.Exp(G2, lp)

	// Io(x)
	xos := []*uint256.Int{uint256.NewInt(1), last_step_position}
	yos := []*uint256.Int{vosu_fl[0], vosu_fl[1]}
	poly_io := f.GFP.LagrangeInterp(xos, yos)

	// Zo(x)
	f1 := []*uint256.Int{f.GFP.Sub(f.GFP.Prime, uint256.NewInt(1)), uint256.NewInt(1)}  // (x - 1)
	f2 := []*uint256.Int{f.GFP.Sub(f.GFP.Prime, last_step_position), uint256.NewInt(1)} // (x - lastpoint)
	poly_zo := f.GFP.MulPolys(f1, f2)

	// log.Printf("io : %v, zo : %v", poly_io, poly_zo)

	leaves_vosu := f.VerifyMultiBranch(tree_roots[0], augmented_positions, tree_vosu)
	leaves_d := f.VerifyMultiBranch(tree_roots[1], augmented_positions, tree_d)
	leaves_b := f.VerifyMultiBranch(tree_roots[2], augmented_positions, tree_b)
	leaves_l := f.VerifyMultiBranch(tree_roots[3], positions, tree_l)

	// log.Printf("leaves %v, %v, %v, %v", len(leaves_vosu), len(leaves_d), len(leaves_b), len(leaves_l))

	for i := 0; i < len(positions); i++ {
		power := uint256.NewInt(uint64(positions[i]))
		x := f.GFP.Exp(G2, power)
		x_to_steps := f.GFP.Exp(x, uint256.NewInt(uint64(f.steps)))

		vi_x := eval_visu[positions[i]]
		vo_x := f.GFP.IntFromBytes(leaves_vosu[i*2])
		vo_xpre := f.GFP.IntFromBytes(leaves_vosu[i*2+1])
		d_x := f.GFP.IntFromBytes(leaves_d[i*2])
		b_x := f.GFP.IntFromBytes(leaves_b[i*2])
		l_x := f.GFP.IntFromBytes(leaves_l[i])

		z := f.GFP.Exp(x, uint256.NewInt(uint64(f.steps)))
		z = f.GFP.Sub(z, uint256.NewInt(1))
		k_x := f.GFP.EvalPolyAt(poly_key, f.GFP.Exp(x, uint256.NewInt(uint64(skip2))))

		// log.Printf("vi : %v, vo : %v, vo_pre : %v, kx : %v", vi_x, vo_x, vo_xpre, k_x)

		// C(x) = Vi(x) - (Vo(x)^3 * Vo(x/g1) - K(x))
		// C(x) = Z(x)*D(x)
		c := f.GFP.Exp(vo_x, uint256.NewInt(3))
		if !vo_xpre.IsZero() {
			c = f.GFP.Mul(c, vo_xpre)
		}
		// c = f.GFP.Sub(c, k_x)
		if !k_x.IsZero() {
			c = f.GFP.Mul(c, k_x)
		}
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

// Vis : Input values
// Vos : Output values
// key : Adress of Prover
func (f *starks) GenerateStarksProofPreKey(vis []byte, vos []byte, key []byte) *dtype.StarksProof {
	visu := f.GFP.LoadUint256FromStream31(vis)
	vosu := f.GFP.LoadUint256FromStream32(vos)
	ks := f.GFP.LoadUint256FromStream32(key)
	log.Printf("Size of Inputs ks(%v), vi(%v), vo(%v)", len(ks), len(visu), len(vosu))

	if len(ks) < len(visu) {
		for i := 0; len(ks) < len(visu); i++ {
			ks = append(ks, ks[i%len(ks)])
		}
	} else {
		ks = ks[:len(visu)]
	}

	for i := len(visu); i < f.steps; i++ {
		visu = append(visu, uint256.NewInt(0))
	}
	for i := len(vosu); i < f.steps; i++ {
		vosu = append(vosu, uint256.NewInt(0))
	}
	for i := len(ks); i < f.steps; i++ {
		ks = append(ks, uint256.NewInt(0))
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

	// Vi(x)
	poly_visu := f.GFP.IDFT(visu[0:f.steps], G1)
	eval_visu := f.GFP.DFT(poly_visu, G2)
	// log.Printf("visu : %v, %v", len(poly_visu), len(eval_visu))
	// for i := 0; i < 10; i++ {
	// 	log.Printf("eval_visu : %v, %v", visu[i], eval_visu[i*f.extFactor])
	// }

	// Vo(x)
	poly_vosu := f.GFP.IDFT(vosu[0:f.steps], G1)
	eval_vosu := f.GFP.DFT(poly_vosu, G2)
	// log.Printf("vosu : %v, %v", len(poly_vosu), len(eval_vosu))
	// for i := 0; i < 10; i++ {
	// 	log.Printf("eval_vosu : %v, %v", vosu[i], eval_vosu[i*f.extFactor])
	// }

	// K(x)
	poly_key := f.GFP.IDFT(ks[0:f.steps], G1)
	eval_key := f.GFP.DFT(poly_key, G2)
	// log.Printf("skip for key : %v", len(eval_key))

	// C(x) = C(Vi(x), Vo(x), K(x))
	// Vo[i]^3 * Vo[i-1] - K[i] = Vi[i]
	// C(x) = Vi(x) - (Vo(x)^3 * Vo(x/g1) - K(x))
	eval_cp := make([]*uint256.Int, precision)
	pre := uint256.NewInt(1)
	for i := 0; i < precision; i++ {
		c := f.GFP.Exp(eval_vosu[i], uint256.NewInt(3))
		c = f.GFP.Mul(c, pre)
		// c = f.GFP.Sub(c, eval_key[i])
		if !eval_key[i].IsZero() {
			c = f.GFP.Mul(c, eval_key[i])
		}
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
	yos := []*uint256.Int{vosu[0], vosu[f.steps-1]}

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
		eval_b[i] = f.GFP.Sub(eval_vosu[i], eval_io[i])
		eval_b[i] = f.GFP.Mul(eval_b[i], eval_inv_zo[i])
	}

	// Compute Merkle Root
	tree_vosu := f.Merklize(eval_vosu)
	tree_key := f.Merklize(eval_key)
	tree_d := f.Merklize(eval_d)
	tree_b := f.Merklize(eval_b)
	mr := tree_vosu[1]
	mr = append(mr, tree_key[1]...)
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

	positions := f.GetPseudorandomIndices(m_root, uint32(precision), 80, uint32(f.extFactor))
	augmented_positions := make([]uint32, len(positions)*2)
	for i := 0; i < len(positions); i++ {
		// In this case, we need previous Vo(x) for encryption/decryption
		augmented_positions[i*2] = positions[i]
		if positions[i] >= uint32(skips) {
			augmented_positions[i*2+1] = positions[i] - uint32(skips)
		} else {
			augmented_positions[i*2+1] = uint32(precision) + positions[i] - uint32(skips)
		}
	}
	// log.Printf("positions : %v", positions)
	// log.Printf("augmented_positions : %v", augmented_positions)
	// log.Printf("tree_l : %v", tree_l[1])

	// proof := make([]interface{}, 9)
	var proof dtype.StarksProof
	proof.MerkleRoot = m_root
	proof.TreeRoots = [][]byte{tree_vosu[1], tree_key[1], tree_d[1], tree_b[1], tree_l[1]}
	proof.TreeVosu = f.MakeMultiBranch(tree_vosu, augmented_positions)
	proof.TreeKey = f.MakeMultiBranch(tree_key, augmented_positions)
	proof.TreeD = f.MakeMultiBranch(tree_d, augmented_positions)
	proof.TreeB = f.MakeMultiBranch(tree_b, augmented_positions)
	proof.TreeL = f.MakeMultiBranch(tree_l, positions)
	proof.VosuFl = []*uint256.Int{vosu[0], vosu[f.steps-1]}
	proof.FriProof = f.ProveLowDegree(eval_l, G2)

	return &proof
}

func (f *starks) VerifyStarksProofPreKey(vis []byte, proof *dtype.StarksProof) bool {
	visu := f.GFP.LoadUint256FromStream31(vis)
	log.Printf("Length of Inputs : vi(%v)", len(visu))

	for i := len(visu); i < f.steps; i++ {
		visu = append(visu, uint256.NewInt(0))
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

	// Vi(x)
	poly_visu := f.GFP.IDFT(visu[0:f.steps], G1)
	eval_visu := f.GFP.DFT(poly_visu, G2)
	// log.Printf("visu : %v, %v", len(poly_visu), len(eval_visu))

	// m_root, _ := proof.MerkleRoot
	// tree_roots, _ := proof[1].([][]byte)
	// tree_vosu, _ := proof[2].([][][]byte)
	// tree_key, _ := proof[3].([][][]byte)
	// tree_d, _ := proof[4].([][][]byte)
	// tree_b, _ := proof[5].([][][]byte)
	// tree_l, _ := proof[6].([][][]byte)
	// vosu_fl, _ := proof[7].([]*uint256.Int)
	// fri_proof, _ := proof[8].([]interface{})

	// Verify root of trees
	// tree_roots = {tree_vosu[1], tree_key[1], tree_d[1], tree_b[1], tree_l[1]}
	mr := proof.TreeRoots[0]
	mr = append(mr, proof.TreeRoots[1]...)
	mr = append(mr, proof.TreeRoots[2]...)
	mr = append(mr, proof.TreeRoots[3]...)
	h1 := hex.EncodeToString(f.GetHashBytes(mr))
	h2 := hex.EncodeToString(proof.MerkleRoot)
	if h1 != h2 {
		log.Printf("Verifying m_root hash : %v != %v", h1, h2)
		return false
	}

	if !f.VerifyLowDegreeProof(proof.TreeRoots[4], proof.FriProof, G2) {
		log.Printf("Low Degree Testing Fail")
		return false
	}

	// L(x) : Linear combination
	k1 := uint256.NewInt(0)
	k1.SetBytes8(proof.MerkleRoot[0:])
	k2 := uint256.NewInt(0)
	k2.SetBytes8(proof.MerkleRoot[8:])
	k3 := uint256.NewInt(0)
	k3.SetBytes8(proof.MerkleRoot[16:])
	k4 := uint256.NewInt(0)
	k4.SetBytes8(proof.MerkleRoot[24:])

	positions := f.GetPseudorandomIndices(proof.MerkleRoot, uint32(precision), 80, uint32(f.extFactor))
	augmented_positions := make([]uint32, len(positions)*2)

	for i := 0; i < len(positions); i++ {
		// In this case, we need previous Vo(x) for encryption/decryption
		augmented_positions[i*2] = positions[i]
		if positions[i] >= uint32(skips) {
			augmented_positions[i*2+1] = positions[i] - uint32(skips)
		} else {
			augmented_positions[i*2+1] = uint32(precision) + positions[i] - uint32(skips)
		}
	}

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

	leaves_vosu := f.VerifyMultiBranch(proof.TreeRoots[0], augmented_positions, proof.TreeVosu)
	leaves_key := f.VerifyMultiBranch(proof.TreeRoots[1], augmented_positions, proof.TreeKey)
	leaves_d := f.VerifyMultiBranch(proof.TreeRoots[2], augmented_positions, proof.TreeD)
	leaves_b := f.VerifyMultiBranch(proof.TreeRoots[3], augmented_positions, proof.TreeB)
	leaves_l := f.VerifyMultiBranch(proof.TreeRoots[4], positions, proof.TreeL)

	// log.Printf("leaves %v, %v, %v, %v", len(leaves_vosu), len(leaves_d), len(leaves_b), len(leaves_l))

	for i := 0; i < len(positions); i++ {
		power := uint256.NewInt(uint64(positions[i]))
		x := f.GFP.Exp(G2, power)
		x_to_steps := f.GFP.Exp(x, uint256.NewInt(uint64(f.steps)))

		vi_x := eval_visu[positions[i]]
		vo_x := f.GFP.IntFromBytes(leaves_vosu[i*2])
		vo_xpre := f.GFP.IntFromBytes(leaves_vosu[i*2+1])
		k_x := f.GFP.IntFromBytes(leaves_key[i*2])
		d_x := f.GFP.IntFromBytes(leaves_d[i*2])
		b_x := f.GFP.IntFromBytes(leaves_b[i*2])
		l_x := f.GFP.IntFromBytes(leaves_l[i])

		z := f.GFP.Exp(x, uint256.NewInt(uint64(f.steps)))
		z = f.GFP.Sub(z, uint256.NewInt(1))
		// k_x := f.GFP.EvalPolyAt(poly_key, f.GFP.Exp(x, uint256.NewInt(uint64(skip2))))

		// log.Printf("vi : %v, vo : %v, vo_pre : %v, kx : %v", vi_x, vo_x, vo_xpre, k_x)

		// C(x) = Vi(x) - (Vo(x)^3 * Vo(x/g1) - K(x))
		// C(x) = Z(x)*D(x)
		c := f.GFP.Exp(vo_x, uint256.NewInt(3))
		if !vo_xpre.IsZero() {
			c = f.GFP.Mul(c, vo_xpre)
		}
		// c = f.GFP.Sub(c, k_x)
		if !k_x.IsZero() {
			c = f.GFP.Mul(c, k_x)
		}
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

func (f *starks) GetSizeStarksProofPreKey(proof *dtype.StarksProof) int {
	size := 0

	size += len(proof.MerkleRoot)

	for i := 0; i < len(proof.TreeRoots); i++ {
		size += len(proof.TreeRoots[i])
	}
	log.Printf("sizof buf : %v", size)

	for i := 0; i < len(proof.TreeVosu); i++ {
		for j := 0; j < len(proof.TreeVosu[i]); j++ {
			size += len(proof.TreeVosu[i][j])
		}
	}

	log.Printf("sizof buf : %v", size)
	for i := 0; i < len(proof.TreeKey); i++ {
		for j := 0; j < len(proof.TreeKey[i]); j++ {
			size += len(proof.TreeKey[i][j])
		}
	}

	log.Printf("sizof buf : %v", size)
	for i := 0; i < len(proof.TreeD); i++ {
		for j := 0; j < len(proof.TreeD[i]); j++ {
			size += len(proof.TreeD[i][j])
		}
	}

	log.Printf("sizof buf : %v", size)
	for i := 0; i < len(proof.TreeB); i++ {
		for j := 0; j < len(proof.TreeB[i]); j++ {
			size += len(proof.TreeB[i][j])
		}
	}

	log.Printf("sizof buf : %v", size)
	for i := 0; i < len(proof.TreeL); i++ {
		for j := 0; j < len(proof.TreeL[i]); j++ {
			size += len(proof.TreeL[i][j])
		}
	}

	log.Printf("sizof buf : %v", size)
	size += len(proof.VosuFl) * 4 * 8

	log.Printf("sizof buf : %v", size)
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

	buf_proof, err := json.Marshal(proof)
	if err != nil {
		log.Printf("json.Marshel error : %v", err)
	}
	log.Printf("sizof buf : %v", len(buf_proof))

	return size
}
