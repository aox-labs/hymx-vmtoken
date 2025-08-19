package vmtoken

import (
	"encoding/json"
	"github.com/aox-labs/hymx-vmtoken/vmtoken/schema"
	vmmSchema "github.com/hymatrix/hymx/vmm/schema"
	"github.com/hymatrix/hymx/vmm/utils"
	goarSchema "github.com/permadao/goar/schema"

	"math/big"
)

func (v *VmToken) handleInfo(from string) (res *vmmSchema.Result, err error) {
	info := v.info
	res = &vmmSchema.Result{
		Messages: []*vmmSchema.ResMessage{
			{
				Target: from,
				Tags: []goarSchema.Tag{
					{Name: "Name", Value: info.Name},
					{Name: "Ticker", Value: info.Ticker},
					{Name: "Logo", Value: info.Logo},
					{Name: "Denomination", Value: info.Decimals},
				},
			},
		},
	}
	return
}

func (v *VmToken) handleSetParams(from string, meta vmmSchema.Meta) (res *vmmSchema.Result, applyErr error) {
	var err error
	defer func() {
		if err != nil {
			// add Burn-Error Notice
			res.Messages = append(res.Messages, &vmmSchema.ResMessage{
				Target: from,
				Tags: []goarSchema.Tag{
					{Name: "Action", Value: "Set-Params-Error"},
					{Name: "Error", Value: err.Error()},
				},
			})
		}
	}()
	if from != v.owner {
		err = schema.ErrIncorrectOwner
		return
	}

	if meta.Params["Owner"] != "" {
		v.owner = meta.Params["Owner"]
	}

	if meta.Params["MintOwner"] != "" {
		v.mintOwner = meta.Params["MintOwner"]
	}

	if meta.Params["FeeRecipient"] != "" {
		v.feeRecipient = meta.Params["FeeRecipient"]
	}

	if meta.Params["Name"] != "" {
		v.info.Name = meta.Params["Name"]
	}

	if meta.Params["Ticker"] != "" {
		v.info.Ticker = meta.Params["Ticker"]
	}

	if meta.Params["Decimals"] != "" {
		v.info.Decimals = meta.Params["Decimals"]
	}

	if meta.Params["Logo"] != "" {
		v.info.Logo = meta.Params["Logo"]
	}

	if meta.Params["BurnFee"] != "" {
		burnFee, ok := new(big.Int).SetString(meta.Params["BurnFee"], 10)
		if ok {
			v.burnFee = burnFee
		}
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
		Cache: v.cacheTokenInfo(),
	}
	return
}

func (v *VmToken) handleTotalSupply(from string) (res *vmmSchema.Result, err error) {
	res = &vmmSchema.Result{
		Messages: []*vmmSchema.ResMessage{
			{
				Target: from,
				Data:   v.totalSupply.String(),
				Tags: []goarSchema.Tag{
					{Name: "Action", Value: "Total-Supply"},
					{Name: "Ticker", Value: v.info.Ticker},
				},
			},
		},
	}
	return
}

func (v *VmToken) handleBalances(from string) (res *vmmSchema.Result, err error) {
	balances := make(map[string]string)

	for k, v := range v.balances {
		balances[k] = v.String()
	}

	balancesJs, err := json.Marshal(balances)
	if err != nil {
		return
	}
	res = &vmmSchema.Result{
		Messages: []*vmmSchema.ResMessage{
			{
				Target: from,
				Data:   string(balancesJs),
				Tags: []goarSchema.Tag{
					{Name: "Ticker", Value: v.info.Ticker},
				},
			},
		},
	}
	return
}

func (v *VmToken) handleBalanceOf(from string, params map[string]string) (res *vmmSchema.Result, err error) {
	accid := from
	if recipient, ok := params["Recipient"]; ok {
		accid = recipient
	} else if target, ok := params["Target"]; ok {
		accid = target
	}
	bal := v.balanceOf(accid)

	res = &vmmSchema.Result{
		Messages: []*vmmSchema.ResMessage{
			{
				Target: from,
				Data:   bal.String(),
				Tags: []goarSchema.Tag{
					{Name: "Balance", Value: bal.String()},
					{Name: "Ticker", Value: v.info.Ticker},
					{Name: "Account", Value: accid},
				},
			},
		},
	}
	return
}

func (v *VmToken) handleTransfer(from string, params map[string]string) (res *vmmSchema.Result, applyErr error) {
	var err error
	defer func() {
		if err != nil {
			// add Burn-Error Notice
			res.Messages = append(res.Messages, &vmmSchema.ResMessage{
				Target: from,
				Tags: []goarSchema.Tag{
					{Name: "Action", Value: "Transfer-Error"},
					{Name: "Error", Value: err.Error()},
				},
			})
		}
	}()

	recipient, ok := params["Recipient"]
	if !ok {
		err = schema.ErrMissingRecipient
		return
	}
	qty, ok := params["Quantity"]
	if !ok {
		err = schema.ErrMissingQuantity
		return
	}
	amt, ok := new(big.Int).SetString(qty, 10)
	if !ok {
		err = schema.ErrInvalidQuantityFormat
		return
	}

	if err = v.transfer(from, recipient, amt); err != nil {
		return
	}

	debitNotice := &vmmSchema.ResMessage{
		Target: from,
		Data:   "You transferred " + qty + " to " + recipient,
		Tags: []goarSchema.Tag{
			{Name: "Ticker", Value: v.info.Ticker},
			{Name: "Action", Value: "Debit-Notice"},
			{Name: "Recipient", Value: recipient},
			{Name: "Quantity", Value: qty},
		},
	}
	creditNotice := &vmmSchema.ResMessage{
		Target: recipient,
		Data:   "You received " + qty + " from " + from,
		Tags: []goarSchema.Tag{
			{Name: "Ticker", Value: v.info.Ticker},
			{Name: "Action", Value: "Credit-Notice"},
			{Name: "Sender", Value: from},
			{Name: "Quantity", Value: qty},
		},
	}
	// udpateBalances :=
	res = &vmmSchema.Result{
		Messages: []*vmmSchema.ResMessage{debitNotice, creditNotice},
		Cache: v.cacheBalances(map[string]*big.Int{
			from:      v.balanceOf(from),
			recipient: v.balanceOf(recipient),
		}),
	}
	return
}

func (v *VmToken) handleMint(from string, params map[string]string) (res *vmmSchema.Result, applyErr error) {
	var err error
	defer func() {
		if err != nil {
			// add Burn-Error Notice
			res.Messages = append(res.Messages, &vmmSchema.ResMessage{
				Target: from,
				Tags: []goarSchema.Tag{
					{Name: "Action", Value: "Mint-Error"},
					{Name: "Error", Value: err.Error()},
				},
			})
		}
	}()
	recipient, ok := params["Recipient"]
	if !ok {
		err = schema.ErrMissingRecipient
		return
	}
	qty, ok := params["Quantity"]
	if !ok {
		err = schema.ErrMissingQuantity
		return
	}
	amt, ok := new(big.Int).SetString(qty, 10)
	if !ok {
		err = schema.ErrInvalidQuantityFormat
		return
	}
	if from != v.mintOwner {
		err = schema.ErrIncorrectOwner
		return
	}

	if err = v.mint(recipient, amt); err != nil {
		return
	}

	debitNotice := &vmmSchema.ResMessage{
		Target: from,
		Tags: []goarSchema.Tag{
			{Name: "Action", Value: "Mint-Notice"},
			{Name: "Recipient", Value: recipient},
			{Name: "Quantity", Value: qty},
			{Name: "Ticker", Value: v.info.Ticker},
		},
	}
	creditNotice := &vmmSchema.ResMessage{
		Target: recipient,
		Tags: []goarSchema.Tag{
			{Name: "Action", Value: "Mint-Notice"},
			{Name: "Recipient", Value: recipient},
			{Name: "Quantity", Value: qty},
			{Name: "Ticker", Value: v.info.Ticker},
		},
	}
	res = &vmmSchema.Result{
		Messages: []*vmmSchema.ResMessage{debitNotice, creditNotice},
		Cache: v.cacheBalances(map[string]*big.Int{
			recipient: v.balanceOf(recipient),
		}),
	}
	return
}

func (v *VmToken) handleBurn(from string, params map[string]string) (res *vmmSchema.Result, applyErr error) {
	recipient, ok := params["Recipient"]
	if !ok {
		recipient, ok = params["X-Recipient"]
		if !ok {
			recipient = from
		}
	}

	var err error
	defer func() {
		if err != nil {
			// add Burn-Error Notice
			res.Messages = append(res.Messages, &vmmSchema.ResMessage{
				Target: from,
				Tags: []goarSchema.Tag{
					{Name: "Action", Value: "Burn-Error"},
					{Name: "Error", Value: err.Error()},
				},
			})
		}
	}()

	_, _, err = utils.IDCheck(recipient)
	if err != nil {
		return
	}

	qty, ok := params["Quantity"]
	if !ok {
		err = schema.ErrMissingQuantity
		return
	}
	amt, ok := new(big.Int).SetString(qty, 10)
	if !ok {
		err = schema.ErrInvalidQuantityFormat
		return
	}

	if err = v.burn(from, amt); err != nil {
		return
	}
	creditNotice := &vmmSchema.ResMessage{
		Target: from,
		Tags: []goarSchema.Tag{
			{Name: "Action", Value: "Burn-Notice"},
			{Name: "Sender", Value: from},
			{Name: "X-Recipient", Value: recipient},
			{Name: "Quantity", Value: new(big.Int).Sub(amt, v.burnFee).String()},
			{Name: "Ticker", Value: v.info.Ticker},
			{Name: "TokenId", Value: v.info.Id},
			{Name: "Fee", Value: v.burnFee.String()},
			{Name: "FeeRecipient", Value: v.feeRecipient},
		},
	}
	res = &vmmSchema.Result{
		Messages: []*vmmSchema.ResMessage{creditNotice},
		Cache: v.cacheBalances(map[string]*big.Int{
			from:           v.balanceOf(from),
			v.feeRecipient: v.balanceOf(v.feeRecipient),
		}),
	}
	return
}

func (v *VmToken) burn(from string, amount *big.Int) (err error) {
	if amount.Cmp(v.burnFee) < 0 {
		err = schema.ErrIncorrectQuantity
		return
	}

	if err = v.sub(from, amount); err != nil {
		return
	}

	if err = v.add(v.feeRecipient, v.burnFee); err != nil {
		return
	}

	v.totalSupply = new(big.Int).Sub(v.totalSupply, new(big.Int).Sub(amount, v.burnFee))
	return
}

func (v *VmToken) mint(to string, amount *big.Int) (err error) {
	_, to, err = utils.IDCheck(to)
	if err != nil {
		return
	}
	if err = v.add(to, amount); err != nil {
		log.Error("mint: token add failed", "err", err)
		return
	}

	v.totalSupply = new(big.Int).Add(v.totalSupply, amount)
	return
}

func (v *VmToken) transfer(from, to string, amount *big.Int) (err error) {
	_, to, err = utils.IDCheck(to)
	if err != nil {
		return
	}
	if err = v.sub(from, amount); err != nil {
		log.Error("transfer: token sub failed", "err", err)
		return
	}
	if err = v.add(to, amount); err != nil {
		log.Error("transfer: token add failed", "err", err)
		return
	}
	return nil

}

func (v *VmToken) sub(accId string, amount *big.Int) error {
	// if amount == 0, then return nil
	if amount.Cmp(big.NewInt(0)) == 0 {
		return nil
	}

	bal := v.balanceOf(accId)

	if bal.Cmp(amount) < 0 {
		return schema.ErrInsufficientBalance
	}
	newBal := new(big.Int).Sub(bal, amount)
	v.balances[accId] = newBal

	return nil
}

func (v *VmToken) add(accId string, amount *big.Int) error {
	// if amount == 0, then return nil
	if amount.Cmp(big.NewInt(0)) == 0 {
		return nil
	}

	bal := v.balanceOf(accId)

	newBal := new(big.Int).Add(bal, amount)
	v.balances[accId] = newBal

	return nil
}

func (v *VmToken) balanceOf(accId string) *big.Int {
	bal, ok := v.balances[accId]
	if !ok {
		bal = big.NewInt(0)
	}
	return bal
}
