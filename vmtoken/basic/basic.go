package basic

import (
	"github.com/aox-labs/hymx-vmtoken/vmtoken/basic/schema"
	"github.com/aox-labs/hymx-vmtoken/vmtoken/db/cache"
	vmmSchema "github.com/hymatrix/hymx/vmm/schema"
	"github.com/hymatrix/hymx/vmm/utils"
	"math/big"
)

type Token struct {
	DB schema.IDB
}

func Spawn(env vmmSchema.Env) (vm vmmSchema.Vm, err error) {
	// Validate required parameters
	requiredParams := []string{"Name", "Ticker", "Decimals"}
	for _, param := range requiredParams {
		if env.Meta.Params[param] == "" {
			err = schema.ErrIncorrectTokenInfo
			return
		}
	}

	// Parse and validate MintOwner with default value
	mintOwnerStr := env.Meta.Params["MintOwner"]
	if mintOwnerStr == "" {
		mintOwnerStr = env.Meta.AccId // Default to owner
	}
	_, mintOwner, err := utils.IDCheck(mintOwnerStr)
	if err != nil {
		err = schema.ErrInvalidMintOwner // Reuse error type for now
		return
	}
	maxSupplyStr := env.Meta.Params["MaxSupply"]
	if mintOwnerStr == "" {
		mintOwnerStr = "0"
	}
	maxSupply, ok := new(big.Int).SetString(maxSupplyStr, 10)
	if !ok {
		err = schema.ErrInvalidMaxSupply // Reuse error type for now
		return
	}

	db := cache.NewBasicToken(schema.Info{
		Id:          env.Meta.ItemId,
		Name:        env.Meta.Params["Name"],
		Ticker:      env.Meta.Params["Ticker"],
		Decimals:    env.Meta.Params["Decimals"],
		Logo:        env.Meta.Params["Logo"],
		Description: env.Meta.Params["Description"],
	}, env.Meta.AccId, mintOwner, maxSupply)
	return &Token{DB: db}, nil
}

func (b *Token) Apply(from string, meta vmmSchema.Meta) (res vmmSchema.Result) {
	switch meta.Action {
	case "Info":
		res = b.handleInfo(from)
	case "Set-Params":
		res = b.handleSetParams(from, meta)
	case "Total-Supply":
		res = b.HandleTotalSupply(from)
	case "Balance":
		res = b.HandleBalanceOf(from, meta.Params)
	case "Transfer":
		res = b.HandleTransfer(meta.ItemId, from, meta.Params)
	case "Mint":
		res = b.handleMint(from, meta.Params)
	}
	return
}

func (b *Token) Checkpoint() (string, error) {
	return b.DB.Checkpoint()
}

func (b *Token) Restore(data string) error {
	return b.DB.Restore(data)
}

func (b *Token) Close() error {
	return nil
}
