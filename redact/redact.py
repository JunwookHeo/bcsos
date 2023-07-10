import hashlib

from pycrypto.zokrates_pycrypto.eddsa import PrivateKey, PublicKey
from pycrypto.zokrates_pycrypto.field import FQ
from pycrypto.zokrates_pycrypto.utils import write_signature_for_zokrates_cli
import os

P = 21888242871839275222246405745257275088548364400416034343698204186575808495617
SEED_KEY = 1997011358982923168928344992199991480689546837621580239342656433234255379025
FZOK = "redact.zok"

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

    def verify(self, pk, mh, sig):
        return pk.verify(sig, mh)

if __name__ == "__main__":
    path = os.getcwd()
    if os.path.exists(os.path.join(path, FZOK)) == False:
        print(path)
        if os.path.exists(os.path.join(path, "redact", FZOK)):            
            os.chdir(os.path.join(path, 'redact'))         
        else:
            print('Cannot find "redact.zok"!!!!')
            os._exit(-1)

    path = os.getcwd()
    os.system(f'cp {FZOK} {os.path.join("out", FZOK)}')
    os.chdir('out')
    os.system(f'zokrates compile --debug -i {FZOK}')
    os.system(f'zokrates setup')
    print(path)

    for i in range(10):
        msg = f"This is my secret message {i}"
        red = redact()
        sig, mh, sk, pk = red.sign(msg)
        is_verify = red.verify(pk, mh, sig)
        print(is_verify)

        fin = 'msgi.txt'
        pin = os.path.join(path, 'out', fin)
        write_signature_for_zokrates_cli(pk, sig, mh, pin)

        a = pk.p.x.n.to_bytes(32, 'big') + pk.p.y.n.to_bytes(32, 'big')
        h1 = int(hashlib.sha256(a).hexdigest(), 16) % P
        print("H1 : %d, 0x%x"%(h1, h1))

        M0 = mh.hex()[:64]
        M1 = mh.hex()[64:]
        b0 = [int(M0[i:i+8], 16) for i in range(0,len(M0), 8)]
        b0.extend([int(M1[i:i+8], 16) for i in range(0,len(M1), 8)])
        b = b""
        b += b"".join(m.to_bytes(4, "big") for m in b0)
        h2 = int(hashlib.sha256(b).hexdigest(), 16) % P
        print("H2 : %d, 0x%x" %(h2, h2))

        with open(pin, "r+") as f:
            content = f.read()
            print(content)
            f.seek(0, 0)
            f.write(str(h1) + ' ' + str(h2) + ' ' + content)


        os.system(f'cat {fin} | xargs zokrates compute-witness -a')
        os.system(f'zokrates generate-proof')
        logs = os.popen(f'zokrates verify').readlines()
        print("Verification===", logs[-1].strip('\n'))
        if logs[-1].strip() == 'PASSED':
            print("Pass")
        else:
            print("Fail")


    