package schema

import "math/big"

const (
	VmTokenBasicModuleFormat           = "hymx.basic.token.0.0.1"
	VmTokenCrossChainModuleFormat      = "hymx.crosschain.token.0.0.1"
	VmTokenCrossChainMultiModuleFormat = "hymx.cross.chain.multi.token.0.0.1"
)

type Info struct {
	Id          string
	Name        string
	Ticker      string
	Decimals    string
	Logo        string
	Description string
}

// BasicSnapshot represents a snapshot of a basic token without Burn functionality
type BasicSnapshot struct {
	Id          string              `json:"id"`
	Name        string              `json:"name"`
	Ticker      string              `json:"ticker"`
	Decimals    string              `json:"decimals"`
	Logo        string              `json:"logo"`
	Description string              `json:"description"`
	TotalSupply *big.Int            `json:"totalSupply"`
	Balances    map[string]*big.Int `json:"balances"`
	Owner       string              `json:"owner"`
	MintOwner   string              `json:"mintOwner"`
	MaxSupply   *big.Int            `json:"maxSupply"`
}

// CrossChainSnapshot extends BasicSnapshot with Burn functionality
type CrossChainSnapshot struct {
	BasicSnapshot
	BurnFee       *big.Int `json:"burnFee"`
	FeeRecipient  string   `json:"feeRecipient"`
	BurnProcessor string   `json:"burnProcessor"`
}

// CrossChainMultiSnapshot extends BasicSnapshot with cross-chain multi-token functionality
type CrossChainMultiSnapshot struct {
	BasicSnapshot
	MintedRecords     map[string]string   `json:"mintedRecords"`
	SourceTokenChains map[string]string   `json:"sourceTokenChains"` // key: sourceTokenId, val: sourceChain
	SourceLockAmounts map[string]*big.Int `json:"sourceLockAmounts"` // key: sourceChain:sourceTokenId, val: lock amount
	BurnFees          map[string]*big.Int `json:"burnFees"`          // key: chainType, val: burn fee
	FeeRecipient      string              `json:"feeRecipient"`
	BurnProcessor     string              `json:"burnProcessor"`
}
