package test

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	nameC     = "c token"
	tickerC   = "cToken"
	decimalsC = "6"
	cToken    string
)

func init() {
	cToken = crosschainToken(nameC, tickerC, decimalsC)
	tokenInfo(cToken)
}

func Test_Cc_Token_Info(t *testing.T) {
	info := getCcTokenInfoByCache(cToken)
	assert.Equal(t, nameC, info.Name)
	assert.Equal(t, tickerC, info.Ticker)
	assert.Equal(t, decimalsC, info.Decimals)
	assert.Equal(t, "", info.Logo)
	assert.Equal(t, "", info.Description)
	assert.Equal(t, hysdk.Bundler.Address, info.Owner)
	assert.Equal(t, hysdk.Bundler.Address, info.MintOwner)
}
