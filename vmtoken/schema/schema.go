package schema

import "math/big"

const (
	VmTokenModuleFormat = "aox.token.0.0.1"
)

type Info struct {
	Id       string
	Name     string
	Ticker   string
	Decimals string
	Logo     string
}

type TokenSnapshot struct {
	Id           string              `json:"id"`
	Name         string              `json:"name"`
	Ticker       string              `json:"ticker"`
	Decimals     string              `json:"decimals"`
	Logo         string              `json:"logo"`
	TotalSupply  *big.Int            `json:"totalSupply"`
	Balances     map[string]*big.Int `json:"balances"`
	Owner        string              `json:"owner"`
	MintOwner    string              `json:"mintOwner"`
	BurnFee      *big.Int            `json:"burnFee"`
	FeeRecipient string              `json:"feeRecipient"`
}
