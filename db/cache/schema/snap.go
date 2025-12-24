package schema

import "math/big"

// BasicSnapshot represents a snapshot of a basic token for checkpoint/restore
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

// CrossChainMultiSnapshot represents a snapshot of a cross-chain multi token for checkpoint/restore
type CrossChainMultiSnapshot struct {
	MintedRecords     map[string]string   `json:"mintedRecords"`     // key: X-MintTxHash val: chainType
	SourceTokenChains map[string]string   `json:"sourceTokenChains"` // key: sourceTokenId, val: sourceChainType
	SourceLockAmounts map[string]*big.Int `json:"sourceLockAmounts"` // key: sourceChain:sourceTokenId, val: source chain locked amount
	BurnFees          map[string]*big.Int `json:"burnFees"`          // key: chainType, val: burn fee
	FeeRecipient      string              `json:"feeRecipient"`
	BurnProcessor     string              `json:"burnProcessor"`
}
