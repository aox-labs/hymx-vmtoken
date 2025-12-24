package schema

import "math/big"

type BasicDB interface {
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

type CrossChainDB interface {
	GetMintedRecord(mintTxHash string) (string, bool)
	SetMintedRecord(mintTxHash, chainType string)

	GetSourceTokenChains() map[string]string
	GetSourceTokenChain(tokenId string) (string, bool)
	SetSourceTokenChain(tokenId, chainType string)

	GetSourceLockAmounts() map[string]*big.Int
	GetSourceLockAmount(tokenId, chainType string) (*big.Int, bool)
	SetSourceLockAmount(tokenId, chainType string, amount *big.Int)

	GetBurnFees() map[string]*big.Int
	GetBurnFee(chainType string) (*big.Int, bool)
	SetBurnFee(chainType string, amount *big.Int)

	GetFeeRecipient() string
	SetFeeRecipient(addr string)

	GetBurnProcessor() string
	SetBurnProcessor(addr string)

	Checkpoint() (data string, err error)
	Restore(data string) error
}
