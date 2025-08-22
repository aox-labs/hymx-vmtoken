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

// CrossChainToken extends BasicToken with Burn functionality
type CrossChainToken struct {
	*BasicToken
	BurnFee       *big.Int
	FeeRecipient  string
	BurnProcessor string
}

// NewCrossChainToken creates a new cross-chain token VM
func NewCrossChainToken(info schema.Info, owner string, mintOwner string, burnFee *big.Int, feeRecipient string, burnProcessor string) *CrossChainToken {
	return &CrossChainToken{
		BasicToken:    NewBasicToken(info, owner, mintOwner),
		BurnFee:       burnFee,
		FeeRecipient:  feeRecipient,
		BurnProcessor: burnProcessor,
	}
}

// SpawnCrossChainToken spawns a cross-chain token VM with Burn support
func SpawnCrossChainToken(env vmmSchema.Env) (vm vmmSchema.Vm, err error) {
	// Validate required parameters
	requiredParams := []string{"Name", "Ticker", "Decimals"}
	for _, param := range requiredParams {
		if env.Meta.Params[param] == "" {
			err = schema.ErrIncorrectTokenInfo
			return
		}
	}

	// Parse and validate BurnFee with default value
	burnFeeStr := env.Meta.Params["BurnFee"]
	if burnFeeStr == "" {
		burnFeeStr = "0"
	}
	burnFee, ok := new(big.Int).SetString(burnFeeStr, 10)
	if !ok {
		err = schema.ErrInvalidQuantityFormat
		return
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

	vm = NewCrossChainToken(schema.Info{
		Id:       env.Id,
		Name:     env.Meta.Params["Name"],
		Ticker:   env.Meta.Params["Ticker"],
		Decimals: env.Meta.Params["Decimals"],
		Logo:     env.Meta.Params["Logo"],
	}, env.AccId, mintOwner, burnFee, feeRecipient, burnProcessor)

	return vm, nil
}

// Override cacheTokenInfo to include Burn-specific fields
func (v *CrossChainToken) CacheTokenInfo() map[string]string {
	tokenInfo := map[string]string{
		"Name":          v.Info.Name,
		"Ticker":        v.Info.Ticker,
		"Logo":          v.Info.Logo,
		"Denomination":  v.Info.Decimals,
		"Owner":         v.Owner,
		"MintOwner":     v.MintOwner,
		"BurnFee":       v.BurnFee.String(),
		"FeeRecipient":  v.FeeRecipient,
		"BurnProcessor": v.BurnProcessor,
	}

	res, _ := json.Marshal(tokenInfo)
	return map[string]string{
		"TokenInfo": string(res),
	}
}

// Override Apply to enable Burn functionality and handle cross-chain specific actions
func (v *CrossChainToken) Apply(from string, meta vmmSchema.Meta) (res *vmmSchema.Result, err error) {
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
		return v.HandleMint(from, meta.Params)
	case "Burn":
		return v.HandleBurn(from, meta.Params)
	default:
		return
	}
}

// Override Checkpoint to include Burn-specific fields
func (v *CrossChainToken) Checkpoint() (data string, err error) {
	snap := schema.CrossChainSnapshot{
		BasicSnapshot: schema.BasicSnapshot{
			Id:          v.Info.Id,
			Name:        v.Info.Name,
			Ticker:      v.Info.Ticker,
			Decimals:    v.Info.Decimals,
			Logo:        v.Info.Logo,
			TotalSupply: v.TotalSupply,
			Balances:    v.Balances,
			Owner:       v.Owner,
			MintOwner:   v.MintOwner,
		},
		BurnFee:       v.BurnFee,
		FeeRecipient:  v.FeeRecipient,
		BurnProcessor: v.BurnProcessor,
	}
	by, err := json.Marshal(snap)
	if err != nil {
		return
	}
	data = string(by)
	return
}

// Override Restore to handle Burn-specific fields
func (v *CrossChainToken) Restore(data string) error {
	snap := &schema.CrossChainSnapshot{}
	if err := json.Unmarshal([]byte(data), snap); err != nil {
		return err
	}

	// Restore base token fields
	v.Owner = snap.Owner
	v.MintOwner = snap.MintOwner
	v.Balances = snap.Balances
	v.TotalSupply = snap.TotalSupply
	v.Info = schema.Info{
		Id:       snap.Id,
		Name:     snap.Name,
		Ticker:   snap.Ticker,
		Decimals: snap.Decimals,
		Logo:     snap.Logo,
	}

	// Restore cross-chain specific fields
	v.FeeRecipient = snap.FeeRecipient
	v.BurnFee = snap.BurnFee
	v.BurnProcessor = snap.BurnProcessor
	return nil
}

// Override HandleInfo to include Burn-specific fields
func (v *CrossChainToken) HandleInfo(from string) (res *vmmSchema.Result, err error) {
	// Define all token info tags
	tags := []goarSchema.Tag{
		{Name: "Name", Value: v.Info.Name},
		{Name: "Ticker", Value: v.Info.Ticker},
		{Name: "Logo", Value: v.Info.Logo},
		{Name: "Denomination", Value: v.Info.Decimals},
		{Name: "Owner", Value: v.Owner},
		{Name: "MintOwner", Value: v.MintOwner},
		{Name: "BurnFee", Value: v.BurnFee.String()},
		{Name: "FeeRecipient", Value: v.FeeRecipient},
		{Name: "BurnProcessor", Value: v.BurnProcessor},
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
func (v *CrossChainToken) HandleSetParams(from string, meta vmmSchema.Meta) (res *vmmSchema.Result, applyErr error) {
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
		v.Owner = meta.Params["Owner"]
	}

	if meta.Params["MintOwner"] != "" {
		v.MintOwner = meta.Params["MintOwner"]
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

	// Handle cross-chain specific parameters
	if feeRecipient, exists := meta.Params["FeeRecipient"]; exists && feeRecipient != "" {
		v.FeeRecipient = feeRecipient
	}

	if burnFeeStr, exists := meta.Params["BurnFee"]; exists && burnFeeStr != "" {
		if burnFee, ok := new(big.Int).SetString(burnFeeStr, 10); ok {
			v.BurnFee = burnFee
		}
	}

	if burnProcessorStr, exists := meta.Params["BurnProcessor"]; exists && burnProcessorStr != "" {
		_, accId, err := utils.IDCheck(burnProcessorStr)
		if err != nil {
			return
		}
		v.BurnProcessor = accId
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

// Cross-chain specific method: HandleBurn
func (v *CrossChainToken) HandleBurn(from string, params map[string]string) (res *vmmSchema.Result, applyErr error) {
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
	_, _, err = utils.IDCheck(recipient)
	if err != nil {
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

	// Execute burn operation
	if err = v.Burn(from, amt); err != nil {
		return
	}

	// Create burn notice message
	netBurnAmount := new(big.Int).Sub(amt, v.BurnFee)
	creditNotice := &vmmSchema.ResMessage{
		Target: v.BurnProcessor,
		Tags: []goarSchema.Tag{
			{Name: "Action", Value: "Burn-Notice"},
			{Name: "Sender", Value: from},
			{Name: "X-Recipient", Value: recipient},
			{Name: "Quantity", Value: netBurnAmount.String()},
			{Name: "Ticker", Value: v.Info.Ticker},
			{Name: "TokenId", Value: v.Info.Id},
			{Name: "Fee", Value: v.BurnFee.String()},
			{Name: "FeeRecipient", Value: v.FeeRecipient},
		},
	}

	// Prepare result with cache updates
	res = &vmmSchema.Result{
		Messages: []*vmmSchema.ResMessage{creditNotice},
		Cache: v.CacheBalances(map[string]*big.Int{
			from:           v.BalanceOf(from),
			v.FeeRecipient: v.BalanceOf(v.FeeRecipient),
		}),
	}
	return
}

// Cross-chain specific method: Burn
func (v *CrossChainToken) Burn(from string, amount *big.Int) (err error) {
	// Validate burn amount is sufficient to cover fee
	if amount.Cmp(v.BurnFee) < 0 {
		err = schema.ErrIncorrectQuantity
		return
	}

	// Deduct full amount from sender
	if err = v.Sub(from, amount); err != nil {
		return
	}

	// Transfer burn fee to fee recipient
	if err = v.Add(v.FeeRecipient, v.BurnFee); err != nil {
		return
	}

	// Calculate net burn amount (total - fee) and reduce total supply
	netBurnAmount := new(big.Int).Sub(amount, v.BurnFee)
	v.TotalSupply = new(big.Int).Sub(v.TotalSupply, netBurnAmount)

	return
}
