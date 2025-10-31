package vmtoken

import (
	"encoding/json"
	"maps"
	"math/big"

	"github.com/aox-labs/hymx-vmtoken/vmtoken/schema"

	vmmSchema "github.com/hymatrix/hymx/vmm/schema"
	"github.com/hymatrix/hymx/vmm/utils"
	goarSchema "github.com/permadao/goar/schema"
)

// CrossChainMultiToken extends BasicToken with cross-chain multi-token functionality
type CrossChainMultiToken struct {
	*BasicToken

	MintedRecords     map[string]string   // key: X-MintTxHash val: chainType
	SourceTokenChains map[string]string   // key: sourceTokenId, val: sourceChainType
	SourceLockAmounts map[string]*big.Int // key: sourceChain:sourceTokenId, val: source chain locked amount
	BurnFees          map[string]*big.Int // key: chainType, val: burn fee
	FeeRecipient      string
	BurnProcessor     string
}

// NewCrossChainMultiToken creates a new cross-chain multi-token VM
func NewCrossChainMultiToken(info schema.Info, owner string, mintOwner string, burnFees map[string]*big.Int, feeRecipient string, burnProcessor string) *CrossChainMultiToken {
	return &CrossChainMultiToken{
		BasicToken:        NewBasicToken(info, owner, mintOwner, big.NewInt(0)),
		MintedRecords:     make(map[string]string),
		SourceTokenChains: make(map[string]string),
		SourceLockAmounts: make(map[string]*big.Int),
		BurnFees:          burnFees,
		FeeRecipient:      feeRecipient,
		BurnProcessor:     burnProcessor,
	}
}

// SpawnCrossChainMultiToken spawns a cross-chain multi-token VM with multi-chain support
func SpawnCrossChainMultiToken(env vmmSchema.Env) (vm vmmSchema.Vm, err error) {
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
		feeRecipientStr = env.AccId // Default to owner
	}
	_, feeRecipient, err := utils.IDCheck(feeRecipientStr)
	if err != nil {
		err = schema.ErrInvalidFeeRecipient
		return
	}

	// Parse and validate MintOwner with default value
	mintOwnerStr := env.Meta.Params["MintOwner"]
	if mintOwnerStr == "" {
		mintOwnerStr = env.AccId // Default to owner
	}
	_, mintOwner, err := utils.IDCheck(mintOwnerStr)
	if err != nil {
		err = schema.ErrInvalidMintOwner
		return
	}

	// Parse and validate BurnProcessor with default value
	burnProcessorStr := env.Meta.Params["BurnProcessor"]
	if burnProcessorStr == "" {
		burnProcessorStr = env.AccId // Default to owner
	}
	_, burnProcessor, err := utils.IDCheck(burnProcessorStr)
	if err != nil {
		err = schema.ErrInvalidFeeRecipient // Reuse error type for now
		return
	}

	vm = NewCrossChainMultiToken(schema.Info{
		Id:          env.Id,
		Name:        env.Meta.Params["Name"],
		Ticker:      env.Meta.Params["Ticker"],
		Decimals:    env.Meta.Params["Decimals"],
		Logo:        env.Meta.Params["Logo"],
		Description: env.Meta.Params["Description"],
	}, env.AccId, mintOwner, burnFees, feeRecipient, burnProcessor)

	return vm, nil
}

// Override cacheTokenInfo to include multi-chain specific fields
func (v *CrossChainMultiToken) CacheTokenInfo() map[string]string {
	// Serialize burn fees map
	burnFeesJson, _ := json.Marshal(v.BurnFees)

	// Serialize source token chains map
	sourceTokenChainsJson, _ := json.Marshal(v.SourceTokenChains)

	// Serialize source lock amounts map
	sourceLockAmountsJson, _ := json.Marshal(v.SourceLockAmounts)

	tokenInfo := map[string]string{
		"Name":              v.Info.Name,
		"Ticker":            v.Info.Ticker,
		"Logo":              v.Info.Logo,
		"Denomination":      v.Info.Decimals,
		"Description":       v.Info.Description,
		"Owner":             v.Owner,
		"MintOwner":         v.MintOwner,
		"BurnFees":          string(burnFeesJson),
		"FeeRecipient":      v.FeeRecipient,
		"BurnProcessor":     v.BurnProcessor,
		"SourceTokenChains": string(sourceTokenChainsJson),
		"SourceLockAmounts": string(sourceLockAmountsJson),
	}

	res, _ := json.Marshal(tokenInfo)
	return map[string]string{
		"TokenInfo": string(res),
	}
}

// Override Apply to enable multi-chain functionality and handle cross-chain specific actions
func (v *CrossChainMultiToken) Apply(from string, meta vmmSchema.Meta) (res *vmmSchema.Result, err error) {
	switch meta.Action {
	case "Info":
		defer func() { // Initialize cache on first Info call
			if !v.InitialSync {
				balMap := v.CacheBalances(v.Balances)
				infoMap := v.CacheTokenInfo()
				mergedMap := make(map[string]string)
				maps.Copy(mergedMap, balMap)
				maps.Copy(mergedMap, infoMap)
				res.Cache = mergedMap
				v.InitialSync = true
			}
		}()
		return v.HandleInfo(from)
	case "Set-Params":
		return v.HandleSetParams(from, meta)
	case "Total-Supply", "TotalSupply":
		return v.HandleTotalSupply(from)
	case "Balance":
		return v.HandleBalanceOf(from, meta.Params)
	case "Transfer":
		return v.HandleTransfer(meta.ItemId, from, meta.Params)
	case "Mint":
		return v.HandleCrossChainMint(from, meta.Params)
	case "Burn":
		return v.HandleCrossChainBurn(from, meta.Params)
	default:
		return
	}
}

// Override Checkpoint to include multi-chain specific fields
func (v *CrossChainMultiToken) Checkpoint() (data string, err error) {
	snap := schema.CrossChainMultiSnapshot{
		BasicSnapshot: schema.BasicSnapshot{
			Id:          v.Info.Id,
			Name:        v.Info.Name,
			Ticker:      v.Info.Ticker,
			Decimals:    v.Info.Decimals,
			Logo:        v.Info.Logo,
			Description: v.Info.Description,
			TotalSupply: v.TotalSupply,
			Balances:    v.Balances,
			Owner:       v.Owner,
			MintOwner:   v.MintOwner,
			MaxSupply:   v.MaxSupply,
		},
		MintedRecords:     v.MintedRecords,
		SourceTokenChains: v.SourceTokenChains,
		SourceLockAmounts: v.SourceLockAmounts,
		BurnFees:          v.BurnFees,
		FeeRecipient:      v.FeeRecipient,
		BurnProcessor:     v.BurnProcessor,
	}
	by, err := json.Marshal(snap)
	if err != nil {
		return
	}
	data = string(by)
	return
}

// Override Restore to handle multi-chain specific fields
func (v *CrossChainMultiToken) Restore(data string) error {
	snap := &schema.CrossChainMultiSnapshot{}
	if err := json.Unmarshal([]byte(data), snap); err != nil {
		return err
	}

	// Restore base token fields
	v.Owner = snap.Owner
	v.MintOwner = snap.MintOwner
	v.Balances = snap.Balances
	if v.Balances == nil {
		v.Balances = make(map[string]*big.Int)
	}
	v.TotalSupply = snap.TotalSupply
	if v.TotalSupply == nil {
		v.TotalSupply = big.NewInt(0)
	}
	v.MaxSupply = snap.MaxSupply
	if v.MaxSupply == nil {
		v.MaxSupply = big.NewInt(0)
	}
	v.Info = schema.Info{
		Id:          snap.Id,
		Name:        snap.Name,
		Ticker:      snap.Ticker,
		Decimals:    snap.Decimals,
		Logo:        snap.Logo,
		Description: snap.Description,
	}

	// Restore multi-chain specific fields
	v.MintedRecords = snap.MintedRecords
	if v.MintedRecords == nil {
		v.MintedRecords = make(map[string]string)
	}
	v.SourceTokenChains = snap.SourceTokenChains
	if v.SourceTokenChains == nil {
		v.SourceTokenChains = make(map[string]string)
	}
	v.SourceLockAmounts = snap.SourceLockAmounts
	if v.SourceLockAmounts == nil {
		v.SourceLockAmounts = make(map[string]*big.Int)
	}
	v.BurnFees = snap.BurnFees
	if v.BurnFees == nil {
		v.BurnFees = make(map[string]*big.Int)
	}
	v.FeeRecipient = snap.FeeRecipient
	v.BurnProcessor = snap.BurnProcessor
	return nil
}

// Override HandleInfo to include multi-chain specific fields
func (v *CrossChainMultiToken) HandleInfo(from string) (res *vmmSchema.Result, err error) {
	// Serialize multi-chain data
	burnFeesJson, _ := json.Marshal(v.BurnFees)
	sourceTokenChainsJson, _ := json.Marshal(v.SourceTokenChains)
	sourceLockAmountsJson, _ := json.Marshal(v.SourceLockAmounts)

	// Define all token info tags
	tags := []goarSchema.Tag{
		{Name: "Name", Value: v.Info.Name},
		{Name: "Ticker", Value: v.Info.Ticker},
		{Name: "Logo", Value: v.Info.Logo},
		{Name: "Denomination", Value: v.Info.Decimals},
		{Name: "Description", Value: v.Info.Description},
		{Name: "Owner", Value: v.Owner},
		{Name: "MintOwner", Value: v.MintOwner},
		{Name: "BurnFees", Value: string(burnFeesJson)},
		{Name: "FeeRecipient", Value: v.FeeRecipient},
		{Name: "BurnProcessor", Value: v.BurnProcessor},
		{Name: "SourceTokenChains", Value: string(sourceTokenChainsJson)},
		{Name: "SourceLockAmounts", Value: string(sourceLockAmountsJson)},
	}

	res = &vmmSchema.Result{
		Messages: []*vmmSchema.ResMessage{
			{
				Target: from,
				Tags:   tags,
			},
		},
	}
	return
}

// Override HandleSetParams to handle Burn-specific parameters
func (v *CrossChainMultiToken) HandleSetParams(from string, meta vmmSchema.Meta) (res *vmmSchema.Result, applyErr error) {
	var err error
	defer func() {
		if err != nil {
			if res == nil {
				res = &vmmSchema.Result{
					Messages: make([]*vmmSchema.ResMessage, 0),
				}
			}
			res.Messages = append(res.Messages, &vmmSchema.ResMessage{
				Target: from,
				Tags: []goarSchema.Tag{
					{Name: "Action", Value: "Set-Params-Error"},
					{Name: "Error", Value: err.Error()},
				},
			})
			res.Error = err.Error()
		}
	}()

	// Check ownership
	if from != v.Owner {
		err = schema.ErrIncorrectOwner
		return
	}

	// Handle base token parameters
	if meta.Params["Owner"] != "" {
		newOwner := meta.Params["Owner"]
		_, newOwner, err = utils.IDCheck(newOwner)
		if err != nil {
			err = schema.ErrInvalidOwner
			return
		}
		v.Owner = newOwner
	}

	if meta.Params["MintOwner"] != "" {
		newOwner := meta.Params["MintOwner"]
		_, newOwner, err = utils.IDCheck(newOwner)
		if err != nil {
			err = schema.ErrInvalidMintOwner
			return
		}
		v.MintOwner = newOwner
	}

	if meta.Params["Name"] != "" {
		v.Info.Name = meta.Params["Name"]
	}

	if meta.Params["Ticker"] != "" {
		v.Info.Ticker = meta.Params["Ticker"]
	}

	if meta.Params["Decimals"] != "" {
		v.Info.Decimals = meta.Params["Decimals"]
	}

	if meta.Params["Logo"] != "" {
		v.Info.Logo = meta.Params["Logo"]
	}

	if meta.Params["Description"] != "" {
		v.Info.Description = meta.Params["Description"]
	}

	// Handle multi-chain specific parameters
	if meta.Params["FeeRecipient"] != "" {
		feeRecipient := meta.Params["FeeRecipient"]
		_, feeRecipient, err = utils.IDCheck(feeRecipient)
		if err != nil {
			err = schema.ErrInvalidFeeRecipient
			return
		}
		v.FeeRecipient = feeRecipient
	}

	// Handle chain-specific burn fees from JSON
	if burnFeesStr, exists := meta.Params["BurnFees"]; exists && burnFeesStr != "" {
		var burnFeesMap map[string]string
		if err = json.Unmarshal([]byte(burnFeesStr), &burnFeesMap); err == nil {
			for chainType, feeStr := range burnFeesMap {
				if burnFee, ok := new(big.Int).SetString(feeStr, 10); ok {
					v.BurnFees[chainType] = burnFee
				}
			}
		}
	}

	if meta.Params["BurnProcessor"] != "" {
		burnProcessor := meta.Params["BurnProcessor"]
		_, burnProcessor, err = utils.IDCheck(burnProcessor)
		if err != nil {
			err = schema.ErrInvalidBurnProcessor
			return
		}
		v.BurnProcessor = burnProcessor
	}

	res = &vmmSchema.Result{
		Messages: []*vmmSchema.ResMessage{
			{
				Target: from,
				Tags: []goarSchema.Tag{
					{Name: "Set-Params-Notice", Value: "success"},
				},
			},
		},
		Cache: v.CacheTokenInfo(),
	}
	return
}

// HandleCrossChainMint handles cross-chain minting with source chain tracking
func (v *CrossChainMultiToken) HandleCrossChainMint(from string, params map[string]string) (res *vmmSchema.Result, applyErr error) {
	var err error
	defer func() {
		if err != nil {
			if res == nil {
				res = &vmmSchema.Result{
					Messages: make([]*vmmSchema.ResMessage, 0),
				}
			}
			res.Messages = append(res.Messages, &vmmSchema.ResMessage{
				Target: from,
				Tags: []goarSchema.Tag{
					{Name: "Action", Value: "Mint-Error"},
					{Name: "Error", Value: err.Error()},
				},
			})
			res.Error = err.Error()
		}
	}()

	// Check minting permission
	if from != v.MintOwner {
		err = schema.ErrIncorrectOwner
		return
	}

	if _, ok := v.MintedRecords[params["X-MintTxHash"]]; ok {
		err = schema.ErrRepeatMint
		return
	}

	// Parse and validate recipient
	recipient, exists := params["Recipient"]
	if !exists {
		err = schema.ErrMissingRecipient
		return
	}

	_, recipient, err = utils.IDCheck(recipient)
	if err != nil {
		err = schema.ErrInvalidRecipient
		return
	}

	// Parse and validate quantity
	quantity, exists := params["Quantity"]
	if !exists {
		err = schema.ErrMissingQuantity
		return
	}

	amount, ok := new(big.Int).SetString(quantity, 10)
	if !ok {
		err = schema.ErrInvalidQuantityFormat
		return
	}

	// Parse source chain and token information
	sourceChainType := params["SourceChainType"]
	if sourceChainType == "" {
		err = schema.ErrMissingSourceChain
		return
	}

	sourceTokenId := params["SourceTokenId"]
	if sourceTokenId == "" {
		err = schema.ErrMissingSourceTokenId
		return
	}

	_, sourceTokenId, err = utils.IDCheck(sourceTokenId)
	if err != nil {
		err = schema.ErrInvalidSourceTokenId
		return
	}

	// verify chainType and tokenId
	chainType, ok := v.SourceTokenChains[sourceTokenId]
	if !ok {
		v.SourceTokenChains[sourceTokenId] = sourceChainType
	} else {
		if chainType != sourceChainType {
			err = schema.ErrIncorrectSourceChainType
			return
		}
	}
	// change balances
	err = v.Mint(recipient, amount)
	if err != nil {
		return
	}

	// change lock amount
	lockKey := sourceChainType + ":" + sourceTokenId
	curLockAmt := v.SourceLockAmounts[lockKey]
	if curLockAmt == nil {
		curLockAmt = big.NewInt(0)
	}
	v.SourceLockAmounts[lockKey] = new(big.Int).Add(curLockAmt, amount)

	// Create mint notice for owner
	ownerNotice := &vmmSchema.ResMessage{
		Target: from,
		Tags: []goarSchema.Tag{
			{Name: "Action", Value: "Mint-Notice"},
			{Name: "Recipient", Value: recipient},
			{Name: "Quantity", Value: quantity},
			{Name: "Ticker", Value: v.Info.Ticker},
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
			{Name: "Ticker", Value: v.Info.Ticker},
			{Name: "SourceChainType", Value: sourceChainType},
			{Name: "SourceTokenId", Value: sourceTokenId},
			{Name: "X-MintTxHash", Value: params["X-MintTxHash"]},
		},
	}

	v.MintedRecords[params["X-MintTxHash"]] = sourceChainType
	mergedMap := v.CacheBalances(map[string]*big.Int{recipient: v.BalanceOf(recipient)})
	maps.Copy(mergedMap, v.CacheTokenInfo())
	res = &vmmSchema.Result{
		Messages: []*vmmSchema.ResMessage{ownerNotice, recipientNotice},
		Cache:    mergedMap,
	}
	return
}

// HandleCrossChainBurn handles cross-chain burning with target chain selection
func (v *CrossChainMultiToken) HandleCrossChainBurn(from string, params map[string]string) (res *vmmSchema.Result, applyErr error) {
	var err error
	defer func() {
		if err != nil {
			if res == nil {
				res = &vmmSchema.Result{
					Messages: make([]*vmmSchema.ResMessage, 0),
				}
			}
			res.Messages = append(res.Messages, &vmmSchema.ResMessage{
				Target: from,
				Tags: []goarSchema.Tag{
					{Name: "Action", Value: "Burn-Error"},
					{Name: "Error", Value: err.Error()},
				},
			})
			res.Error = err.Error()
		}
	}()

	// Determine recipient (default to sender if not specified)
	recipient := params["Recipient"]
	if recipient == "" {
		recipient = params["X-Recipient"]
		if recipient == "" {
			recipient = from
		}
	}

	// Validate recipient address
	_, recipient, err = utils.IDCheck(recipient)
	if err != nil {
		err = schema.ErrInvalidRecipient
		return
	}

	// Parse and validate quantity
	qty, exists := params["Quantity"]
	if !exists {
		err = schema.ErrMissingQuantity
		return
	}

	amt, ok := new(big.Int).SetString(qty, 10)
	if !ok {
		err = schema.ErrInvalidQuantityFormat
		return
	}

	// Parse target chain
	targetTokenId := params["TargetTokenId"]
	if targetTokenId == "" {
		err = schema.ErrMissingTargetTokenId
		return
	}

	_, targetTokenId, err = utils.IDCheck(targetTokenId)
	if err != nil {
		err = schema.ErrInvalidTargetTokenId
		return
	}

	targetChainType, ok := v.SourceTokenChains[targetTokenId]
	if !ok {
		err = schema.ErrIncorrectTargetTokenId
		return
	}

	// get burn fee
	burnFee, ok := v.BurnFees[targetChainType]
	if !ok {
		err = schema.ErrMissingBurnFee
		return
	}

	// Execute cross-chain burn operation
	if err = v.crossChainBurn(from, amt, burnFee, targetChainType, targetTokenId); err != nil {
		return
	}

	// Create burn notice message
	netBurnAmount := new(big.Int).Sub(amt, burnFee)
	creditNotice := &vmmSchema.ResMessage{
		Target: v.BurnProcessor,
		Tags: []goarSchema.Tag{
			{Name: "Action", Value: "Burn-Notice"},
			{Name: "Sender", Value: from},
			{Name: "X-Recipient", Value: recipient},
			{Name: "Quantity", Value: netBurnAmount.String()},
			{Name: "Ticker", Value: v.Info.Ticker},
			{Name: "WrappedTokenId", Value: v.Info.Id},
			{Name: "Fee", Value: burnFee.String()},
			{Name: "FeeRecipient", Value: v.FeeRecipient},
			{Name: "TargetChainType", Value: targetChainType},
			{Name: "TargetTokenId", Value: targetTokenId},
		},
	}

	// Prepare result with cache updates
	mergedMap := v.CacheBalances(map[string]*big.Int{from: v.BalanceOf(from), v.FeeRecipient: v.BalanceOf(v.FeeRecipient)})
	maps.Copy(mergedMap, v.CacheTokenInfo())
	res = &vmmSchema.Result{
		Messages: []*vmmSchema.ResMessage{creditNotice},
		Cache:    mergedMap,
	}
	return
}

// crossChainBurn burns tokens and reduces target chain lock amounts
func (v *CrossChainMultiToken) crossChainBurn(from string, amount *big.Int,
	burnFee *big.Int, targetChainType, targetTokenId string) (err error) {
	// Validate burn amount is sufficient to cover fee
	if amount.Cmp(burnFee) < 0 {
		err = schema.ErrIncorrectQuantity
		return
	}

	lockKey := targetChainType + ":" + targetTokenId
	lockAmt := v.SourceLockAmounts[lockKey]
	if lockAmt == nil {
		err = schema.ErrLockAmountEmpty
		return
	}
	netBurnAmount := new(big.Int).Sub(amount, burnFee)
	if lockAmt.Cmp(netBurnAmount) < 0 {
		err = schema.ErrInsufficientLockAmount
		return
	}

	// Deduct full amount from sender
	if err = v.Sub(from, amount); err != nil {
		return
	}

	// Transfer burn fee to fee recipient
	if err = v.Add(v.FeeRecipient, burnFee); err != nil {
		return
	}

	// Calculate net burn amount (total - fee) and reduce total supply
	v.TotalSupply = new(big.Int).Sub(v.TotalSupply, netBurnAmount)

	// Reduce lock amount for the target chain
	v.SourceLockAmounts[lockKey] = new(big.Int).Sub(lockAmt, netBurnAmount)
	return
}
