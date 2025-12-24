package basic

import (
	"github.com/aox-labs/hymx-vmtoken/schema"
	"github.com/hymatrix/hymx/vmm/utils"
	"math/big"
)

// Core token operations
func (b *Token) Mint(to string, amount *big.Int) (err error) {
	// Validate and normalize recipient address
	_, to, err = utils.IDCheck(to)
	if err != nil {
		return
	}

	// Add tokens to recipient
	if err = b.Add(to, amount); err != nil {
		return
	}

	// Increase total supply
	b.DB.SetTotalSupply(new(big.Int).Add(b.DB.GetTotalSupply(), amount))
	return
}

func (b *Token) Transfer(from, to string, amount *big.Int) (err error) {
	// Validate and normalize recipient address
	_, to, err = utils.IDCheck(to)
	if err != nil {
		return
	}

	// Deduct tokens from sender
	if err = b.Sub(from, amount); err != nil {
		return
	}

	// Add tokens to recipient
	if err = b.Add(to, amount); err != nil {
		return
	}

	return nil
}

func (b *Token) Sub(accId string, amount *big.Int) error {
	// Skip operation if amount is zero
	if amount.Cmp(big.NewInt(0)) == 0 {
		return nil
	}

	// Check sufficient balance
	currentBalance, err := b.DB.BalanceOf(accId)
	if err != nil {
		return err
	}
	if currentBalance.Cmp(amount) < 0 {
		return schema.ErrInsufficientBalance
	}

	// Calculate new balance and update
	return b.DB.UpdateBalance(accId, new(big.Int).Sub(currentBalance, amount))
}

func (b *Token) Add(accId string, amount *big.Int) error {
	// Skip operation if amount is zero
	if amount.Cmp(big.NewInt(0)) == 0 {
		return nil
	}

	// Get current balance and calculate new balance
	currentBalance, err := b.DB.BalanceOf(accId)
	if err != nil {
		return err
	}

	return b.DB.UpdateBalance(accId, new(big.Int).Add(currentBalance, amount))
}
