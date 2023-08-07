import os
import time
import json
import hashlib
from datetime import datetime

from pycrypto.zokrates_pycrypto.eddsa import PrivateKey, PublicKey
from pycrypto.zokrates_pycrypto.field import FQ
from pycrypto.zokrates_pycrypto.utils import write_signature_for_zokrates_cli

P = 21888242871839275222246405745257275088548364400416034343698204186575808495617
SEED_KEY = 1997011358982923168928344992199991480689546837621580239342656433234255379025
FZOK = "redact.zok"
# FBTC = "./blocks.json"
FBTC = "../blocks_2023_10.json"

MLEVEL = {"C":2, "B":2, "T":1}

NUM_CONJ = 10

def RBO_MSG(t, l, *args, **kwargs):
    h = MLEVEL.get(t)
    if  h != None and l <= h:
        dt = datetime.now()
        print(dt, *args, **kwargs)

class redact:
    def __init__(self, seed=SEED_KEY):
        self.seed = seed
        self.msk, self.mpk = self.getmasterkeys()
        
    def getmasterkeys(self):
        key = FQ(self.seed)
        sk = PrivateKey(key)
        pk = PublicKey.from_private(sk)
        return sk, pk

    def getsignkey(self, msg):
        mh = hashlib.sha256(msg.encode("utf-8")).digest()
        p = mh + self.msk.fe.n.to_bytes(32, 'big')
        digest = hashlib.sha256(p).digest()
        key = FQ(int(digest.hex(), 16))
        sk = PrivateKey(key)
        pk = PublicKey.from_private(sk)
        return sk, pk

    def sign(self, msg):
        sk, pk = self.getsignkey(msg)
        mh = hashlib.sha512(msg.encode("utf-8")).digest()
        sig = sk.sign(mh)        
        return sig, mh, sk, pk

    def signwithkey(self, msg, sk):
        mh = hashlib.sha512(msg.encode("utf-8")).digest()
        sig = sk.sign(mh)        
        return sig, mh

    def verify(self, pk, mh, sig):
        return pk.verify(sig, mh)

    def gethashpk(self, pk):
        a = pk.p.x.n.to_bytes(32, 'big') + pk.p.y.n.to_bytes(32, 'big')
        h = int(hashlib.sha256(a).hexdigest(), 16) % P
        RBO_MSG("C", 3, "H1 : %d, 0x%x"%(h, h))
        return h
    
    def gethashmsg(self, msg):
        mh = hashlib.sha512(msg.encode("utf-8")).digest()
        M0 = mh.hex()[:64]
        M1 = mh.hex()[64:]
        b0 = [int(M0[i:i+8], 16) for i in range(0,len(M0), 8)]
        b0.extend([int(M1[i:i+8], 16) for i in range(0,len(M1), 8)])
        b = b""
        b += b"".join(m.to_bytes(4, "big") for m in b0)
        h = int(hashlib.sha256(b).hexdigest(), 16) % P
        RBO_MSG("C", 3, "H2 : %d, 0x%x" %(h, h))
        return h

    def gethashtransaction(self, hpk, tid, hmsg):
        t = hpk.to_bytes(32, 'big') + tid.to_bytes(32, 'big') + hmsg.to_bytes(32, 'big')        
        h = int(hashlib.sha256(t).hexdigest(), 16) % P
        RBO_MSG("C", 3, "TID : %d, 0x%x" %(h, h))
        return h


class btcpaser:
    def swapbytes(self, s):
        d = [s[i:i+2] for i in range(0, len(s), 2)]
        return "".join(reversed(d))

    def readvariant(self, s):
        t = s[0:2]
        if t == 'fd':
            return 6, int(self.swapbytes(s[2:6]), 16)
        elif t == 'fe':
            return 10, int(self.swapbytes(s[2:10]), 16)
        elif t == 'ff':
            return 18, int(self.swapbytes(s[2:18]), 16)
        else:
            return 2, int(t, 16)
    
class btcblock(btcpaser):
    def __init__(self, raw_block):
        self.raw_block = raw_block
        self.trpositions = []
        self.parse(raw_block)
        
    def parse(self, raw_block):
        pos = 0
        RBO_MSG("B", 4, "RAW Block : ", raw_block[:200])
        self.version = self.swapbytes(raw_block[pos:pos+8])
        pos += 8
        RBO_MSG("B", 3, "Version : ", self.version)
        self.prvhash = self.swapbytes(raw_block[pos:pos+64])
        pos += 64
        RBO_MSG("B", 2, "Previous Hash : ", self.prvhash)
        self.merkleroot = self.swapbytes(raw_block[pos:pos+64])
        pos += 64
        RBO_MSG("B", 3, "Merkle Root : ", self.merkleroot)
        self.timestamp = self.swapbytes(raw_block[pos:pos+8])
        pos += 8
        RBO_MSG("B", 3, "Timestamp : ", int(self.timestamp, 16))
        self.difficulty = self.swapbytes(raw_block[pos:pos+8])
        pos += 8
        RBO_MSG("B", 3, "Difficulty : ", int(self.difficulty, 16))
        self.nonce = self.swapbytes(raw_block[pos:pos+8])
        pos += 8
        RBO_MSG("B", 3, "Nonce : ", int(self.nonce, 16))
        n, self.ntx = self.readvariant(raw_block[pos:pos+20])
        pos += n
        RBO_MSG("B", 2, "Num tx : ", self.ntx)
        
        for i in range(self.ntx):
            startp = pos
            RBO_MSG("T", 3, "====== Transaction : ", i, raw_block[pos:pos+100])
            version = self.swapbytes(raw_block[pos:pos+8])
            pos += 8
            RBO_MSG("T", 3, "Version : ", int(version, 16))
            opt = raw_block[pos:pos+4]
            witness =  False
            if opt == '0001':
                pos += 4
                witness =  True
            
            RBO_MSG("T", 3, "witness : ", witness)
            
            n, nin = self.readvariant(raw_block[pos:pos+20])
            pos += n
            RBO_MSG("T", 2, "Num Input : ", nin)

            for j in range(nin):
                RBO_MSG("T", 3, "====== Input : ", j)
                preout = raw_block[pos:pos+64]
                pos += 64
                RBO_MSG("T", 3, "previous output : ", preout)

                index = self.swapbytes(raw_block[pos:pos+8])
                pos += 8
                RBO_MSG("T", 3, "Index : ", index)

                n, scrlen = self.readvariant(raw_block[pos:pos+20])
                pos += n
                RBO_MSG("T", 3, "Script Length : ", scrlen)
                
                if scrlen > 0:
                    scrlen *= 2
                    script = raw_block[pos:pos+scrlen]
                    pos += scrlen
                    RBO_MSG("T", 3, "Script : ", script)
                
                sequence = self.swapbytes(raw_block[pos:pos+8])
                pos += 8
                RBO_MSG("T", 3, "Sequence : ", sequence)

            n, nout = self.readvariant(raw_block[pos:pos+20])
            pos += n
            RBO_MSG("T", 2, "Num output : ", nout)
            
            for j in range(nout):
                RBO_MSG("T", 3, "====== Output : ", j)
                txvalue = self.swapbytes(raw_block[pos:pos+16])
                pos += 16
                RBO_MSG("T", 3, "Value : ", txvalue)

                n, scrlen = self.readvariant(raw_block[pos:pos+20])
                pos += n
                RBO_MSG("T", 3, "Script length : ", scrlen)

                if scrlen > 0:
                    scrlen *= 2
                    script = raw_block[pos:pos+scrlen]
                    pos += scrlen
                    RBO_MSG("T", 3, "Lock Script : ", script)

            if witness == True:
                for j in range(nin):
                    n, nwit = self.readvariant(raw_block[pos:pos+20])
                    pos += n
                    RBO_MSG("T", 3, "Num witness : ", nwit)
                    for k in range(nwit):
                        n, witlen = self.readvariant(raw_block[pos:pos+20])
                        pos += n
                        RBO_MSG("T", 3, "Witness length : ", witlen)

                        witlen *= 2
                        witness = self.swapbytes(raw_block[pos:pos+witlen])
                        pos += witlen
                        RBO_MSG("T", 3, "Witness : ", witness)

            locktime = self.swapbytes(raw_block[pos:pos+8])
            pos += 8
            RBO_MSG("T", 3, "Lock time : ", int(locktime, 16))
            endp = pos
            self.trpositions.append((startp, endp)) 
    
    def itertransactions(self):
        for (start, end) in self.trpositions:
            yield self.raw_block[start:end]
                        
        
def gettransaction(path):
    with open(path, 'r') as f:
        # Reading from json file
        lines = f.readlines()
        RBO_MSG("C", 1, "==================================")
        for jb in lines:
            b = json.loads(jb)['data']            
            hight = list(b.keys())[0]            
            raw_block = b[hight]['raw_block']            
            RBO_MSG("C", 1, hight, len(raw_block))
            btc = btcblock(raw_block)
            trs = btc.itertransactions()
            for tr in trs:
                yield tr

if __name__ == "__main__":
    workingpath = os.getcwd()

    if os.path.exists(os.path.join(workingpath, FZOK)) == False:
        RBO_MSG("C", 4, workingpath)
        if os.path.exists(os.path.join(workingpath, "redact", FZOK)):            
            os.chdir(os.path.join(workingpath, 'redact'))         
        else:
            RBO_MSG("C", 1, 'Cannot find "redact.zok"!!!!')
            os._exit(-1)

    zkpath = os.getcwd()

    for f in os.listdir(os.path.join(zkpath, 'out')):
        os.remove(os.path.join(zkpath, 'out', f))

    os.system(f'cp {FZOK} {os.path.join("out", FZOK)}')
    os.chdir('out')
    os.system(f'zokrates compile --debug -i {FZOK}')
    os.system(f'zokrates setup')
    RBO_MSG("C", 4, zkpath)

    trs = gettransaction(os.path.join(workingpath, FBTC))
    
    sim_start = time.time()
    transaction_count = 0
    signing_time = 0
    proof_time = 0
    # verification_time = 0
    verification_time = [0] * NUM_CONJ
    verification_count = [0] * NUM_CONJ
    pretid = 0  # previous transaction id
    sk, pk = None, None

    for i, tr in enumerate(trs):        
        RBO_MSG("C", 2, "Simulation Time", (time.time() - sim_start))

        if i%NUM_CONJ == 0:
            pretid = 0

        red = redact()

        transaction_count += 1
        # Signing transaction
        start_time = time.time()
        if i%NUM_CONJ == 0:
            sig, mh, sk, pk = red.sign(tr)
        else:
            sig, mh = red.signwithkey(tr, sk)

        signing_time += (time.time() - start_time)
        RBO_MSG("C", 2, "Signing", transaction_count, signing_time/transaction_count)

        is_verify = red.verify(pk, mh, sig)
        RBO_MSG("C", 3, is_verify)

        fin = 'msgi.txt'
        pin = os.path.join(zkpath, 'out', fin)
        write_signature_for_zokrates_cli(pk, sig, mh, pin)

        h1 = red.gethashpk(pk)
        h2 = red.gethashmsg(tr)
        
        tid = red.gethashtransaction(h1, pretid, h2)

        with open(pin, "r+") as f:
            content = f.read()
            RBO_MSG("C", 4, content)
            f.seek(0, 0)
            f.write(str(h1) + ' ' + str(h2) + ' ' + content)

        # Generate Proof for zk-SNARKs
        start_time = time.time()
        os.system(f'cat {fin} | xargs zokrates compute-witness -a')
        os.system(f'zokrates generate-proof')
        proof_time += (time.time() - start_time)
        RBO_MSG("C", 2, "Generate Proof ", transaction_count, proof_time/transaction_count)

        tf = f'{tid}.tr'
        tf = os.path.join(zkpath, 'out', tf)
        with open(tf, "w+") as f:
            pf = os.path.join(zkpath, 'out', 'proof.json')
            proof = {}
            with open(pf, "r+") as fp:
                proof = json.load(fp)
            tr_json = json.dumps({'hpk':h1, 'pretid':pretid, 'hmsg':h2, 'msg':tr, 'proof':proof})
            content = f.write(tr_json)

        #update its previous transaction
        if pretid != 0:
            ptf = f'{pretid}.tr'
            ptf = os.path.join(zkpath, 'out', ptf)
            ptr_json = {}
            with open(ptf, "r+") as f:
                ptr_json = json.load(f)
            ptr_json['msg'] = {}
            ptr_json['proof'] = {}
            ptr_json['nexttid'] = tid
            with open(ptf, "w") as f:
                json.dump(ptr_json, f)

        pretid = tid # update current tid to previous tid for the next turn

        start_time = time.time()
        logs = os.popen(f'zokrates verify').readlines()
        verification_time[i%NUM_CONJ] += (time.time() - start_time)
       
        cur = tid
        while(cur):
            # load current transaction
            ptf = f'{cur}.tr'
            ptf = os.path.join(zkpath, 'out', ptf)
            cur_json = {}
            with open(ptf, "r+") as f:
                cur_json = json.load(f)
            prv = cur_json['pretid']

            #load previous transaction
            if prv != 0:
                start_time = time.time()
        
                ptf = f'{prv}.tr'
                ptf = os.path.join(zkpath, 'out', ptf)
                pre_json = {}
                with open(ptf, "r+") as f:
                    pre_json = json.load(f)
                
                assert(pre_json['nexttid'] == cur)
                assert(pre_json['hpk'] == pre_json['hpk'])
                temp_id = red.gethashtransaction(int(pre_json['hpk']),  int(pre_json['pretid']), int(pre_json['hmsg']))
                assert(prv == temp_id)

                verification_time[i%NUM_CONJ] += (time.time() - start_time)

            cur = prv
        
        
        verification_count[i%NUM_CONJ] += 1

        RBO_MSG("C", 2, "Verification ", verification_count[i%NUM_CONJ], verification_time[i%NUM_CONJ]/verification_count[i%NUM_CONJ], logs[-1].strip('\n'))
        if logs[-1].strip() == 'PASSED':
            RBO_MSG("C", 3, "Pass")
        else:
            RBO_MSG("C", 2, "Fail")

        if transaction_count >= NUM_CONJ:
            break

        

    RBO_MSG("C", 1, "========================================================")
    RBO_MSG("C", 1, "Num Transaction ", transaction_count)
    RBO_MSG("C", 1, "Signing Time ", signing_time / transaction_count)
    RBO_MSG("C", 1, "Generate Proof Time ", proof_time / transaction_count)
    for i, (t, c) in enumerate(zip(verification_time, verification_count)):
        RBO_MSG("C", 1, f"{i+1} Verification Time ", t / c)


    