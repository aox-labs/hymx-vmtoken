package crosschain

import (
	"encoding/json"
	"maps"
	"math/big"

	"github.com/aox-labs/hymx-vmtoken/vmtoken/crosschain/schema"
	vmmSchema "github.com/hymatrix/hymx/vmm/schema"
	"github.com/hymatrix/hymx/vmm/utils"
	goarSchema "github.com/permadao/goar/schema"
)

func (t *Token) handleInfo(from string) (res vmmSchema.Result) {
	// Serialize multi-chain data
	burnFeesJson, _ := json.Marshal(t.db.GetBurnFees())
	sourceTokenChainsJson, _ := json.Marshal(t.db.GetSourceTokenChains())
	sourceLockAmountsJson, _ := json.Marshal(t.db.GetSourceLockAmounts())
	feeRecipient := t.db.GetFeeRecipient()
	burnProcessor := t.db.GetBurnProcessor()

	// Define all token info tags
	info := t.basic.DB.Info()
	tags := []goarSchema.Tag{
		{Name: "Name", Value: info.Name},
		{Name: "Ticker", Value: info.Ticker},
		{Name: "Logo", Value: info.Logo},
		{Name: "Decimals", Value: info.Decimals},
		{Name: "Description", Value: info.Description},
		{Name: "Owner", Value: t.basic.DB.Owner()},
		{Name: "MintOwner", Value: t.basic.DB.MintOwner()},
		{Name: "BurnFees", Value: string(burnFeesJson)},
		{Name: "FeeRecipient", Value: feeRecipient},
		{Name: "BurnProcessor", Value: burnProcessor},
		{Name: "SourceTokenChains", Value: string(sourceTokenChainsJson)},
		{Name: "SourceLockAmounts", Value: string(sourceLockAmountsJson)},
	}

	res.Messages = []*vmmSchema.ResMessage{
		{
			Target: from,
			Tags:   tags,
		},
	}
	res.Cache = t.initCache()
	return
}

func (t *Token) handleSetParams(from string, meta vmmSchema.Meta) (res vmmSchema.Result) {
	// Check ownership
	if from != t.basic.DB.Owner() {
		res.Error = schema.ErrIncorrectOwner
		return
	}

	// Handle base token parameters
	if meta.Params["TokenOwner"] != "" {
		_, newOwner, err := utils.IDCheck(meta.Params["TokenOwner"])
		if err != nil {
			res.Error = schema.ErrInvalidOwner
			return
		}
		t.basic.DB.SetOwner(newOwner)
	}

	if meta.Params["MintOwner"] != "" {
		_, newOwner, err := utils.IDCheck(meta.Params["MintOwner"])
		if err != nil {
			res.Error = schema.ErrInvalidMintOwner
			return
		}
		t.basic.DB.SetMintOwner(newOwner)
	}

	info := t.basic.DB.Info()
	if meta.Params["Name"] != "" {
		info.Name = meta.Params["Name"]
	}

	if meta.Params["Ticker"] != "" {
		info.Ticker = meta.Params["Ticker"]
	}

	if meta.Params["Decimals"] != "" {
		info.Decimals = meta.Params["Decimals"]
	}

	if meta.Params["Logo"] != "" {
		info.Logo = meta.Params["Logo"]
	}

	if meta.Params["Description"] != "" {
		info.Description = meta.Params["Description"]
	}
	t.basic.DB.SetInfo(info)

	// Handle multi-chain specific parameters
	if meta.Params["FeeRecipient"] != "" {
		_, feeRecipient, err := utils.IDCheck(meta.Params["FeeRecipient"])
		if err != nil {
			res.Error = schema.ErrInvalidFeeRecipient
			return
		}
		t.db.SetFeeRecipient(feeRecipient)
	}

	// Handle chain-specific burn fees from JSON
	if burnFeesStr, exists := meta.Params["BurnFees"]; exists && burnFeesStr != "" {
		var burnFeesMap map[string]string
		if err := json.Unmarshal([]byte(burnFeesStr), &burnFeesMap); err == nil {
			for chainType, feeStr := range burnFeesMap {
				if burnFee, ok := new(big.Int).SetString(feeStr, 10); ok {
					t.db.SetBurnFee(chainType, burnFee)
				}
			}
		}
	}

	if meta.Params["BurnProcessor"] != "" {
		_, burnProcessor, err := utils.IDCheck(meta.Params["BurnProcessor"])
		if err != nil {
			res.Error = schema.ErrInvalidBurnProcessor
			return
		}
		t.db.SetBurnProcessor(burnProcessor)
	}

	res.Messages = []*vmmSchema.ResMessage{
		{
			Target: from,
			Tags: []goarSchema.Tag{
				{Name: "Set-Params-Notice", Value: "success"},
			},
		},
	}
	res.Cache = t.cacheTokenInfo()
	return
}

func (t *Token) handleCrossChainMint(from string, params map[string]string) (res vmmSchema.Result) {
	// Check minting permission
	if from != t.basic.DB.MintOwner() {
		res.Error = schema.ErrIncorrectOwner
		return
	}

	if _, ok := t.db.GetMintedRecord(params["X-MintTxHash"]); ok {
		res.Error = schema.ErrRepeatMint
		return
	}

	// Parse and validate recipient
	recipient, exists := params["Recipient"]
	if !exists {
		res.Error = schema.ErrMissingRecipient
		return
	}

	_, recipient, err := utils.IDCheck(recipient)
	if err != nil {
		res.Error = schema.ErrInvalidRecipient
		return
	}

	// Parse and validate quantity
	quantity, exists := params["Quantity"]
	if !exists {
		res.Error = schema.ErrMissingQuantity
		return
	}

	amount, ok := new(big.Int).SetString(quantity, 10)
	if !ok {
		res.Error = schema.ErrInvalidQuantityFormat
		return
	}

	// Parse source chain and token information
	sourceChainType := params["SourceChainType"]
	if sourceChainType == "" {
		res.Error = schema.ErrMissingSourceChain
		return
	}

	sourceTokenId := params["SourceTokenId"]
	if sourceTokenId == "" {
		res.Error = schema.ErrMissingSourceTokenId
		return
	}

	_, sourceTokenId, err = utils.IDCheck(sourceTokenId)
	if err != nil {
		res.Error = schema.ErrInvalidSourceTokenId
		return
	}

	// verify chainType and tokenId
	chainType, ok := t.db.GetSourceTokenChain(sourceTokenId)
	if !ok {
		t.db.SetSourceTokenChain(sourceTokenId, sourceChainType)
	} else {
		if chainType != sourceChainType {
			res.Error = schema.ErrIncorrectSourceChainType
			return
		}
	}
	// change balances
	err = t.basic.Mint(recipient, amount)
	if err != nil {
		res.Error = err
		return
	}

	// change lock amount
	curLockAmt, ok := t.db.GetSourceLockAmount(sourceTokenId, sourceChainType)
	if !ok {
		curLockAmt = big.NewInt(0)
	}
	t.db.SetSourceLockAmount(sourceTokenId, sourceChainType, new(big.Int).Add(curLockAmt, amount))

	// Create mint notice for owner
	ownerNotice := &vmmSchema.ResMessage{
		Target: from,
		Tags: []goarSchema.Tag{
			{Name: "Action", Value: "Mint-Notice"},
			{Name: "Recipient", Value: recipient},
			{Name: "Quantity", Value: quantity},
			{Name: "Ticker", Value: t.basic.DB.Info().Ticker},
			{Name: "SourceChainType", Value: sourceChainType},
			{Name: "SourceTokenId", Value: sourceTokenId},
			{Name: "X-MintTxHash", Value: params["X-MintTxHash"]},
		},
	}

	// Create mint notice for recipient
	recipientNotice := &vmmSchema.ResMessage{
		Target: recipient,
		Tags: []goarSchema.Tag{
			{Name: "Action", Value: "Mint-Notice"},
			{Name: "Recipient", Value: recipient},
			{Name: "Quantity", Value: quantity},
			{Name: "Ticker", Value: t.basic.DB.Info().Ticker},
			{Name: "SourceChainType", Value: sourceChainType},
			{Name: "SourceTokenId", Value: sourceTokenId},
			{Name: "X-MintTxHash", Value: params["X-MintTxHash"]},
		},
	}

	t.db.SetMintedRecord(params["X-MintTxHash"], sourceChainType)

	res.Messages = []*vmmSchema.ResMessage{ownerNotice, recipientNotice}
	res.Cache = map[string]string{}
	maps.Copy(res.Cache, t.cacheTokenInfo())
	maps.Copy(res.Cache, t.basic.CacheTotalSupply())
	maps.Copy(res.Cache, t.basic.CacheChangeBalance(recipient))
	return
}

// handleCrossChainBurn handles cross-chain burning with target chain selection
func (t *Token) handleCrossChainBurn(from string, params map[string]string) (res vmmSchema.Result) {
	// Determine recipient (default to sender if not specified)
	recipient := params["Recipient"]
	if recipient == "" {
		recipient = params["X-Recipient"]
		if recipient == "" {
			recipient = from
		}
	}

	// Validate recipient address
	_, recipient, err := utils.IDCheck(recipient)
	if err != nil {
		res.Error = schema.ErrInvalidRecipient
		return
	}

	// Parse and validate quantity
	qty, exists := params["Quantity"]
	if !exists {
		res.Error = schema.ErrMissingQuantity
		return
	}

	amt, ok := new(big.Int).SetString(qty, 10)
	if !ok {
		res.Error = schema.ErrInvalidQuantityFormat
		return
	}

	// Parse target chain
	targetTokenId := params["TargetTokenId"]
	if targetTokenId == "" {
		res.Error = schema.ErrMissingTargetTokenId
		return
	}

	_, targetTokenId, err = utils.IDCheck(targetTokenId)
	if err != nil {
		res.Error = schema.ErrInvalidTargetTokenId
		return
	}

	targetChainType, ok := t.db.GetSourceTokenChain(targetTokenId)
	if !ok {
		res.Error = schema.ErrIncorrectTargetTokenId
		return
	}

	// get burn fee
	burnFee, ok := t.db.GetBurnFee(targetChainType)
	if !ok {
		res.Error = schema.ErrMissingBurnFee
		return
	}

	// Execute cross-chain burn operation
	if err = t.crossChainBurn(from, amt, burnFee, targetChainType, targetTokenId); err != nil {
		res.Error = err
		return
	}

	// Create burn notice message
	netBurnAmount := new(big.Int).Sub(amt, burnFee)
	creditNotice := &vmmSchema.ResMessage{
		Target: t.db.GetBurnProcessor(),
		Tags: []goarSchema.Tag{
			{Name: "Action", Value: "Burn-Notice"},
			{Name: "Sender", Value: from},
			{Name: "X-Recipient", Value: recipient},
			{Name: "Quantity", Value: netBurnAmount.String()},
			{Name: "Ticker", Value: t.basic.DB.Info().Ticker},
			{Name: "WrappedTokenId", Value: t.basic.DB.Info().Id},
			{Name: "Fee", Value: burnFee.String()},
			{Name: "FeeRecipient", Value: t.db.GetFeeRecipient()},
			{Name: "TargetChainType", Value: targetChainType},
			{Name: "TargetTokenId", Value: targetTokenId},
		},
	}

	// Prepare result with cache updates
	res.Messages = []*vmmSchema.ResMessage{creditNotice}
	res.Cache = map[string]string{}
	maps.Copy(res.Cache, t.cacheTokenInfo())
	maps.Copy(res.Cache, t.basic.CacheChangeBalance(from, t.db.GetFeeRecipient()))
	maps.Copy(res.Cache, t.basic.CacheTotalSupply())
	return
}

// crossChainBurn burns tokens and reduces target chain lock amounts
func (t *Token) crossChainBurn(from string, amount *big.Int,
	burnFee *big.Int, targetChainType, targetTokenId string) (err error) {
	// Validate burn amount is sufficient to cover fee
	if amount.Cmp(burnFee) < 0 {
		err = schema.ErrIncorrectQuantity
		return
	}

	lockAmt, ok := t.db.GetSourceLockAmount(targetTokenId, targetChainType)
	if !ok {
		err = schema.ErrLockAmountEmpty
		return
	}
	netBurnAmount := new(big.Int).Sub(amount, burnFee)
	if lockAmt.Cmp(netBurnAmount) < 0 {
		err = schema.ErrInsufficientLockAmount
		return
	}

	// Deduct full amount from sender
	if err = t.basic.Sub(from, amount); err != nil {
		return
	}

	// Transfer burn fee to fee recipient
	if err = t.basic.Add(t.db.GetFeeRecipient(), burnFee); err != nil {
		return
	}

	// Calculate net burn amount (total - fee) and reduce total supply
	t.basic.DB.SetTotalSupply(new(big.Int).Sub(t.basic.DB.GetTotalSupply(), netBurnAmount))

	// Reduce lock amount for the target chain
	t.db.SetSourceLockAmount(targetTokenId, targetChainType, new(big.Int).Sub(lockAmt, netBurnAmount))
	return
}
