package crosschain

import (
	"encoding/json"
	"github.com/aox-labs/hymx-vmtoken/vmtoken/db/cache"
	"github.com/hymatrix/hymx/vmm/utils"
	"math/big"

	"github.com/aox-labs/hymx-vmtoken/vmtoken/basic"
	basicSchema "github.com/aox-labs/hymx-vmtoken/vmtoken/basic/schema"
	"github.com/aox-labs/hymx-vmtoken/vmtoken/crosschain/schema"
	vmmSchema "github.com/hymatrix/hymx/vmm/schema"
)

type Token struct {
	basic *basic.Token
	db    schema.IDB
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

	// Parse and validate BurnFees for different chains
	burnFees := make(map[string]*big.Int)

	// Parse chain-specific burn fees from JSON
	burnFeesStr := env.Meta.Params["BurnFees"]
	if burnFeesStr != "" {
		var burnFeesMap map[string]string
		if err = json.Unmarshal([]byte(burnFeesStr), &burnFeesMap); err == nil {
			for chainType, feeStr := range burnFeesMap {
				if burnFee, ok := new(big.Int).SetString(feeStr, 10); ok {
					burnFees[chainType] = burnFee
				}
			}
		}
	}

	// Parse and validate FeeRecipient with default value
	feeRecipientStr := env.Meta.Params["FeeRecipient"]
	if feeRecipientStr == "" {
		feeRecipientStr = env.Meta.AccId // Default to owner
	}
	_, feeRecipient, err := utils.IDCheck(feeRecipientStr)
	if err != nil {
		err = schema.ErrInvalidFeeRecipient
		return
	}

	// Parse and validate MintOwner with default value
	mintOwnerStr := env.Meta.Params["MintOwner"]
	if mintOwnerStr == "" {
		mintOwnerStr = env.Meta.AccId // Default to owner
	}
	_, mintOwner, err := utils.IDCheck(mintOwnerStr)
	if err != nil {
		err = schema.ErrInvalidMintOwner
		return
	}

	// Parse and validate BurnProcessor with default value
	burnProcessorStr := env.Meta.Params["BurnProcessor"]
	if burnProcessorStr == "" {
		burnProcessorStr = env.Meta.AccId // Default to owner
	}
	_, burnProcessor, err := utils.IDCheck(burnProcessorStr)
	if err != nil {
		err = schema.ErrInvalidFeeRecipient // Reuse error type for now
		return
	}

	basicToken := &basic.Token{
		DB: cache.NewBasicToken(basicSchema.Info{
			Id:          env.Meta.ItemId,
			Name:        env.Meta.Params["Name"],
			Ticker:      env.Meta.Params["Ticker"],
			Decimals:    env.Meta.Params["Decimals"],
			Logo:        env.Meta.Params["Logo"],
			Description: env.Meta.Params["Description"],
		}, env.Meta.AccId, mintOwner, big.NewInt(0)),
	}
	return &Token{
		basic: basicToken,
		db:    cache.NewCrossChainToken(burnFees, feeRecipient, burnProcessor),
	}, nil
}

func (t *Token) Apply(from string, meta vmmSchema.Meta) (res vmmSchema.Result) {
	switch meta.Action {
	case "Info":
		res = t.handleInfo(from)
	case "Set-Params":
		res = t.handleSetParams(from, meta)
	case "Total-Supply":
		res = t.basic.HandleTotalSupply(from)
	case "Balance":
		res = t.basic.HandleBalanceOf(from, meta.Params)
	case "Transfer":
		res = t.basic.HandleTransfer(meta.ItemId, from, meta.Params)
	case "Mint":
		res = t.handleCrossChainMint(from, meta.Params)
	case "Burn":
		res = t.handleCrossChainBurn(from, meta.Params)
	}
	return
}

func (t *Token) Checkpoint() (data string, err error) {
	basicCheckpoint, err := t.basic.Checkpoint()
	if err != nil {
		return "", err
	}

	ccCheckpoint, err := t.db.Checkpoint()
	if err != nil {
		return "", err
	}

	snap := map[string]string{
		"basic": basicCheckpoint,
		"cc":    ccCheckpoint,
	}

	by, err := json.Marshal(snap)
	if err != nil {
		return
	}
	data = string(by)
	return
}

func (t *Token) Restore(data string) error {
	snap := map[string]string{}
	if err := json.Unmarshal([]byte(data), &snap); err != nil {
		return err
	}
	if err := t.basic.DB.Restore(snap["basic"]); err != nil {
		return err
	}
	if err := t.db.Restore(snap["cc"]); err != nil {
		return err
	}
	return nil
}

func (t *Token) Close() error {
	return nil
}
