package schema

import "math/big"

type IDB interface {
	Info() Info
	SetInfo(newInfo Info)
	Owner() string
	SetOwner(newOwner string)
	MintOwner() string
	SetMintOwner(newOwner string)
	MaxSupply() *big.Int
	SetMaxSupply(*big.Int) error

	GetTotalSupply() *big.Int
	SetTotalSupply(*big.Int)
	BalanceOf(accId string) (*big.Int, error)
	Balances() (map[string]*big.Int, error)
	UpdateBalance(accId string, amount *big.Int) error

	CacheInitial() bool
	CacheInitialed()

	Checkpoint() (data string, err error)
	Restore(data string) error
}
