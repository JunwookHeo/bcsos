package dbagent

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/junwookheo/bcsos/blockchainnode/network"
	"github.com/junwookheo/bcsos/common/bitcoin"
	"github.com/junwookheo/bcsos/common/blockchain"
	"github.com/junwookheo/bcsos/common/config"
	"github.com/junwookheo/bcsos/common/dtype"
	"github.com/junwookheo/bcsos/common/poscipher"
	"github.com/junwookheo/bcsos/common/serial"
	"github.com/junwookheo/bcsos/common/starks"
	"github.com/junwookheo/bcsos/common/wallet"
)

type btcBlock struct {
	timestamp int64
	height    int
	hashprev  string
	hash      string // hash of plain data
	hashenc   string // hash of encrypted data
	hashkey   string // hash of key data of the previous encrypted block used when enctypting
	encblock  []byte // encrypted block to encrypt the next block
}

type btcDBStatus struct {
	Timestamp     time.Time
	ID            int
	TotalBlocks   int
	TotalSize     int
	TimeGenProof  int
	TimeVerifyFwd int
	TimeVerifyRev int
	SizeProof     int
	TimePosAcc    int
}

type btcDBProof struct {
	Timestamp    time.Time
	ProofHeight  int
	ProofBlock   string
	TimeGenProof int
	SizeGenProof int
}

type btcDBVerif struct {
	Timestamp        time.Time
	VerifHeight      int
	VerifBlock       string
	TimeVerifFwd     int
	TimeVerifRev     int
	TimeLastBlock    int64
	TimeRcvLastBlock int64
	TimeRcvPoS       int64
	IsSuccess        int
}

type latency struct {
	NumMeasure int64
	TotalDealy int64
}

type btcdbagent struct {
	db        *sql.DB
	sclass    int
	dbstatus  btcDBStatus
	dirpath   string
	lastblock btcBlock
	avgdelay  latency
	mutex     sync.Mutex
}

const SIZE_PROOF = 8 + 32 + 32*config.NUM_CONSECUTIVE_HASHES + 32*config.NUM_CONSECUTIVE_HASHES
const SIZE_STARKS = 65536 / 8 / 4
const (
	V_SUCCESS         int = 0
	V_FAIL_TIME           = 1
	V_FAIL_VERIFY         = 1 << 1
	V_FAIL_VERIFY_FWD     = 1 << 2
	V_FAIL_VERIFY_BAC     = 1 << 3
)

func (a *btcdbagent) Close() {
	a.db.Close()
}

func (a *btcdbagent) GetLatestBlockHash() (string, int) {
	log.Panicln("GetLatestBlockHash")
	return "", -1
}

func (a *btcdbagent) RemoveObject(hash string) bool {
	log.Panicln("RemoveObject")
	return false
}

func (a *btcdbagent) AddBlockHeader(hash string, h *blockchain.BlockHeader) int64 {
	log.Panicln("AddBlockHeader")
	return -1
}

func (a *btcdbagent) GetBlockHeader(hash string, h *blockchain.BlockHeader) int64 {
	log.Panicln("GetBlockHeader")
	return -1
}

func (a *btcdbagent) AddTransaction(t *blockchain.Transaction) int64 {
	log.Panicln("AddTransaction")
	return -1
}

func (a *btcdbagent) GetTransaction(hash string, t *blockchain.Transaction) int64 {
	log.Panicln("GetTransaction")
	return -1
}

func (a *btcdbagent) addBtcBlocktoList(b *btcBlock) int64 {
	// if id := a.GetObject(obj); id != 0 {
	// 	// log.Printf("Replicatoin exists : %v - %v", id, obj)
	// 	return id
	// }

	id := func() int64 {
		a.mutex.Lock()
		defer a.mutex.Unlock()
		st, err := a.db.Prepare("INSERT INTO btcblocklist (timestamp, height, hashprev, hash, hashenc, hashkey) VALUES (?, ?, ?, ?, ?, ?)")
		if err != nil {
			log.Printf("Prepare update BTC Block object error : %v", err)

			return -1
		}
		defer st.Close()

		rst, err := st.Exec(b.timestamp, b.height, b.hashprev, b.hash, b.hashenc, b.hashkey)
		if err != nil {
			log.Panicf("Exec adding object error : %v", err)
			return -1
		}

		id, _ := rst.LastInsertId()
		return id
	}()

	// Update db status after adding object
	// a.updateAddDBStatus(id)
	return id
}

func (a *btcdbagent) getEncryptKeyforGenesis() []byte {
	w := wallet.NewWallet(config.WALLET_PATH)
	key := sha256.Sum256(w.GetAddress()[:])
	return key[:]
	// return w.PublicKey //GetAddress()
}

func (a *btcdbagent) encryptPoSWithVariableLength(key, s []byte) (string, []byte) {
	// return poscipher.EncryptPoSWithVariableLength(key, s)
	return poscipher.EncryptPoSWithPrimeFieldPreKey(key, s)
}

func (a *btcdbagent) decryptPoSWithVariableLength(key, s []byte) []byte {
	// return poscipher.DecryptPoSWithVariableLength(key, s)
	return poscipher.DecryptPoSWithPrimeFieldPreKey(key, s)
}

func (a *btcdbagent) getHashforPoSKey(key []byte, ls int) string {
	return poscipher.GetHashforPoSKey(key, ls)
}

func (a *btcdbagent) AddNewBlock(ib interface{}) int64 {
	sb, ok := ib.(*bitcoin.BlockPkt)
	if !ok {
		log.Panicf("Type mismatch : %v", ok)
		return -1
	}

	block := bitcoin.NewBlock()
	rb := bitcoin.NewRawBlock(sb.Block)
	hashprev := ""
	if a.lastblock.height > -1 { // if it is not the first block
		_ = rb.ReadUint32()
		hbuf := rb.ReverseBuf(rb.ReadBytes(32))
		hashprev = hex.EncodeToString(hbuf)
	}

	block.SetHash(rb.GetRawBytes(0, 80))
	hash := block.GetHashString()
	s := rb.GetBlockBytes(SIZE_STARKS * 31)
	size := len(s)
	addr := a.getEncryptKeyforGenesis()

	// Enctypting a new block
	key := a.lastblock.encblock
	hashenc, encblock := a.encryptPoSWithVariableLength(key, poscipher.CalculateXorWithAddress(addr, s))
	a.lastblock.timestamp = sb.Timestamp
	a.lastblock.hash = hash
	a.lastblock.height += 1
	a.lastblock.hashprev = hashprev
	a.lastblock.hashenc = hashenc
	a.lastblock.encblock = encblock
	a.lastblock.hashkey = a.getHashforPoSKey(key, size)

	// Add a new btc block to list in db
	a.addBtcBlocktoList(&a.lastblock)

	err := ioutil.WriteFile(filepath.Join(a.dirpath, hash), encblock, 0777)
	if err != nil {
		log.Panicf("Wrinting block err : %v", err)
		return -1
	}

	{
		a.mutex.Lock()
		defer a.mutex.Unlock()
		status := &a.dbstatus
		status.TotalBlocks += 1
		status.TotalSize += size
	}

	return int64(a.lastblock.height)
}

func (a *btcdbagent) getEncryptInfoWithHeight(height int) *btcBlock {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	rows, err := a.db.Query(`SELECT *  FROM btcblocklist WHERE height=?;`, height)
	if err != nil {
		log.Printf("Show latest db status Error : %v", err)
		return nil
	}

	defer rows.Close()

	for rows.Next() {
		id := 0
		bi := btcBlock{}
		rows.Scan(&id, &bi.timestamp, &bi.height, &bi.hashprev, &bi.hash, &bi.hashenc, &bi.hashkey)

		return &bi
	}

	return nil
}

func (a *btcdbagent) getEncryptInfoWithPreviousHash(hash string) *btcBlock {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	rows, err := a.db.Query(`SELECT *  FROM btcblocklist WHERE hashprev=?;`, hash)
	if err != nil {
		log.Printf("Show latest db status Error : %v", err)
		return nil
	}

	defer rows.Close()

	for rows.Next() {
		id := 0
		bi := btcBlock{}
		rows.Scan(&id, &bi.timestamp, &bi.height, &bi.hashprev, &bi.hash, &bi.hashenc, &bi.hashkey)

		return &bi
	}

	return nil
}

func (a *btcdbagent) getEncryptInfoWithHash(hash string) *btcBlock {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	rows, err := a.db.Query(`SELECT *  FROM btcblocklist WHERE hash=?;`, hash)
	if err != nil {
		log.Printf("Show latest db status Error : %v", err)
		return nil
	}

	defer rows.Close()

	for rows.Next() {
		id := 0
		bi := btcBlock{}
		rows.Scan(&id, &bi.timestamp, &bi.height, &bi.hashprev, &bi.hash, &bi.hashenc, &bi.hashkey)

		return &bi
	}

	return nil
}

// func hashToUint32(b []byte) uint32 {
// 	x := uint32(0)
// 	for i := 0; i < len(b); i += 4 {
// 		x += uint32(b[i+3]) | uint32(b[i+2])<<8 | uint32(b[i+1])<<16 | uint32(b[i])<<24
// 	}
// 	return x
// }

// height : the start height of n-consecutive encrypted blocks
func (a *btcdbagent) generateProof(height int) *dtype.PoSProof {
	// List up the encrypted block hash and key hash of encrypted blocks
	// And calculate Merkle Root

	start := time.Now().UnixNano()

	var hashroots [][]byte
	proof := dtype.PoSProof{}
	bis := make([]*btcBlock, config.NUM_CONSECUTIVE_HASHES)
	for i := 0; i < config.NUM_CONSECUTIVE_HASHES; i++ {
		bi := a.getEncryptInfoWithHeight(height + i)
		bis[i] = bi
		he, _ := hex.DecodeString(bi.hashenc)
		proof.HashEncs = append(proof.HashEncs, he)
		hk, _ := hex.DecodeString(bi.hashkey)
		proof.HashKeys = append(proof.HashKeys, hk)
		mh := blockchain.CalMerkleNodeHash(he, hk)
		hashroots = append(hashroots, mh)
	}

	// select a block using the merkle root
	proof.Root = blockchain.CalMerkleRootHash(hashroots)
	// randomize block selection. exclude the last block for forward verification
	proof.Selected = int(poscipher.HashToUint32(proof.Root) % uint32(config.NUM_CONSECUTIVE_HASHES-1))
	proof.Hash = bis[proof.Selected].hash
	log.Printf("Proof selected : %v - %v", proof.Selected, proof.Hash)

	// create proof, merkle root, {hash, key hash}, encrypted block, timestamp
	eb, err := ioutil.ReadFile(filepath.Join(a.dirpath, proof.Hash))
	if err != nil {
		log.Panicf("Reading encryped block err : %v", err)
		return nil
	}
	proof.EncBlock = eb
	proof.Timestamp = time.Now().UnixNano()
	proof.Address = a.getEncryptKeyforGenesis()
	log.Printf("Proof : %v - %v", proof.Timestamp, proof.Hash)

	gap := int(time.Now().UnixNano() - start)
	{
		a.mutex.Lock()
		status := &a.dbstatus
		status.TimeGenProof += gap
		status.SizeProof += len(eb) + SIZE_PROOF
		a.mutex.Unlock()
	}
	dbproof := btcDBProof{}
	dbproof.ProofHeight = bis[proof.Selected].height
	dbproof.ProofBlock = bis[proof.Selected].hash
	dbproof.SizeGenProof = len(eb) + SIZE_PROOF
	dbproof.TimeGenProof = gap

	a.updateDBProof(&dbproof)
	log.Printf("Proof stats : %v", a.dbstatus)
	return &proof
}

// height : the start height of n-consecutive encrypted blocks
func (a *btcdbagent) generateStarksProof(height int, hash string) *dtype.NonInteractiveProof {
	start := time.Now().UnixNano()

	bi := a.getEncryptInfoWithHeight(height)
	if bi == nil {
		log.Printf("Get Encrypt block info error : %v", height)
		return nil
	}

	cb, err := ioutil.ReadFile(filepath.Join(a.dirpath, bi.hash))
	if err != nil {
		log.Panicf("Reading encryped block err : %v", err)
		return nil
	}

	key, err := ioutil.ReadFile(filepath.Join(a.dirpath, bi.hashprev))
	if err != nil {
		log.Panicf("Reading encryped block key err : %v", err)
		return nil
	}

	vis := a.decryptPoSWithVariableLength(key, cb)

	f := starks.NewStarks(SIZE_STARKS)
	starks_proof := f.GenerateStarksProofPreKey(hash, vis, cb, key)
	proof_size := f.GetSizeStarksProofPreKey(starks_proof)
	gap := int(time.Now().UnixNano() - start)
	{
		a.mutex.Lock()
		status := &a.dbstatus
		status.TimeGenProof += gap
		status.SizeProof += proof_size
		a.mutex.Unlock()
	}
	dbproof := btcDBProof{}
	dbproof.ProofHeight = bi.height
	dbproof.ProofBlock = bi.hash
	dbproof.SizeGenProof = proof_size
	dbproof.TimeGenProof = gap

	a.updateDBProof(&dbproof)
	log.Printf("Proof stats : %v", a.dbstatus)
	var proof dtype.NonInteractiveProof
	// proof := make([]interface{}, 3)
	proof.Address = a.getEncryptKeyforGenesis()
	proof.Hash = bi.hash
	proof.Starks = starks_proof

	return &proof
}

// hash : new block's hash for randomization
func (a *btcdbagent) GetRandomHeightForNConsecutiveBlocks(hash string) int {
	// Perform PoS when block height is larger than NUM_CONSECUTIVE_HASHES*3
	if a.lastblock.height <= config.NUM_CONSECUTIVE_HASHES+1 {
		return -1
	}
	// Select block
	tmp, _ := hex.DecodeString(hash)
	hb := sha256.Sum256(tmp)
	ri := poscipher.HashToUint32(hb[:]) % uint32(a.lastblock.height-(config.NUM_CONSECUTIVE_HASHES+1)) // margin 10 blocks
	ri += 1                                                                                            // Exclude genesys block
	log.Printf("Block selector : %v", ri)

	return int(ri)
}

// hash : new block's hash for randomization
func (a *btcdbagent) getRandomHeightForProofBlocks(hash string) int {
	// Perform PoS when block height is larger than 6
	LEASTBLOCKS := 2
	if a.lastblock.height <= LEASTBLOCKS {
		log.Printf("Block selector Max Height short: %v", a.lastblock.height)
		return -1
	}
	// Select block
	ri := poscipher.GetRandIntFromHash(hash) % int(a.lastblock.height-1) // margin 10 blocks

	ri += 1 // Exclude genesys block
	log.Printf("Block selector : %v - max : %v", ri, a.lastblock.height)

	return int(ri)
}

// hash : new block's hash for randomization
func (a *btcdbagent) GetNonInteractiveStarksProof(hash string) *dtype.NonInteractiveProof {
	// Perform PoS when block height is larger than NUM_CONSECUTIVE_HASHES*3
	ri := a.getRandomHeightForProofBlocks(hash)
	if ri == -1 {
		return nil
	}

	return a.generateStarksProof(int(ri), hash)
}

func (a *btcdbagent) GetInteractiveProof(height int) *dtype.PoSProof {
	return a.generateProof(int(height))
}

func (a *btcdbagent) getDecryptBlock(bi *btcBlock) []byte {
	eb, err := ioutil.ReadFile(filepath.Join(a.dirpath, bi.hash))
	if err != nil {
		log.Panicf("Reading encryped block err : %v", err)
		return nil
	}

	key, err := ioutil.ReadFile(filepath.Join(a.dirpath, bi.hashprev))
	if err != nil {
		log.Panicf("Reading encryped block key err : %v", err)
		return nil
	}

	addr := a.getEncryptKeyforGenesis()
	return poscipher.CalculateXorWithAddress(addr, a.decryptPoSWithVariableLength(key, eb))
}

func (a *btcdbagent) verifyInterActiveProofStorage_Fwd(proof *dtype.PoSProof) bool {
	bi := a.getEncryptInfoWithPreviousHash(proof.Hash)
	if bi == nil {
		log.Printf("Get Encrypt block info error : %v", proof.Hash)
		return false
	}
	// log.Printf("Get Block info for verification : %v", bi)

	// Forward verification : Get original block by decrypting bk and bk+1
	b := a.getDecryptBlock(bi)
	if b == nil {
		log.Printf("Get Decrypt block error : %v", bi.hash)
		return false
	}

	// block := bitcoin.NewBlock()
	// rb := bitcoin.NewRawBlock(hex.EncodeToString(b))
	// _ = rb.ReadUint32()
	// hbuf := rb.ReverseBuf(rb.ReadBytes(32))
	// hashprev := hex.EncodeToString(hbuf)

	// block.SetHash(rb.GetRawBytes(0, 80))
	// hash := block.GetHashString()
	// s := rb.GetBlockBytes()
	// size := len(s)
	// log.Printf("FWD : %v, %v, %v", hashprev, hash, size)

	// Get Key block(Previous encrypt block) from bk and ebk
	addr := proof.Address
	hashkey, _ := a.encryptPoSWithVariableLength(proof.EncBlock, poscipher.CalculateXorWithAddress(addr, b))
	if hashkey == hex.EncodeToString(proof.HashEncs[proof.Selected+1]) {
		log.Printf("Verifying PoS Success : %v", hashkey)
		return true
	}

	log.Printf("Verifying PoS Fail : %v", hashkey)
	log.Printf("Thash : %v", hex.EncodeToString(proof.HashKeys[proof.Selected]))
	log.Printf("Thash : %v", hex.EncodeToString(proof.HashKeys[proof.Selected+1]))
	return false
}

func (a *btcdbagent) verifyInterActiveProofStorage_Rev(proof *dtype.PoSProof) bool {
	bi := a.getEncryptInfoWithHash(proof.Hash)
	if bi == nil {
		log.Printf("Get Encrypt block info error : %v", proof.Hash)
		return false
	}
	// log.Printf("Get Block info for verification : %v", bi)

	// Forward verification : Get original block by decrypting bk and bk-1
	b := a.getDecryptBlock(bi)
	if b == nil {
		log.Printf("Get Decrypt block error : %v", bi.hash)
		return false
	}

	// block := bitcoin.NewBlock()
	// rb := bitcoin.NewRawBlock(hex.EncodeToString(b))
	// _ = rb.ReadUint32()
	// hbuf := rb.ReverseBuf(rb.ReadBytes(32))
	// hashprev := hex.EncodeToString(hbuf)

	// block.SetHash(rb.GetRawBytes(0, 80))
	// hash := block.GetHashString()
	// s := rb.GetBlockBytes()
	// size := len(s)
	// log.Printf("REV : %v, %v, %v", hashprev, hash, size)

	// Get Key block(Previous block) from bk and ebk
	addr := proof.Address
	peb := a.decryptPoSWithVariableLength(poscipher.CalculateXorWithAddress(addr, b), proof.EncBlock)
	hashkey := poscipher.GetHashString(peb)
	// log.Printf("PEB : %x", peb[0:80])

	if hashkey == hex.EncodeToString(proof.HashKeys[proof.Selected]) {
		log.Printf("Verifying PoS Success : %v", hashkey)
		return true
	}

	log.Printf("Verifying PoS Fail : %v", hashkey)
	log.Printf("Thash : %v", hex.EncodeToString(proof.HashKeys[proof.Selected]))
	return false
}

func (a *btcdbagent) VerifyInterActiveProofStorage(proof *dtype.PoSProof) bool {
	bi := a.getEncryptInfoWithHash(proof.Hash)
	dbverif := btcDBVerif{}
	dbverif.VerifBlock = bi.hash
	dbverif.VerifHeight = bi.height
	dbverif.IsSuccess = V_SUCCESS

	start := time.Now().UnixNano()
	ret1 := a.verifyInterActiveProofStorage_Rev(proof)
	if !ret1 {
		dbverif.IsSuccess |= V_FAIL_VERIFY_BAC
	}
	gap := int(time.Now().UnixNano() - start)
	{
		a.mutex.Lock()
		status := &a.dbstatus
		status.TimeVerifyRev += gap
		a.mutex.Unlock()
	}

	dbverif.TimeRcvPoS = start
	dbverif.TimeVerifRev = gap
	start = time.Now().UnixNano()
	ret2 := a.verifyInterActiveProofStorage_Fwd(proof)
	if !ret2 {
		dbverif.IsSuccess |= V_FAIL_VERIFY_FWD
	}

	gap = int(time.Now().UnixNano() - start)
	{
		a.mutex.Lock()
		status := &a.dbstatus
		status.TimeVerifyFwd += gap
		a.mutex.Unlock()
	}

	dbverif.TimeVerifFwd = gap
	a.updateDBVerif(&dbverif)
	return ret1 && ret2
}

func (a *btcdbagent) VerifyNonInterActiveProofStorage(tlb int64, trb int64, trp int64, proof *dtype.NonInteractiveProof) bool {
	addr := proof.Address
	hash := proof.Hash
	starks_proof := proof.Starks

	start := time.Now().UnixNano()
	bi := a.getEncryptInfoWithHash(hash)
	if bi == nil {
		log.Printf("Get Encrypt block info error : %v", hash)
		return false
	}

	dbverif := btcDBVerif{}
	dbverif.VerifBlock = bi.hash
	dbverif.VerifHeight = bi.height
	dbverif.TimeLastBlock = tlb
	dbverif.TimeRcvLastBlock = trb
	dbverif.TimeRcvPoS = trp
	dbverif.IsSuccess = V_SUCCESS

	diff := int64(0)
	weight := int64(0)
	if trb > tlb {
		diff = trb - tlb
		a.avgdelay.TotalDealy += diff
		a.avgdelay.NumMeasure += 1
		weight = (a.avgdelay.TotalDealy / a.avgdelay.NumMeasure)
	}

	if (trp > trb) && (trp-trb-weight)/1000000 > int64(config.MAX_PROOF_TIME_MSEC) {
		log.Printf("Verify Proof : Time Exceed %v", (trp-trb)/1000000)
		dbverif.IsSuccess |= V_FAIL_TIME
	}

	b := a.getDecryptBlock(bi)
	if b == nil {
		log.Printf("Get Decrypt block error : %v", bi.hash)
		dbverif.IsSuccess |= V_FAIL_VERIFY
		return false
	}
	vis := poscipher.CalculateXorWithAddress(addr, b)

	f := starks.NewStarks(65536 / 8 / 4)
	ret := f.VerifyStarksProofPreKey(vis, starks_proof)
	if !ret {
		dbverif.IsSuccess |= V_FAIL_VERIFY
	}

	gap := int(time.Now().UnixNano() - start)
	{
		a.mutex.Lock()
		status := &a.dbstatus
		status.TimeVerifyFwd += gap
		a.mutex.Unlock()
	}

	dbverif.TimeVerifFwd = gap
	a.updateDBVerif(&dbverif)

	return ret
}

func (a *btcdbagent) initLastBlock() {
	a.lastblock.timestamp = time.Now().UnixNano()
	a.lastblock.encblock = a.getEncryptKeyforGenesis() // encblock is data to encrypt the next block
	a.lastblock.hash = poscipher.GetHashString(a.lastblock.encblock)
	a.lastblock.hashenc = a.lastblock.hash // This block does not need to be encrypted
	a.lastblock.hashkey = ""               // There is no key
	a.lastblock.height = -1                // This is key for the first block(B0)
	a.lastblock.hashprev = ""              // No previous block

	// TODO : Add to it to db
	a.addBtcBlocktoList(&a.lastblock)
}

func (a *btcdbagent) GetLastBlockTime() int64 {
	return a.lastblock.timestamp
}

func (a *btcdbagent) GetBlock(hash string, b *blockchain.Block) int64 {
	log.Panicln("GetBlock")
	return -1
}

func (a *btcdbagent) ShowAllObjets() bool {
	log.Panicln("ShowAllObjets")
	return false
}

func (a *btcdbagent) GetDBDataSize() uint64 {
	log.Panicln("GetDBDataSize")
	return 0
}

func (a *btcdbagent) GetDBStatus() *DBStatus {
	log.Panicln("GetDBStatus")
	return nil
}

func (a *btcdbagent) GetTransactionwithUniform(num int, hashes *[]RemoverbleObj) bool {
	log.Panicln("GetTransactionwithUniform")
	return false
}

func (a *btcdbagent) GetTransactionwithExponential(num int, hashes *[]RemoverbleObj) bool {
	log.Panicln("GetTransactionwithExponential")
	return false
}

func (a *btcdbagent) DeleteNoAccedObjects() {
	// No Need this function for PoS
}

func (a *btcdbagent) UpdateDBNetworkQuery(fromqc int, toqc int, totalqc int) {
	log.Panicln("UpdateDBNetworkQuery")
}

func (a *btcdbagent) UpdateDBNetworkDelay(addtime int, hop int) {
	log.Panicln("UpdateDBNetworkDelay")
}

func (a *btcdbagent) getLatestDBStatus(status *btcDBStatus) bool {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	rows, err := a.db.Query(`SELECT *  FROM dbstatus WHERE id = (SELECT MAX(id)  FROM dbstatus);`)
	if err != nil {
		log.Printf("Show latest db status Error : %v", err)
		return false
	}

	defer rows.Close()

	for rows.Next() {
		rows.Scan(&status.ID, &status.Timestamp, &status.TotalBlocks, &status.TotalSize, &status.TimeGenProof, &status.TimeVerifyFwd, &status.TimeVerifyRev, &status.SizeProof, &status.TimePosAcc)

		return true
	}

	return false
}

func (a *btcdbagent) updateDBStatus() {
	getHash := func(status btcDBStatus) string {
		status.ID = 0
		status.Timestamp = time.Time{}
		byte_status := sha256.Sum256(serial.Serialize(status))
		return hex.EncodeToString(byte_status[:])
	}

	last_hash := getHash(a.dbstatus)

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		<-ticker.C
		func() {
			a.mutex.Lock()
			defer a.mutex.Unlock()
			status := &a.dbstatus
			hash_status := getHash(*status)
			if last_hash == hash_status {
				return
			}

			last_hash = hash_status

			st, err := a.db.Prepare(`INSERT INTO dbstatus (timestamp, totalblocks, totalsize, timegenproof, timeverifyfwd, timeverifyrev, sizeproof, timeposacc) 
					VALUES ( datetime('now'), ?, ?, ?, ?, ?, ?, ?)`)
			if err != nil {
				log.Printf("Prepare adding dbstatus error : %v", err)
				return
			}
			defer st.Close()

			rst, err := st.Exec(status.TotalBlocks, status.TotalSize, status.TimeGenProof, status.TimeVerifyFwd, status.TimeVerifyRev, status.SizeProof, status.TimePosAcc)
			if err != nil {
				log.Panicf("Exec adding dbstatus error : %v", err)
				return
			}

			id, _ := rst.LastInsertId()
			status.ID = int(id)
			// log.Printf("Update dbstatus : %v", *status)
		}()
	}
}

func (a *btcdbagent) updateDBProof(dbproof *btcDBProof) {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	st, err := a.db.Prepare(`INSERT INTO prooftbl (timestamp, proofheight, proofblock, timegenproof, sizegenproof) VALUES ( datetime('now'), ?, ?, ?, ?)`)
	if err != nil {
		log.Printf("Prepare adding prooftbl error : %v", err)
		return
	}
	defer st.Close()

	_, err = st.Exec(dbproof.ProofHeight, dbproof.ProofBlock, dbproof.TimeGenProof, dbproof.SizeGenProof)
	if err != nil {
		log.Panicf("Exec adding prooftbl error : %v", err)
		return
	}
	log.Printf("Exec adding prooftbl  : %v, %v, %v, %v", dbproof.ProofHeight, dbproof.ProofBlock, dbproof.TimeGenProof, dbproof.SizeGenProof)
}

func (a *btcdbagent) updateDBVerif(dbverif *btcDBVerif) {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	st, err := a.db.Prepare(`INSERT INTO veriftbl (timestamp, verifheight, verifblock, timeveriffwd, timeverifrev, timelastblock, timercvlastblock, timercvpos, issuccess) VALUES ( datetime('now'), ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		log.Printf("Prepare adding veriftbl error : %v", err)
		return
	}
	defer st.Close()

	_, err = st.Exec(dbverif.VerifHeight, dbverif.VerifBlock, dbverif.TimeVerifFwd, dbverif.TimeVerifRev, dbverif.TimeLastBlock, dbverif.TimeRcvLastBlock, dbverif.TimeRcvPoS, dbverif.IsSuccess)
	if err != nil {
		log.Panicf("Exec adding veriftbl error : %v", err)
		return
	}
	log.Printf("Exec adding veriftbl  : %v", dbverif)
}

func newDBBtcSqlite(path string) DBAgent {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		log.Panicf("Open sqlite db error : %v", err)
	}

	create_objtbl := `CREATE TABLE IF NOT EXISTS btcblocklist (
		id      	INTEGER  PRIMARY KEY AUTOINCREMENT,
		timestamp	INTEGER,
		height		INTEGER,
		hashprev   	TEXT,
		hash    	TEXT,
		hashenc		TEXT,
		hashkey		TEXT
	);`

	st, err := db.Prepare(create_objtbl)
	if err != nil {
		log.Panicf("create_objtlb error %v", err)
	}
	defer st.Close()

	st.Exec()

	// totalquery : query objects including local storage
	// queryfrom : the number of received queries to get deleted transactions
	// queryto : the number of send queries to get deleted transactions
	create_statustbl := `CREATE TABLE IF NOT EXISTS dbstatus (
		id      			INTEGER  PRIMARY KEY AUTOINCREMENT,
		timestamp			DATETIME,
		totalblocks			INTEGER,
		totalsize			INTEGER,
		timegenproof		INTEGER,
		timeverifyfwd		INTEGER,
		timeverifyrev		INTEGER,
		sizeproof			INTEGER,
		timeposacc			INTEGER
	);`

	st, err = db.Prepare(create_statustbl)
	if err != nil {
		log.Panicf("create_statustlb error %v", err)
	}
	defer st.Close()

	st.Exec()

	create_prooftbl := `CREATE TABLE IF NOT EXISTS prooftbl (
		id      			INTEGER  PRIMARY KEY AUTOINCREMENT,
		timestamp			DATETIME,
		proofheight			INTEGER,
		proofblock			TEXT,
		timegenproof		INTEGER,
		sizegenproof		INTEGER
	);`

	st, err = db.Prepare(create_prooftbl)
	if err != nil {
		log.Panicf("create_statustlb error %v", err)
	}
	defer st.Close()

	st.Exec()

	create_veriftbl := `CREATE TABLE IF NOT EXISTS veriftbl (
		id      			INTEGER  PRIMARY KEY AUTOINCREMENT,
		timestamp			DATETIME,
		verifheight			INTEGER,
		verifblock			TEXT,
		timeveriffwd		INTEGER,
		timeverifrev		INTEGER,
		timelastblock		INTEGER,
		timercvlastblock	INTEGER,
		timercvpos			INTEGER,
		issuccess			INTEGER
	);`

	st, err = db.Prepare(create_veriftbl)
	if err != nil {
		log.Panicf("create_statustlb error %v", err)
	}
	defer st.Close()

	st.Exec()

	ni := network.NodeInfoInst()
	local := ni.GetLocalddr()

	dba := btcdbagent{db: db, sclass: local.SC, dbstatus: btcDBStatus{Timestamp: time.Now()}, dirpath: "", lastblock: btcBlock{}, avgdelay: latency{0, 0}, mutex: sync.Mutex{}}
	dba.getLatestDBStatus(&dba.dbstatus)
	go dba.updateDBStatus()

	dba.initLastBlock()

	dba.dirpath = path + ".blocks"
	err = os.RemoveAll(dba.dirpath)
	if err != nil {
		log.Panicf("Error Remove Dir : %v", err)
	}

	err = os.MkdirAll(dba.dirpath, 0777)
	if err != nil {
		log.Panicf("Erro Make Dir %v: %v", path, err)
	}

	return &dba
}
