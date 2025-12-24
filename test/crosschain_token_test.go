package test

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	nameC         = "c token"
	tickerC       = "cToken"
	decimalsC     = "6"
	burnFeeC      = "100"
	feeRecipientC = "0x4002ED1a1410aF1b4930cF6c479ae373dEbD6223"
	cToken        string
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
	assert.Equal(t, hysdk.Bundler.Address, info.FeeRecipient)
}

func Test_Cc_Token_CrossChainMint(t *testing.T) {
	recipient := "0xe688b84b23f322a994A53dbF8E15FA82CDB71127"
	quantity := "5000000"
	sourceChainType := "ethereum"
	sourceTokenId := "0xa0B86c33C6b7C8C8c8C8C8c8c8c8c8c8C8C8C8C8"
	mintTxHash := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"

	// Perform cross-chain mint
	crossChainMint(cToken, recipient, quantity, sourceChainType, sourceTokenId, mintTxHash)

	// Verify balance increased
	bal := getBalanceByCache(cToken, recipient)
	assert.Equal(t, quantity, bal.String())

	// Verify total supply increased
	total := getTotalSupplyByCache(cToken)
	assert.Equal(t, quantity, total.String())

	// Verify SourceTokenChains updated
	info := getCcTokenInfoByCache(cToken)
	sourceTokenChains := parseSourceTokenChains(info.SourceTokenChains)
	assert.Contains(t, sourceTokenChains, sourceTokenId)
	assert.Equal(t, sourceChainType, sourceTokenChains[sourceTokenId])

	// Verify SourceLockAmounts updated
	sourceLockAmounts := parseSourceLockAmounts(info.SourceLockAmounts)
	lockKey := sourceChainType + ":" + sourceTokenId
	lockAmount, exists := sourceLockAmounts[lockKey]
	assert.True(t, exists, "SourceLockAmount should exist for key: %s", lockKey)
	assert.Equal(t, quantity, lockAmount.String())
}

func Test_Cc_Token_CrossChainMint_DuplicateTxHash(t *testing.T) {
	recipient := "0xe688b84b23f322a994A53dbF8E15FA82CDB71127"
	quantity := "1000000"
	sourceChainType := "ethereum"
	sourceTokenId := "0x9f6d7a165C454008f2c8Bd72A21340b588F8C60a"
	mintTxHash := "0x9999999999999999999999999999999999999999999999999999999999999999"

	// First mint should succeed
	crossChainMint(cToken, recipient, quantity, sourceChainType, sourceTokenId, mintTxHash)

	// Verify first mint succeeded
	bal := getBalanceByCache(cToken, recipient)
	assert.GreaterOrEqual(t, bal.Cmp(mustParseBigInt(quantity)), 0)

	// Verify SourceTokenChains and SourceLockAmounts updated after first mint
	info := getCcTokenInfoByCache(cToken)
	sourceTokenChains := parseSourceTokenChains(info.SourceTokenChains)
	assert.Contains(t, sourceTokenChains, sourceTokenId)
	assert.Equal(t, sourceChainType, sourceTokenChains[sourceTokenId])

	sourceLockAmounts := parseSourceLockAmounts(info.SourceLockAmounts)
	lockKey := sourceChainType + ":" + sourceTokenId
	lockAmount, exists := sourceLockAmounts[lockKey]
	assert.True(t, exists)
	assert.Equal(t, quantity, lockAmount.String())
}

func Test_Cc_Token_CrossChainMint_MultipleMints(t *testing.T) {
	recipient := "0xe688b84b23f322a994A53dbF8E15FA82CDB71127"
	sourceChainType := "bsc"
	sourceTokenId := "0xdAC17F958D2ee523a2206206994597C13D831ec7"
	lockKey := sourceChainType + ":" + sourceTokenId

	// First mint
	quantity1 := "2000000"
	mintTxHash1 := "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	crossChainMint(cToken, recipient, quantity1, sourceChainType, sourceTokenId, mintTxHash1)

	// Verify after first mint
	info1 := getCcTokenInfoByCache(cToken)
	sourceLockAmounts1 := parseSourceLockAmounts(info1.SourceLockAmounts)
	lockAmount1, exists := sourceLockAmounts1[lockKey]
	assert.True(t, exists)
	assert.Equal(t, quantity1, lockAmount1.String())

	// Second mint with different tx hash (should accumulate)
	quantity2 := "3000000"
	mintTxHash2 := "0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	crossChainMint(cToken, recipient, quantity2, sourceChainType, sourceTokenId, mintTxHash2)

	// Verify after second mint - lock amount should be accumulated
	info2 := getCcTokenInfoByCache(cToken)
	sourceLockAmounts2 := parseSourceLockAmounts(info2.SourceLockAmounts)
	lockAmount2, exists := sourceLockAmounts2[lockKey]
	assert.True(t, exists)
	expectedTotal := new(big.Int).Add(mustParseBigInt(quantity1), mustParseBigInt(quantity2))
	assert.Equal(t, expectedTotal, lockAmount2)

	// Verify SourceTokenChains remains the same
	sourceTokenChains2 := parseSourceTokenChains(info2.SourceTokenChains)
	assert.Contains(t, sourceTokenChains2, sourceTokenId)
	assert.Equal(t, sourceChainType, sourceTokenChains2[sourceTokenId])
}

func Test_Cc_Token_Transfer(t *testing.T) {
	acc := hysdk.GetAddress()
	quantity := big.NewInt(100)
	sourceChainType := "base"
	sourceTokenId := "0x24AEE57BEe69D076fEdec0F6396dCC011D2Daeb4"
	mintTxHash1 := "0x1111111111111111111111111111111111111111111111111111111111111111"

	// Mint tokens via cross-chain mint
	crossChainMint(cToken, acc, quantity.String(), sourceChainType, sourceTokenId, mintTxHash1)

	// Verify SourceTokenChains and SourceLockAmounts updated after mint
	info := getCcTokenInfoByCache(cToken)
	sourceTokenChains := parseSourceTokenChains(info.SourceTokenChains)
	assert.Contains(t, sourceTokenChains, sourceTokenId)
	assert.Equal(t, sourceChainType, sourceTokenChains[sourceTokenId])

	sourceLockAmounts := parseSourceLockAmounts(info.SourceLockAmounts)
	lockKey := sourceChainType + ":" + sourceTokenId
	lockAmount, exists := sourceLockAmounts[lockKey]
	assert.True(t, exists)
	assert.Equal(t, quantity.String(), lockAmount.String())

	addr01 := "0x4838B106FCe9647Bdf1E7877BF73cE8B0BAD5f97"
	amt01 := big.NewInt(20)
	transfer(cToken, addr01, amt01.String())

	addr02 := "IrsYir2xZr3qixMRnTtKdX7d90maasV1iJL2AGiHhqQ"
	amt02 := big.NewInt(30)
	transfer(cToken, addr02, amt02.String())

	bal := getBalanceByCache(cToken, acc)
	assert.Equal(t, new(big.Int).Sub(new(big.Int).Sub(quantity, amt01), amt02), bal)

	bal = getBalanceByCache(cToken, addr01)
	assert.Equal(t, amt01, bal)

	bal = getBalanceByCache(cToken, addr02)
	assert.Equal(t, amt02, bal)
}

func Test_Cc_Token_CrossChainBurn(t *testing.T) {
	// Set up: First mint tokens via cross-chain mint to establish source token chain mapping
	targetTokenId := "0xE0554a476A092703abdB3Ef35c80e0D76d32939F"
	chainType := "ethereum"
	acc := hysdk.GetAddress()
	feeRecipient := "0x6d2e03b7EfFEae98BD302A9F836D0d6Ab0002766" // Different from acc
	mintQuantity := "10000"
	mintTxHash := "0x2222222222222222222222222222222222222222222222222222222222222222"

	// Set up burn fee for the chain type and fee recipient
	setCcTokenBurnFee(cToken, chainType, burnFeeC)
	setCcTokenParams(cToken, "", feeRecipient, "", "")

	// Mint tokens via cross-chain mint (this establishes the source token chain mapping)
	crossChainMint(cToken, acc, mintQuantity, chainType, targetTokenId, mintTxHash)

	// Verify SourceTokenChains and SourceLockAmounts after mint
	infoAfterMint := getCcTokenInfoByCache(cToken)
	sourceTokenChains := parseSourceTokenChains(infoAfterMint.SourceTokenChains)
	assert.Contains(t, sourceTokenChains, targetTokenId)
	assert.Equal(t, chainType, sourceTokenChains[targetTokenId])

	lockKey := chainType + ":" + targetTokenId
	sourceLockAmountsAfterMint := parseSourceLockAmounts(infoAfterMint.SourceLockAmounts)
	initialLockAmount, exists := sourceLockAmountsAfterMint[lockKey]
	assert.True(t, exists)
	assert.Equal(t, mintQuantity, initialLockAmount.String())

	// Get initial balances and total supply
	initialBal := getBalanceByCache(cToken, acc)
	initialFeeRecipientBal := getBalanceByCache(cToken, feeRecipient)
	initialTotal := getTotalSupplyByCache(cToken)

	// Perform cross-chain burn operation
	burnQuantity := "1000"
	crossChainBurn(cToken, burnQuantity, targetTokenId, "")

	// Get balances after burn
	finalBal := getBalanceByCache(cToken, acc)
	finalFeeRecipientBal := getBalanceByCache(cToken, feeRecipient)
	finalTotal := getTotalSupplyByCache(cToken)

	// Calculate expected values (burnQuantity - burnFee)
	burnAmount := mustParseBigInt(burnQuantity)
	burnFee := mustParseBigInt(burnFeeC)
	expectedBurn := new(big.Int).Sub(burnAmount, burnFee)

	// Verify sender balance decreased by burn amount (full amount deducted)
	expectedBal := new(big.Int).Sub(initialBal, burnAmount)
	assert.Equal(t, expectedBal, finalBal)

	// Verify fee recipient received the burn fee
	expectedFeeRecipientBal := new(big.Int).Add(initialFeeRecipientBal, burnFee)
	assert.Equal(t, expectedFeeRecipientBal, finalFeeRecipientBal)

	// Verify total supply decreased by net burn (quantity - fee)
	expectedTotal := new(big.Int).Sub(initialTotal, expectedBurn)
	assert.Equal(t, expectedTotal, finalTotal)

	// Verify SourceLockAmounts decreased after burn
	infoAfterBurn := getCcTokenInfoByCache(cToken)
	sourceLockAmountsAfterBurn := parseSourceLockAmounts(infoAfterBurn.SourceLockAmounts)
	finalLockAmount, exists := sourceLockAmountsAfterBurn[lockKey]
	assert.True(t, exists)
	expectedLockAmount := new(big.Int).Sub(initialLockAmount, expectedBurn)
	assert.Equal(t, expectedLockAmount, finalLockAmount)
}

func Test_Cc_Token_CrossChainBurn_WithRecipient(t *testing.T) {
	// Set up: Mint tokens via cross-chain mint
	targetTokenId := "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48"
	chainType := "ethereum"
	acc := hysdk.GetAddress()
	feeRecipient := "0x4838B106FCe9647Bdf1E7877BF73cE8B0BAD5f97" // Different from acc
	mintQuantity := "5000"
	mintTxHash := "0x3333333333333333333333333333333333333333333333333333333333333333"

	// Set up burn fee for the chain type and fee recipient
	setCcTokenBurnFee(cToken, chainType, burnFeeC)
	setCcTokenParams(cToken, "", feeRecipient, "", "")

	// Mint tokens via cross-chain mint
	crossChainMint(cToken, acc, mintQuantity, chainType, targetTokenId, mintTxHash)

	// Verify SourceTokenChains and SourceLockAmounts after mint
	infoAfterMint := getCcTokenInfoByCache(cToken)
	sourceTokenChains := parseSourceTokenChains(infoAfterMint.SourceTokenChains)
	assert.Contains(t, sourceTokenChains, targetTokenId)
	assert.Equal(t, chainType, sourceTokenChains[targetTokenId])

	lockKey := chainType + ":" + targetTokenId
	sourceLockAmountsAfterMint := parseSourceLockAmounts(infoAfterMint.SourceLockAmounts)
	initialLockAmount, exists := sourceLockAmountsAfterMint[lockKey]
	assert.True(t, exists)
	assert.Equal(t, mintQuantity, initialLockAmount.String())

	// Get initial balances and total supply
	initialBal := getBalanceByCache(cToken, acc)
	initialFeeRecipientBal := getBalanceByCache(cToken, feeRecipient)
	initialTotal := getTotalSupplyByCache(cToken)

	// Perform cross-chain burn with recipient hint
	burnQuantity := "2000"
	recipient := "0x6d2e03b7EfFEae98BD302A9F836D0d6Ab0002766"
	crossChainBurn(cToken, burnQuantity, targetTokenId, recipient)

	// Get balances after burn
	finalBal := getBalanceByCache(cToken, acc)
	finalFeeRecipientBal := getBalanceByCache(cToken, feeRecipient)
	finalTotal := getTotalSupplyByCache(cToken)

	// Calculate expected values
	burnAmount := mustParseBigInt(burnQuantity)
	burnFee := mustParseBigInt(burnFeeC)
	expectedBurn := new(big.Int).Sub(burnAmount, burnFee)

	// Verify sender balance decreased by burn amount (full amount deducted)
	expectedBal := new(big.Int).Sub(initialBal, burnAmount)
	assert.Equal(t, expectedBal, finalBal)

	// Verify fee recipient received the burn fee
	expectedFeeRecipientBal := new(big.Int).Add(initialFeeRecipientBal, burnFee)
	assert.Equal(t, expectedFeeRecipientBal, finalFeeRecipientBal)

	// Verify total supply decreased by net burn
	expectedTotal := new(big.Int).Sub(initialTotal, expectedBurn)
	assert.Equal(t, expectedTotal, finalTotal)

	// Verify SourceLockAmounts decreased after burn
	infoAfterBurn := getCcTokenInfoByCache(cToken)
	sourceLockAmountsAfterBurn := parseSourceLockAmounts(infoAfterBurn.SourceLockAmounts)
	finalLockAmount, exists := sourceLockAmountsAfterBurn[lockKey]
	assert.True(t, exists)
	expectedLockAmount := new(big.Int).Sub(initialLockAmount, expectedBurn)
	assert.Equal(t, expectedLockAmount, finalLockAmount)
}

func Test_Cc_Token_SetParams(t *testing.T) {
	newBurnFee := "200"
	newFeeRecipient := "0x6d2e03b7EfFEae98BD302A9F836D0d6Ab0002766"
	newBurnProcessor := "N-f6SD01hUFW22OBMy4Yy7EB2CF3Jrb_JOm7arVdaNc"
	newName := "Updated Cross-Chain Token"

	setCcTokenParams(cToken, newBurnFee, newFeeRecipient, newBurnProcessor, newName)

	info := getCcTokenInfoByCache(cToken)
	assert.Equal(t, newName, info.Name)
	assert.Equal(t, newFeeRecipient, info.FeeRecipient)
	assert.Equal(t, newBurnProcessor, info.BurnProcessor)
}

func Test_Cc_Token_SetTokenOwner(t *testing.T) {
	newOwner := "VQsAJmeAXtL6LEsUQodQqdiTYKzbYtcAD1300ETsCAE"
	setTokenOwner(cToken, newOwner)
	info := getCcTokenInfoByCache(cToken)
	assert.Equal(t, newOwner, info.Owner)
}
