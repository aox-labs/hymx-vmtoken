package schema

import (
	"math/big"
)

type IDB interface {
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
