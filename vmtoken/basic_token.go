package vmtoken

import (
	"encoding/json"
	"maps"
	"math/big"
	"strings"

	"github.com/aox-labs/hymx-vmtoken/vmtoken/schema"
	"github.com/hymatrix/hymx/common"
	vmmSchema "github.com/hymatrix/hymx/vmm/schema"
	"github.com/hymatrix/hymx/vmm/utils"
	goarSchema "github.com/permadao/goar/schema"
)

var log = common.NewLog("vm_token")

// BasicToken represents a basic token without Burn functionality
type BasicToken struct {
	InitialSync bool
	Info        schema.Info

	TotalSupply *big.Int
	Balances    map[string]*big.Int
	Owner       string
	MintOwner   string
}

// NewBasicToken creates a new basic token VM
func NewBasicToken(info schema.Info, owner string, mintOwner string) *BasicToken {
	return &BasicToken{
		InitialSync: false,
		Info:        info,
		TotalSupply: big.NewInt(0),
		Balances:    map[string]*big.Int{},
		Owner:       owner,
		MintOwner:   mintOwner,
	}
}

// SpawnBasicToken spawns a basic token VM without Burn support
func SpawnBasicToken(env vmmSchema.Env) (vm vmmSchema.Vm, err error) {
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
		mintOwnerStr = env.AccId // Default to owner
	}
	_, mintOwner, err := utils.IDCheck(mintOwnerStr)
	if err != nil {
		err = schema.ErrInvalidMintOwner // Reuse error type for now
		return
	}

	vm = NewBasicToken(schema.Info{
		Id:       env.Id,
		Name:     env.Meta.Params["Name"],
		Ticker:   env.Meta.Params["Ticker"],
		Decimals: env.Meta.Params["Decimals"],
		Logo:     env.Meta.Params["Logo"],
	}, env.AccId, mintOwner)

	return vm, nil
}

func (v *BasicToken) CacheTokenInfo() map[string]string {
	tokenInfo := map[string]string{
		"Name":         v.Info.Name,
		"Ticker":       v.Info.Ticker,
		"Logo":         v.Info.Logo,
		"Denomination": v.Info.Decimals,
		"Owner":        v.Owner,
		"MintOwner":    v.MintOwner,
	}
	res, _ := json.Marshal(tokenInfo)
	return map[string]string{
		"TokenInfo": string(res),
	}
}

func (v *BasicToken) CacheBalances(updateBalances map[string]*big.Int) map[string]string {
	cacheMap := make(map[string]string)
	for k, vl := range updateBalances {
		if vl == nil {
			vl = big.NewInt(0)
		}
		cacheMap["Balances:"+k] = vl.String()
	}
	cacheMap["TotalSupply"] = v.TotalSupply.String()
	return cacheMap
}

func (v *BasicToken) Apply(from string, meta vmmSchema.Meta) (res *vmmSchema.Result, err error) {
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
	// Basic token does not support Burn
	default:
		return
	}
}

func (v *BasicToken) Checkpoint() (data string, err error) {
	snap := schema.BasicSnapshot{
		Id:          v.Info.Id,
		Name:        v.Info.Name,
		Ticker:      v.Info.Ticker,
		Decimals:    v.Info.Decimals,
		Logo:        v.Info.Logo,
		TotalSupply: v.TotalSupply,
		Balances:    v.Balances,
		Owner:       v.Owner,
		MintOwner:   v.MintOwner,
	}
	by, err := json.Marshal(snap)
	if err != nil {
		return
	}
	data = string(by)
	return
}

func (v *BasicToken) Restore(data string) error {
	snap := &schema.BasicSnapshot{}
	if err := json.Unmarshal([]byte(data), snap); err != nil {
		return err
	}

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
	return nil
}

func (v *BasicToken) Close() error {
	return nil
}

// Basic token specific methods
func (v *BasicToken) HandleInfo(from string) (res *vmmSchema.Result, err error) {
	// Define all token info tags
	tags := []goarSchema.Tag{
		{Name: "Name", Value: v.Info.Name},
		{Name: "Ticker", Value: v.Info.Ticker},
		{Name: "Logo", Value: v.Info.Logo},
		{Name: "Denomination", Value: v.Info.Decimals},
		{Name: "Owner", Value: v.Owner},
		{Name: "MintOwner", Value: v.MintOwner},
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

func (v *BasicToken) HandleSetParams(from string, meta vmmSchema.Meta) (res *vmmSchema.Result, applyErr error) {
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
	if from != v.Owner {
		err = schema.ErrIncorrectOwner
		return
	}

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

func (v *BasicToken) HandleTotalSupply(from string) (res *vmmSchema.Result, err error) {
	res = &vmmSchema.Result{
		Messages: []*vmmSchema.ResMessage{
			{
				Target: from,
				Data:   v.TotalSupply.String(),
				Tags: []goarSchema.Tag{
					{Name: "Action", Value: "Total-Supply"},
					{Name: "Ticker", Value: v.Info.Ticker},
				},
			},
		},
	}
	return
}

func (v *BasicToken) HandleBalanceOf(from string, params map[string]string) (res *vmmSchema.Result, err error) {
	// Determine account to query (default to sender if not specified)
	accountId := from
	if recipient, exists := params["Recipient"]; exists && recipient != "" {
		accountId = recipient
	} else if target, exists := params["Target"]; exists && target != "" {
		accountId = target
	}

	balance := v.BalanceOf(accountId)

	res = &vmmSchema.Result{
		Messages: []*vmmSchema.ResMessage{
			{
				Target: from,
				Data:   balance.String(),
				Tags: []goarSchema.Tag{
					{Name: "Balance", Value: balance.String()},
					{Name: "Ticker", Value: v.Info.Ticker},
					{Name: "Account", Value: accountId},
				},
			},
		},
	}
	return
}

func (v *BasicToken) HandleTransfer(itemId, from string, params map[string]string) (res *vmmSchema.Result, applyErr error) {
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
					{Name: "Action", Value: "Transfer-Error"},
					{Name: "TransactionId", Value: itemId},
					{Name: "Error", Value: err.Error()},
				},
			})
			res.Error = err.Error()
		}
	}()

	// Parse and validate recipient
	recipient, exists := params["Recipient"]
	if !exists {
		err = schema.ErrMissingRecipient
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

	// Execute transfer operation
	if err = v.Transfer(from, recipient, amount); err != nil {
		return
	}

	// Create debit notice for sender
	debitNotice := &vmmSchema.ResMessage{
		Target: from,
		Data:   "You transferred " + quantity + " to " + recipient,
		Tags: []goarSchema.Tag{
			{Name: "Ticker", Value: v.Info.Ticker},
			{Name: "Action", Value: "Debit-Notice"},
			{Name: "Recipient", Value: recipient},
			{Name: "Quantity", Value: quantity},
			{Name: "TransactionId", Value: itemId},
		},
	}

	// Create credit notice for recipient
	creditNotice := &vmmSchema.ResMessage{
		Target: recipient,
		Data:   "You received " + quantity + " from " + from,
		Tags: []goarSchema.Tag{
			{Name: "Ticker", Value: v.Info.Ticker},
			{Name: "Action", Value: "Credit-Notice"},
			{Name: "Sender", Value: from},
			{Name: "Quantity", Value: quantity},
			{Name: "TransactionId", Value: itemId},
		},
	}

	// Forward X- prefixed tags to both messages
	for key, value := range params {
		if strings.HasPrefix(key, "X-") {
			debitNotice.Tags = append(debitNotice.Tags, goarSchema.Tag{Name: key, Value: value})
			creditNotice.Tags = append(creditNotice.Tags, goarSchema.Tag{Name: key, Value: value})
		}
	}

	res = &vmmSchema.Result{
		Messages: []*vmmSchema.ResMessage{debitNotice, creditNotice},
		Cache: v.CacheBalances(map[string]*big.Int{
			from:      v.BalanceOf(from),
			recipient: v.BalanceOf(recipient),
		}),
	}
	return
}

func (v *BasicToken) HandleMint(from string, params map[string]string) (res *vmmSchema.Result, applyErr error) {
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

	// Parse and validate recipient
	recipient, exists := params["Recipient"]
	if !exists {
		err = schema.ErrMissingRecipient
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

	// Execute mint operation
	if err = v.Mint(recipient, amount); err != nil {
		return
	}

	// Create mint notice for owner
	ownerNotice := &vmmSchema.ResMessage{
		Target: from,
		Tags: []goarSchema.Tag{
			{Name: "Action", Value: "Mint-Notice"},
			{Name: "Recipient", Value: recipient},
			{Name: "Quantity", Value: quantity},
			{Name: "Ticker", Value: v.Info.Ticker},
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
		},
	}

	res = &vmmSchema.Result{
		Messages: []*vmmSchema.ResMessage{ownerNotice, recipientNotice},
		Cache: v.CacheBalances(map[string]*big.Int{
			recipient: v.BalanceOf(recipient),
		}),
	}
	return
}

// Core token operations
func (v *BasicToken) Mint(to string, amount *big.Int) (err error) {
	// Validate and normalize recipient address
	_, to, err = utils.IDCheck(to)
	if err != nil {
		return
	}

	// Add tokens to recipient
	if err = v.Add(to, amount); err != nil {
		log.Error("mint: token add failed", "err", err)
		return
	}

	// Increase total supply
	v.TotalSupply = new(big.Int).Add(v.TotalSupply, amount)
	return
}

func (v *BasicToken) Transfer(from, to string, amount *big.Int) (err error) {
	// Validate and normalize recipient address
	_, to, err = utils.IDCheck(to)
	if err != nil {
		return
	}

	// Deduct tokens from sender
	if err = v.Sub(from, amount); err != nil {
		log.Error("transfer: token sub failed", "err", err)
		return
	}

	// Add tokens to recipient
	if err = v.Add(to, amount); err != nil {
		log.Error("transfer: token add failed", "err", err)
		return
	}

	return nil
}

func (v *BasicToken) Sub(accId string, amount *big.Int) error {
	// Skip operation if amount is zero
	if amount.Cmp(big.NewInt(0)) == 0 {
		return nil
	}

	// Check sufficient balance
	currentBalance := v.BalanceOf(accId)
	if currentBalance.Cmp(amount) < 0 {
		return schema.ErrInsufficientBalance
	}

	// Calculate new balance and update
	newBalance := new(big.Int).Sub(currentBalance, amount)
	v.Balances[accId] = newBalance

	return nil
}

func (v *BasicToken) Add(accId string, amount *big.Int) error {
	// Skip operation if amount is zero
	if amount.Cmp(big.NewInt(0)) == 0 {
		return nil
	}

	// Get current balance and calculate new balance
	currentBalance := v.BalanceOf(accId)
	newBalance := new(big.Int).Add(currentBalance, amount)

	// Update balance
	v.Balances[accId] = newBalance

	return nil
}

func (v *BasicToken) BalanceOf(accId string) *big.Int {
	balance, exists := v.Balances[accId]
	if !exists {
		balance = big.NewInt(0)
	}
	return balance
}
