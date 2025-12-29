package crosschain

import (
	"encoding/json"
	"github.com/aox-labs/hymx-vmtoken/schema"
	"maps"
)

func (t *Token) initCache() (cache map[string]string) {
	if t.basic.DB.CacheInitial() {
		return
	}
	defer t.basic.DB.CacheInitialed()
	cache = map[string]string{}
	maps.Copy(cache, t.basic.CacheBalances())
	maps.Copy(cache, t.basic.CacheTotalSupply())
	maps.Copy(cache, t.cacheTokenInfo())
	return
}

func (t *Token) cacheTokenInfo() map[string]string {
	// Serialize burn fees map
	burnFeesJson, _ := json.Marshal(t.db.GetBurnFees())

	// Serialize source token chains map
	sourceTokenChainsJson, _ := json.Marshal(t.db.GetSourceTokenChains())

	// Serialize source lock amounts map
	sourceLockAmountsJson, _ := json.Marshal(t.db.GetSourceLockAmounts())

	info := t.basic.DB.Info()
	cacheInfo := schema.CrossChainCacheInfo{
		Name:              info.Name,
		Ticker:            info.Ticker,
		Decimals:          info.Decimals,
		Logo:              info.Logo,
		Description:       info.Description,
		Owner:             t.basic.DB.Owner(),
		MintOwner:         t.basic.DB.MintOwner(),
		BurnFees:          string(burnFeesJson),
		FeeRecipient:      t.db.GetFeeRecipient(),
		BurnProcessor:     t.db.GetBurnProcessor(),
		SourceTokenChains: string(sourceTokenChainsJson),
		SourceLockAmounts: string(sourceLockAmountsJson),
	}

	res, _ := json.Marshal(cacheInfo)
	return map[string]string{
		"info": string(res),
	}
}
