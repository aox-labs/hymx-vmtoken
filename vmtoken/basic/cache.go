package basic

import (
	"encoding/json"
	"maps"
	"math/big"
)

func (b *Token) initCache() (cache map[string]string) {
	if !b.DB.CacheInitial() {
		return
	}
	defer b.DB.CacheInitialed()

	maps.Copy(cache, b.CacheBalances())
	maps.Copy(cache, b.CacheTotalSupply())
	maps.Copy(cache, b.cacheTokenInfo())
	return
}

func (b *Token) cacheTokenInfo() map[string]string {
	info := b.DB.Info()
	tokenInfo := map[string]string{
		"Name":         info.Name,
		"Ticker":       info.Ticker,
		"Logo":         info.Logo,
		"Denomination": info.Decimals,
		"Description":  info.Description,
		"Owner":        b.DB.Owner(),
		"MintOwner":    b.DB.MintOwner(),
		"MaxSupply":    b.DB.MaxSupply().String(),
	}
	res, _ := json.Marshal(tokenInfo)
	return map[string]string{
		"token-info": string(res),
	}
}

func (b *Token) CacheChangeBalance(updateAccounts ...string) map[string]string {
	cacheMap := make(map[string]string)
	for _, accId := range updateAccounts {
		bal, err := b.DB.BalanceOf(accId)
		if err != nil {
			bal = big.NewInt(0)
		}
		cacheMap["balances:"+accId] = bal.String()
	}
	return cacheMap
}

func (b *Token) CacheTotalSupply() map[string]string {
	cacheMap := make(map[string]string)
	cacheMap["total-supply"] = b.DB.GetTotalSupply().String()

	return cacheMap
}

func (b *Token) CacheBalances() map[string]string {
	cacheMap := make(map[string]string)
	balances, err := b.DB.Balances()
	if err == nil {
		balanceBy, _ := json.Marshal(balances)
		cacheMap["balances"] = string(balanceBy)
	}

	for k, vl := range balances {
		if vl == nil {
			vl = big.NewInt(0)
		}
		cacheMap["balances:"+k] = vl.String()
	}
	return cacheMap
}
