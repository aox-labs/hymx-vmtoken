package basic

import (
	"encoding/json"
	"github.com/aox-labs/hymx-vmtoken/schema"
	vmmSchema "github.com/hymatrix/hymx/vmm/schema"
	"github.com/hymatrix/hymx/vmm/utils"
	goarSchema "github.com/permadao/goar/schema"
	"golang.org/x/exp/maps"
	"math/big"
	"strings"
)

func (b *Token) handleInfo(from string) (res vmmSchema.Result) {
	info := b.DB.Info()
	cache := b.initCache()
	c, _ := json.Marshal(cache)
	res.Messages = []*vmmSchema.ResMessage{
		{
			Target: from,
			Tags: []goarSchema.Tag{
				{Name: "Name", Value: info.Name},
				{Name: "Ticker", Value: info.Ticker},
				{Name: "Logo", Value: info.Logo},
				{Name: "Decimals", Value: info.Decimals},
				{Name: "Description", Value: info.Description},
				{Name: "Owner", Value: b.DB.Owner()},
				{Name: "MintOwner", Value: b.DB.MintOwner()},
				{Name: "MaxSupply", Value: b.DB.MaxSupply().String()},
			},
			Data: string(c),
		},
	}
	res.Cache = cache
	return
}

func (b *Token) handleSetParams(from string, meta vmmSchema.Meta) (res vmmSchema.Result) {
	if from != b.DB.Owner() {
		res.Error = schema.ErrIncorrectOwner
		return
	}

	if meta.Params["TokenOwner"] != "" {
		_, newOwner, err := utils.IDCheck(meta.Params["TokenOwner"])
		if err != nil {
			res.Error = schema.ErrInvalidOwner
			return
		}
		b.DB.SetOwner(newOwner)
	}

	if meta.Params["MintOwner"] != "" {
		_, newOwner, err := utils.IDCheck(meta.Params["MintOwner"])
		if err != nil {
			res.Error = schema.ErrInvalidMintOwner
			return
		}
		b.DB.SetMintOwner(newOwner)
	}

	info := b.DB.Info()
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

	b.DB.SetInfo(info)

	res.Messages = []*vmmSchema.ResMessage{
		{
			Target: from,
			Tags: []goarSchema.Tag{
				{Name: "Set-Params-Notice", Value: "success"},
			},
		},
	}
	res.Cache = b.cacheTokenInfo()
	return
}

func (b *Token) HandleTotalSupply(from string) (res vmmSchema.Result) {
	res.Messages = []*vmmSchema.ResMessage{
		{
			Target: from,
			Data:   b.DB.GetTotalSupply().String(),
			Tags: []goarSchema.Tag{
				{Name: "Action", Value: "Total-Supply"},
			},
		},
	}
	return
}

func (b *Token) HandleBalanceOf(from string, params map[string]string) (res vmmSchema.Result) {
	// Determine account to query (default to sender if not specified)
	accountId := from
	if recipient, exists := params["Recipient"]; exists && recipient != "" {
		accountId = recipient
	} else if target, ok := params["Target"]; ok && target != "" {
		accountId = target
	}
	_, accountId, err := utils.IDCheck(accountId)
	if err != nil {
		res.Error = schema.ErrInvalidRecipient
		return
	}

	balance, err := b.DB.BalanceOf(accountId)
	if err != nil {
		res.Error = err
		return
	}

	res.Messages = []*vmmSchema.ResMessage{
		{
			Target: from,
			Data:   balance.String(),
			Tags: []goarSchema.Tag{
				{Name: "Balance", Value: balance.String()},
				{Name: "Ticker", Value: b.DB.Info().Ticker},
				{Name: "Account", Value: accountId},
			},
		},
	}
	return
}

func (b *Token) HandleTransfer(itemId, from string, params map[string]string) (res vmmSchema.Result) {
	_, from, err := utils.IDCheck(from)
	if err != nil {
		res.Error = schema.ErrInvalidFrom
		return
	}

	// Parse and validate recipient
	recipient, exists := params["Recipient"]
	if !exists {
		res.Error = schema.ErrMissingRecipient
		return
	}

	_, recipient, err = utils.IDCheck(recipient)
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

	// Execute transfer operation
	if err = b.Transfer(from, recipient, amount); err != nil {
		res.Error = err
		return
	}

	// Create debit notice for sender
	debitNotice := &vmmSchema.ResMessage{
		Target: from,
		Data:   "You transferred " + quantity + " to " + recipient,
		Tags: []goarSchema.Tag{
			{Name: "Ticker", Value: b.DB.Info().Ticker},
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
			{Name: "Ticker", Value: b.DB.Info().Ticker},
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

	res.Messages = []*vmmSchema.ResMessage{debitNotice, creditNotice}
	res.Cache = b.CacheChangeBalance(from, recipient)
	return
}

func (b *Token) handleMint(from string, params map[string]string) (res vmmSchema.Result) {
	// Check minting permission
	if from != b.DB.MintOwner() {
		res.Error = schema.ErrIncorrectOwner
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

	if b.DB.MaxSupply() != nil && b.DB.MaxSupply().Cmp(big.NewInt(0)) > 0 {
		if big.NewInt(0).Add(b.DB.GetTotalSupply(), amount).Cmp(b.DB.MaxSupply()) > 0 {
			res.Error = schema.ErrInsufficientMaxSupply
			return
		}
	}

	// Execute mint operation
	if err = b.Mint(recipient, amount); err != nil {
		res.Error = err
		return
	}

	// Create mint notice for owner
	ownerNotice := &vmmSchema.ResMessage{
		Target: from,
		Tags: []goarSchema.Tag{
			{Name: "Action", Value: "Mint-Notice"},
			{Name: "Recipient", Value: recipient},
			{Name: "Quantity", Value: quantity},
			{Name: "Ticker", Value: b.DB.Info().Ticker},
		},
	}

	// Create mint notice for recipient
	recipientNotice := &vmmSchema.ResMessage{
		Target: recipient,
		Tags: []goarSchema.Tag{
			{Name: "Action", Value: "Mint-Notice"},
			{Name: "Recipient", Value: recipient},
			{Name: "Quantity", Value: quantity},
			{Name: "Ticker", Value: b.DB.Info().Ticker},
		},
	}

	res.Messages = []*vmmSchema.ResMessage{ownerNotice, recipientNotice}
	res.Cache = map[string]string{}
	maps.Copy(res.Cache, b.CacheChangeBalance(recipient))
	maps.Copy(res.Cache, b.CacheTotalSupply())
	return
}
