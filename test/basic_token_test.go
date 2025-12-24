package test

import (
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

var (
	nameB     = "b token"
	tickerB   = "bToken"
	decimalsB = "12"
	maxSupply = "0"
	bToken    string
)

func init() {
	bToken = basicToken(nameB, tickerB, decimalsB, maxSupply)
	tokenInfo(bToken)
}

func Test_Basic_Token_Info(t *testing.T) {
	info := getBasicTokenInfoByCache(bToken)
	assert.Equal(t, nameB, info.Name)
	assert.Equal(t, tickerB, info.Ticker)
	assert.Equal(t, decimalsB, info.Decimals)
	assert.Equal(t, maxSupply, info.MaxSupply)
	assert.Equal(t, hysdk.Bundler.Address, info.Owner)
	assert.Equal(t, hysdk.Bundler.Address, info.MintOwner)
	assert.Equal(t, "", info.Logo)
	assert.Equal(t, "", info.Description)
}

func Test_Basic_Token_SetTokenOwner(t *testing.T) {
	newOwner := "VQsAJmeAXtL6LEsUQodQqdiTYKzbYtcAD1300ETsCAE"
	setTokenOwner(bToken, newOwner)
	info := getBasicTokenInfoByCache(bToken)
	assert.Equal(t, newOwner, info.Owner)
}

func Test_Basic_Token_Mint(t *testing.T) {
	recipient := "0xe688b84b23f322a994A53dbF8E15FA82CDB71127"
	quantity := "5000000"
	basicTokenMint(bToken, recipient, quantity)

	bal := getBalanceByCache(bToken, recipient)
	assert.Equal(t, quantity, bal.String())
	total := getTotalSupplyByCache(bToken)
	assert.Equal(t, quantity, total.String())
}

func Test_Basic_Token_Transfer(t *testing.T) {
	acc := hysdk.GetAddress()
	quantity := big.NewInt(100)
	basicTokenMint(bToken, acc, quantity.String())

	addr01 := "0x4838B106FCe9647Bdf1E7877BF73cE8B0BAD5f97"
	amt01 := big.NewInt(20)
	transfer(bToken, addr01, amt01.String())

	addr02 := "IrsYir2xZr3qixMRnTtKdX7d90maasV1iJL2AGiHhqQ"
	amt02 := big.NewInt(30)
	transfer(bToken, addr02, amt02.String())

	bal := getBalanceByCache(bToken, acc)
	assert.Equal(t, new(big.Int).Sub(new(big.Int).Sub(quantity, amt01), amt02), bal)

	bal = getBalanceByCache(bToken, addr01)
	assert.Equal(t, amt01, bal)

	bal = getBalanceByCache(bToken, addr02)
	assert.Equal(t, amt02, bal)
}
