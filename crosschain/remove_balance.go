package crosschain

import (
	"maps"
	"math/big"

	"github.com/aox-labs/hymx-vmtoken/schema"
	vmmSchema "github.com/hymatrix/hymx/vmm/schema"
	"github.com/hymatrix/hymx/vmm/utils"
)

// handleRemoveBalance lets the directly signed Owner cancel an account's entire
// balance against the specified source token. It is not a cross-chain withdrawal.
func (t *Token) handleRemoveBalance(from string, meta vmmSchema.Meta) (res vmmSchema.Result) {
	if from != t.basic.DB.Owner() || meta.AccId != from || meta.FromProcess != "" {
		res.Error = schema.ErrIncorrectOwner
		return
	}
	if meta.Params["Recipient"] == "" {
		res.Error = schema.ErrMissingRecipient
		return
	}
	_, account, err := utils.IDCheck(meta.Params["Recipient"])
	if err != nil {
		res.Error = schema.ErrInvalidRecipient
		return
	}
	if meta.Params["SourceTokenId"] == "" {
		res.Error = schema.ErrMissingSourceTokenId
		return
	}
	_, sourceToken, err := utils.IDCheck(meta.Params["SourceTokenId"])
	if err != nil {
		res.Error = schema.ErrInvalidSourceTokenId
		return
	}
	chain, ok := t.db.GetSourceTokenChain(sourceToken)
	if !ok {
		res.Error = schema.ErrInvalidSourceTokenId
		return
	}
	amount, err := t.basic.DB.BalanceOf(account)
	if err != nil || amount.Sign() <= 0 {
		res.Error = schema.ErrInsufficientBalance
		return
	}
	lock, ok := t.db.GetSourceLockAmount(sourceToken, chain)
	if !ok {
		res.Error = schema.ErrLockAmountEmpty
		return
	}
	if lock.Cmp(amount) < 0 {
		res.Error = schema.ErrInsufficientLockAmount
		return
	}
	supply := t.basic.DB.GetTotalSupply()
	if supply.Cmp(amount) < 0 {
		res.Error = schema.ErrIncorrectQuantity
		return
	}

	// UpdateBalance(0) removes the map entry. Validate everything before mutation.
	if err := t.basic.DB.UpdateBalance(account, big.NewInt(0)); err != nil {
		res.Error = err
		return
	}
	t.basic.DB.SetTotalSupply(new(big.Int).Sub(supply, amount))
	t.db.SetSourceLockAmount(sourceToken, chain, new(big.Int).Sub(lock, amount))
	res.Cache = t.basic.CacheChangeBalance(account)
	maps.Copy(res.Cache, t.basic.CacheTotalSupply())
	maps.Copy(res.Cache, t.cacheTokenInfo())
	res.Output = map[string]string{
		"Account": account, "Removed": amount.String(), "SourceTokenId": sourceToken,
	}
	// No Messages, Spawns or Assignments: in particular, never emit Burn-Notice.
	return
}
