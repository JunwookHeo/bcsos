package mining

import (
	"sync"

	"github.com/junwookheo/bcsos/common/wallet"
)

type WalletMgr struct {
	w     *wallet.Wallet
	mutex sync.Mutex
}

var (
	wm         *WalletMgr
	oncewallet sync.Once
)

func (wm *WalletMgr) GetWallet() *wallet.Wallet {
	return wm.w
}

func WalletMgrInst(path string) *WalletMgr {
	if path == "" {
		return wm
	}

	oncewallet.Do(func() {
		wm = &WalletMgr{
			w:     wallet.NewWallet(path),
			mutex: sync.Mutex{},
		}
	})
	return wm
}
