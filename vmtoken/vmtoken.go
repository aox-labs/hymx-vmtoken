package vmtoken

import (
	"encoding/json"
	"github.com/aox-labs/hymx-vmtoken/vmtoken/schema"
	"github.com/hymatrix/hymx/common"
	vmmSchema "github.com/hymatrix/hymx/vmm/schema"
	"maps"
	"math/big"
)

var log = common.NewLog("vmtoken")

func SpawnVmToken(env vmmSchema.Env) (vm vmmSchema.Vm, err error) {
	name := env.Meta.Params["Name"]
	ticker := env.Meta.Params["Ticker"]
	decimals := env.Meta.Params["Decimals"]
	logo := env.Meta.Params["Logo"]

	if name == "" || ticker == "" || decimals == "" {
		err = schema.ErrIncorrectTokenInfo
		return
	}

	vm = New(schema.Info{
		Id:       env.Id,
		Name:     name,
		Ticker:   ticker,
		Decimals: decimals,
		Logo:     logo,
	}, env.AccId)
	return vm, nil
}

type VmToken struct {
	initialSync bool
	info        schema.Info

	totalSupply  *big.Int
	balances     map[string]*big.Int
	owner        string
	mintOwner    string
	burnFee      *big.Int
	feeRecipient string
}

func New(info schema.Info, owner string) *VmToken {
	return &VmToken{
		initialSync:  false,
		info:         info,
		totalSupply:  big.NewInt(0),
		balances:     map[string]*big.Int{},
		owner:        owner,
		mintOwner:    owner,
		burnFee:      big.NewInt(0),
		feeRecipient: owner,
	}
}

func (v *VmToken) cacheTokenInfo() map[string]string {
	tokenInfo := map[string]string{
		"Name":         v.info.Name,
		"Ticker":       v.info.Ticker,
		"Denomination": v.info.Decimals,
		"Logo":         v.info.Logo,
		"Owner":        v.owner,
		"MintOwner":    v.mintOwner,
		"BurnFee":      v.burnFee.String(),
		"FeeRecipient": v.feeRecipient,
	}
	res, _ := json.Marshal(tokenInfo)
	return map[string]string{
		"TokenInfo": string(res),
	}
}

func (v *VmToken) cacheBalances(updateBalances map[string]*big.Int) map[string]string {
	cacheMap := make(map[string]string)
	for k, vl := range updateBalances {
		if vl == nil {
			vl = big.NewInt(0)
		}
		cacheMap["Balances:"+k] = vl.String()
	}
	cacheMap["TotalSupply"] = v.totalSupply.String()
	return cacheMap
}

func (v *VmToken) Apply(from string, meta vmmSchema.Meta) (res *vmmSchema.Result, err error) {
	switch meta.Action {
	case "Info":
		defer func() { // init cache
			if v.initialSync == false {
				balMap := v.cacheBalances(v.balances)
				infoMap := v.cacheTokenInfo()
				mergedMap := make(map[string]string)
				maps.Copy(mergedMap, balMap)
				maps.Copy(mergedMap, infoMap)
				res.Cache = mergedMap

				v.initialSync = true
			}
		}()
		return v.handleInfo(from)
	case "Set-Params":
		return v.handleSetParams(from, meta)
	case "Total-Supply", "TotalSupply":
		return v.handleTotalSupply(from)
	// case "Balances":
	// 	return v.handleBalances(from)
	case "Balance":
		return v.handleBalanceOf(from, meta.Params)
	case "Transfer":
		return v.handleTransfer(meta.ItemId, from, meta.Params)
	case "Mint":
		return v.handleMint(from, meta.Params)
	// case "Burn":
	// 	return v.handleBurn(from, meta.Params)
	default:
		return
	}
}

func (v *VmToken) Checkpoint() (data string, err error) {
	snap := schema.TokenSnapshot{
		Id:           v.info.Id,
		Name:         v.info.Name,
		Ticker:       v.info.Ticker,
		Decimals:     v.info.Decimals,
		Logo:         v.info.Logo,
		TotalSupply:  v.totalSupply,
		Balances:     v.balances,
		Owner:        v.owner,
		MintOwner:    v.mintOwner,
		BurnFee:      v.burnFee,
		FeeRecipient: v.feeRecipient,
	}
	by, err := json.Marshal(snap)
	if err != nil {
		return
	}
	data = string(by)
	return
}

func (v *VmToken) Restore(data string) error {
	snap := &schema.TokenSnapshot{}
	if err := json.Unmarshal([]byte(data), snap); err != nil {
		return err
	}

	v.owner = snap.Owner
	v.mintOwner = snap.MintOwner
	v.feeRecipient = snap.FeeRecipient
	v.burnFee = snap.BurnFee
	v.balances = snap.Balances
	v.totalSupply = snap.TotalSupply
	v.info = schema.Info{
		Id:       snap.Id,
		Name:     snap.Name,
		Ticker:   snap.Ticker,
		Decimals: snap.Decimals,
		Logo:     snap.Logo,
	}
	return nil
}

func (v *VmToken) Close() error {
	return nil
}
