package cache

import (
	"encoding/json"
	"math/big"
	"sync"

	"github.com/aox-labs/hymx-vmtoken/vmtoken/basic/schema"
	dbSchema "github.com/aox-labs/hymx-vmtoken/vmtoken/db/cache/schema"
	"github.com/hymatrix/hymx/vmm/utils"
)

type BasicToken struct {
	info schema.Info

	maxSupply   *big.Int
	totalSupply *big.Int
	balances    map[string]*big.Int
	owner       string
	mintOwner   string
	initialSync bool
	rwlock      sync.RWMutex
}

func NewBasicToken(info schema.Info, owner string, mintOwner string, maxSupply *big.Int) *BasicToken {
	_, mintOwner, _ = utils.IDCheck(mintOwner)
	return &BasicToken{
		info:        info,
		maxSupply:   maxSupply,
		totalSupply: big.NewInt(0),
		balances:    map[string]*big.Int{},
		owner:       owner,
		mintOwner:   mintOwner,
	}
}

func (b *BasicToken) Info() schema.Info {
	b.rwlock.RLock()
	defer b.rwlock.RUnlock()
	return b.info
}

func (b *BasicToken) SetInfo(newInfo schema.Info) {
	b.rwlock.Lock()
	defer b.rwlock.Unlock()
	b.info = newInfo
}

func (b *BasicToken) Owner() string {
	b.rwlock.RLock()
	defer b.rwlock.RUnlock()
	return b.owner
}

func (b *BasicToken) SetOwner(newOwner string) {
	b.rwlock.Lock()
	defer b.rwlock.Unlock()
	b.owner = newOwner
}

func (b *BasicToken) MintOwner() string {
	b.rwlock.RLock()
	defer b.rwlock.RUnlock()
	return b.mintOwner
}

func (b *BasicToken) SetMintOwner(newOwner string) {
	b.rwlock.Lock()
	defer b.rwlock.Unlock()
	_, mintOwner, _ := utils.IDCheck(newOwner)
	b.mintOwner = mintOwner
}

func (b *BasicToken) MaxSupply() *big.Int {
	b.rwlock.RLock()
	defer b.rwlock.RUnlock()
	if b.maxSupply == nil {
		return big.NewInt(0)
	}
	return new(big.Int).Set(b.maxSupply)
}

func (b *BasicToken) SetMaxSupply(newMaxSupply *big.Int) error {
	b.rwlock.Lock()
	defer b.rwlock.Unlock()
	if newMaxSupply == nil {
		b.maxSupply = big.NewInt(0)
	} else {
		b.maxSupply = new(big.Int).Set(newMaxSupply)
	}
	return nil
}

func (b *BasicToken) GetTotalSupply() *big.Int {
	b.rwlock.RLock()
	defer b.rwlock.RUnlock()
	if b.totalSupply == nil {
		return big.NewInt(0)
	}
	return new(big.Int).Set(b.totalSupply)
}

func (b *BasicToken) SetTotalSupply(newTotalSupply *big.Int) {
	b.rwlock.Lock()
	defer b.rwlock.Unlock()
	if newTotalSupply == nil {
		b.totalSupply = big.NewInt(0)
	} else {
		b.totalSupply = new(big.Int).Set(newTotalSupply)
	}
}

func (b *BasicToken) BalanceOf(accId string) (*big.Int, error) {
	_, accId, err := utils.IDCheck(accId)
	if err != nil {
		return nil, err
	}
	b.rwlock.RLock()
	defer b.rwlock.RUnlock()
	balance, exists := b.balances[accId]
	if !exists || balance == nil {
		return big.NewInt(0), nil
	}
	return new(big.Int).Set(balance), nil
}

func (b *BasicToken) Balances() (map[string]*big.Int, error) {
	b.rwlock.RLock()
	defer b.rwlock.RUnlock()
	result := make(map[string]*big.Int, len(b.balances))
	for k, v := range b.balances {
		if v == nil {
			result[k] = big.NewInt(0)
		} else {
			result[k] = new(big.Int).Set(v)
		}
	}
	return result, nil
}

func (b *BasicToken) UpdateBalance(accId string, amount *big.Int) error {
	_, accId, err := utils.IDCheck(accId)
	if err != nil {
		return err
	}
	b.rwlock.Lock()
	defer b.rwlock.Unlock()
	if b.balances == nil {
		b.balances = make(map[string]*big.Int)
	}
	if amount == nil || amount.Cmp(big.NewInt(0)) == 0 {
		delete(b.balances, accId)
	} else {
		b.balances[accId] = new(big.Int).Set(amount)
	}
	return nil
}

func (b *BasicToken) CacheInitial() bool {
	b.rwlock.RLock()
	defer b.rwlock.RUnlock()
	return b.initialSync
}

func (b *BasicToken) CacheInitialed() {
	b.rwlock.Lock()
	defer b.rwlock.Unlock()
	b.initialSync = true
}

func (b *BasicToken) Checkpoint() (data string, err error) {
	b.rwlock.RLock()
	defer b.rwlock.RUnlock()
	snap := dbSchema.BasicSnapshot{
		Id:          b.info.Id,
		Name:        b.info.Name,
		Ticker:      b.info.Ticker,
		Decimals:    b.info.Decimals,
		Logo:        b.info.Logo,
		Description: b.info.Description,
		TotalSupply: b.totalSupply,
		Balances:    b.balances,
		Owner:       b.owner,
		MintOwner:   b.mintOwner,
		MaxSupply:   b.maxSupply,
	}
	by, err := json.Marshal(snap)
	if err != nil {
		return "", err
	}
	return string(by), nil
}

func (b *BasicToken) Restore(data string) error {
	snap := &dbSchema.BasicSnapshot{}
	if err := json.Unmarshal([]byte(data), snap); err != nil {
		return err
	}
	b.rwlock.Lock()
	defer b.rwlock.Unlock()
	b.owner = snap.Owner
	b.mintOwner = snap.MintOwner
	b.balances = snap.Balances
	if b.balances == nil {
		b.balances = make(map[string]*big.Int)
	}
	b.totalSupply = snap.TotalSupply
	if b.totalSupply == nil {
		b.totalSupply = big.NewInt(0)
	}
	b.maxSupply = snap.MaxSupply
	if b.maxSupply == nil {
		b.maxSupply = big.NewInt(0)
	}
	b.info = schema.Info{
		Id:          snap.Id,
		Name:        snap.Name,
		Ticker:      snap.Ticker,
		Decimals:    snap.Decimals,
		Logo:        snap.Logo,
		Description: snap.Description,
	}
	return nil
}
