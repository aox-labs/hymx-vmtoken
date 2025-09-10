package example

import (
	"fmt"
	"testing"

	"github.com/aox-labs/hymx-vmtoken/vmtoken"
	vmmSchema "github.com/hymatrix/hymx/vmm/schema"
)

func TestCrossChainMultiToken(t *testing.T) {
	// Create cross-chain multi token
	fmt.Println("=== Cross-Chain Multi Token Test ===")

	// Initialize environment parameters
	env := vmmSchema.Env{
		Id:    "multi-usd-token",
		AccId: "owner-address",
		Meta: vmmSchema.Meta{
			Action: "Spawn",
			Params: map[string]string{
				"Name":          "Multi-Chain USD Token",
				"Ticker":        "mUSD",
				"Decimals":      "6",
				"Logo":          "https://example.com/musd-logo.png",
				"BurnFees":      `{"ethereum":"500000","bsc":"40000","polygon":"30000"}`,
				"FeeRecipient":  "fee-recipient-address",
				"BurnProcessor": "burn-processor-address",
				"MintOwner":     "mint-owner-address",
			},
		},
	}

	// Create cross-chain multi token
	vm, err := vmtoken.SpawnCrossChainMultiToken(env)
	if err != nil {
		fmt.Printf("Failed to create token: %v\n", err)
		return
	}

	multiToken := vm.(*vmtoken.CrossChainMultiToken)
	fmt.Printf("Token created successfully: %s (%s)\n", multiToken.Info.Name, multiToken.Info.Ticker)

	// Test cross-chain Mint functionality
	fmt.Println("\n=== Testing Cross-Chain Mint Functionality ===")

	// User A cross-chains 100 USDC from Ethereum
	fmt.Println("User A cross-chains 100 USDC from Ethereum...")
	mintParams := map[string]string{
		"Recipient":       "user-a-address",
		"Quantity":        "100000000", // 100 USDC (6 decimals)
		"SourceChainType": "ethereum",
		"SourceTokenId":   "usdc-ethereum-0xa0b86c33c6b7c8c8c8c8c8c8c8c8c8c8c8c8c8c8",
		"X-MintTxHash":    "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
	}

	result, err := multiToken.HandleCrossChainMint("mint-owner-address", mintParams)
	if err != nil {
		fmt.Printf("Cross-chain Mint failed: %v\n", err)
	} else {
		fmt.Println("Cross-chain Mint successful!")
		for _, msg := range result.Messages {
			fmt.Printf("Message: %s -> %s\n", msg.Target, msg.Tags[0].Value)
		}
	}

	// User B cross-chains 500 USDT from BSC
	fmt.Println("\nUser B cross-chains 500 USDT from BSC...")
	mintParams2 := map[string]string{
		"Recipient":       "user-b-address",
		"Quantity":        "500000000", // 500 USDT (6 decimals)
		"SourceChainType": "bsc",
		"SourceTokenId":   "usdt-bsc-0x55d398326f99059ff775485246999027b3197955",
		"X-MintTxHash":    "0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
	}

	result2, err := multiToken.HandleCrossChainMint("mint-owner-address", mintParams2)
	if err != nil {
		fmt.Printf("Cross-chain Mint failed: %v\n", err)
	} else {
		fmt.Println("Cross-chain Mint successful!")
		for _, msg := range result2.Messages {
			fmt.Printf("Message: %s -> %s\n", msg.Target, msg.Tags[0].Value)
		}
	}

	// User A transfers 50 mUSD to User C
	fmt.Println("\n=== User A transfers 50 mUSD to User C ===")
	transferParams := map[string]string{
		"Recipient": "user-c-address",
		"Quantity":  "50000000", // 50 mUSD (6 decimals)
	}

	transferResult, err := multiToken.HandleTransfer("tx-id-1", "user-a-address", transferParams)
	if err != nil {
		fmt.Printf("Transfer failed: %v\n", err)
	} else {
		fmt.Println("Transfer successful!")
		for _, msg := range transferResult.Messages {
			fmt.Printf("Message: %s -> %s\n", msg.Target, msg.Tags[0].Value)
		}
	}

	// User C attempts to withdraw to BSC (should succeed)
	fmt.Println("\n=== User C withdraws to BSC ===")
	burnParams := map[string]string{
		"Recipient":     "user-c-bsc-address",
		"Quantity":      "50000000", // 50 mUSD
		"TargetTokenId": "usdt-bsc-0x55d398326f99059ff775485246999027b3197955",
	}

	burnResult, err := multiToken.HandleCrossChainBurn("user-c-address", burnParams)
	if err != nil {
		fmt.Printf("Withdrawal failed: %v\n", err)
	} else {
		fmt.Println("Withdrawal successful!")
		for _, msg := range burnResult.Messages {
			fmt.Printf("Message: %s -> %s\n", msg.Target, msg.Tags[0].Value)
		}
	}

	// User B attempts to withdraw to Ethereum (should succeed)
	fmt.Println("\n=== User B withdraws to Ethereum ===")
	burnParams2 := map[string]string{
		"Recipient":     "user-b-ethereum-address",
		"Quantity":      "200000000", // 200 mUSD
		"TargetTokenId": "usdc-ethereum-0xa0b86c33c6b7c8c8c8c8c8c8c8c8c8c8c8c8c8c8",
	}

	burnResult2, err := multiToken.HandleCrossChainBurn("user-b-address", burnParams2)
	if err != nil {
		fmt.Printf("Withdrawal failed: %v\n", err)
	} else {
		fmt.Println("Withdrawal successful!")
		for _, msg := range burnResult2.Messages {
			fmt.Printf("Message: %s -> %s\n", msg.Target, msg.Tags[0].Value)
		}
	}

	// Test error case: attempt to withdraw to non-existent token
	fmt.Println("\n=== Testing Error Case: Withdraw to Non-existent Token ===")
	burnParams3 := map[string]string{
		"Recipient":     "user-a-address",
		"Quantity":      "10000000", // 10 mUSD
		"TargetTokenId": "non-existent-token-id",
	}

	burnResult3, err := multiToken.HandleCrossChainBurn("user-a-address", burnParams3)
	if err != nil {
		fmt.Printf("Expected error: %v\n", err)
	} else {
		fmt.Println("Unexpected success!")
		for _, msg := range burnResult3.Messages {
			fmt.Printf("Message: %s -> %s\n", msg.Target, msg.Tags[0].Value)
		}
	}

	// Test error case: attempt to withdraw more than locked amount
	fmt.Println("\n=== Testing Error Case: Withdraw More Than Locked Amount ===")
	burnParams4 := map[string]string{
		"Recipient":     "user-a-address",
		"Quantity":      "1000000000", // 1000 mUSD (exceeds locked amount)
		"TargetTokenId": "usdc-ethereum-0xa0b86c33c6b7c8c8c8c8c8c8c8c8c8c8c8c8c8c8",
	}

	burnResult4, err := multiToken.HandleCrossChainBurn("user-a-address", burnParams4)
	if err != nil {
		fmt.Printf("Expected error: %v\n", err)
	} else {
		fmt.Println("Unexpected success!")
		for _, msg := range burnResult4.Messages {
			fmt.Printf("Message: %s -> %s\n", msg.Target, msg.Tags[0].Value)
		}
	}

	// Final state query
	fmt.Println("\n=== Final State Query ===")

	// Query user balances
	balanceResult, err := multiToken.HandleBalanceOf("user-a-address", map[string]string{})
	if err == nil {
		for _, msg := range balanceResult.Messages {
			fmt.Printf("User A balance: %s\n", msg.Data)
		}
	}

	balanceResult2, err := multiToken.HandleBalanceOf("user-b-address", map[string]string{})
	if err == nil {
		for _, msg := range balanceResult2.Messages {
			fmt.Printf("User B balance: %s\n", msg.Data)
		}
	}

	balanceResult3, err := multiToken.HandleBalanceOf("user-c-address", map[string]string{})
	if err == nil {
		for _, msg := range balanceResult3.Messages {
			fmt.Printf("User C balance: %s\n", msg.Data)
		}
	}

	// Query token information
	fmt.Println("\n=== Token Information Query ===")
	infoResult, err := multiToken.HandleInfo("owner-address")
	if err == nil {
		for _, msg := range infoResult.Messages {
			fmt.Printf("Token info: %s\n", msg.Tags[0].Value)
		}
	}

	fmt.Println("\n=== Test Completed ===")
}

// TestCrossChainMultiTokenErrorCases tests error cases
func TestCrossChainMultiTokenErrorCases(t *testing.T) {
	fmt.Println("=== Cross-Chain Multi Token Error Cases Test ===")

	// Create token
	env := vmmSchema.Env{
		Id:    "multi-usd-token-error",
		AccId: "owner-address",
		Meta: vmmSchema.Meta{
			Action: "Spawn",
			Params: map[string]string{
				"Name":          "Multi-Chain USD Token",
				"Ticker":        "mUSD",
				"Decimals":      "6",
				"BurnFees":      `{"ethereum":"500000","bsc":"40000"}`,
				"FeeRecipient":  "fee-recipient-address",
				"BurnProcessor": "burn-processor-address",
				"MintOwner":     "mint-owner-address",
			},
		},
	}

	vm, err := vmtoken.SpawnCrossChainMultiToken(env)
	if err != nil {
		fmt.Printf("Failed to create token: %v\n", err)
		return
	}

	multiToken := vm.(*vmtoken.CrossChainMultiToken)

	// Test 1: Unauthorized user attempts to Mint
	fmt.Println("\n--- Test 1: Unauthorized User Attempts to Mint ---")
	mintParams := map[string]string{
		"Recipient":       "user-address",
		"Quantity":        "100000000",
		"SourceChainType": "ethereum",
		"SourceTokenId":   "usdc-ethereum-0xa0b86c33c6b7c8c8c8c8c8c8c8c8c8c8c8c8c8c8",
	}

	_, err = multiToken.HandleCrossChainMint("unauthorized-user", mintParams)
	if err != nil {
		fmt.Printf("Expected error: %v\n", err)
	} else {
		fmt.Println("Unexpected success!")
	}

	// Test 2: Missing required parameters
	fmt.Println("\n--- Test 2: Missing Required Parameters ---")
	mintParams2 := map[string]string{
		"Recipient": "user-address",
		"Quantity":  "100000000",
		// Missing SourceChainType and SourceTokenId
	}

	_, err = multiToken.HandleCrossChainMint("mint-owner-address", mintParams2)
	if err != nil {
		fmt.Printf("Expected error: %v\n", err)
	} else {
		fmt.Println("Unexpected success!")
	}

	// Test 3: Invalid quantity format
	fmt.Println("\n--- Test 3: Invalid Quantity Format ---")
	mintParams3 := map[string]string{
		"Recipient":       "user-address",
		"Quantity":        "invalid-number",
		"SourceChainType": "ethereum",
		"SourceTokenId":   "usdc-ethereum-0xa0b86c33c6b7c8c8c8c8c8c8c8c8c8c8c8c8c8c8",
	}

	_, err = multiToken.HandleCrossChainMint("mint-owner-address", mintParams3)
	if err != nil {
		fmt.Printf("Expected error: %v\n", err)
	} else {
		fmt.Println("Unexpected success!")
	}

	// Test 4: Attempt to Burn non-existent token
	fmt.Println("\n--- Test 4: Attempt to Burn Non-existent Token ---")
	burnParams := map[string]string{
		"Recipient":     "user-address",
		"Quantity":      "10000000",
		"TargetTokenId": "non-existent-token",
	}

	_, err = multiToken.HandleCrossChainBurn("user-address", burnParams)
	if err != nil {
		fmt.Printf("Expected error: %v\n", err)
	} else {
		fmt.Println("Unexpected success!")
	}

	fmt.Println("\n=== Error Cases Test Completed ===")
}

// TestCrossChainMultiTokenComplexScenario tests complex scenarios
func TestCrossChainMultiTokenComplexScenario(t *testing.T) {
	fmt.Println("=== Cross-Chain Multi Token Complex Scenario Test ===")

	// Create token
	env := vmmSchema.Env{
		Id:    "multi-usd-token-complex",
		AccId: "owner-address",
		Meta: vmmSchema.Meta{
			Action: "Spawn",
			Params: map[string]string{
				"Name":          "Multi-Chain USD Token",
				"Ticker":        "mUSD",
				"Decimals":      "6",
				"BurnFees":      `{"ethereum":"500000","bsc":"40000","polygon":"30000","arbitrum":"60000"}`,
				"FeeRecipient":  "fee-recipient-address",
				"BurnProcessor": "burn-processor-address",
				"MintOwner":     "mint-owner-address",
			},
		},
	}

	vm, err := vmtoken.SpawnCrossChainMultiToken(env)
	if err != nil {
		fmt.Printf("Failed to create token: %v\n", err)
		return
	}

	multiToken := vm.(*vmtoken.CrossChainMultiToken)

	// Scenario 1: Multiple tokens cross-chain simultaneously
	fmt.Println("\n--- Scenario 1: Multiple Tokens Cross-chain Simultaneously ---")

	// Ethereum USDC
	mintParams1 := map[string]string{
		"Recipient":       "user-1",
		"Quantity":        "100000000", // 100 USDC
		"SourceChainType": "ethereum",
		"SourceTokenId":   "usdc-ethereum-0xa0b86c33c6b7c8c8c8c8c8c8c8c8c8c8c8c8c8c8",
		"X-MintTxHash":    "0x1111111111111111111111111111111111111111111111111111111111111111",
	}

	// BSC USDT
	mintParams2 := map[string]string{
		"Recipient":       "user-2",
		"Quantity":        "200000000", // 200 USDT
		"SourceChainType": "bsc",
		"SourceTokenId":   "usdt-bsc-0x55d398326f99059ff775485246999027b3197955",
		"X-MintTxHash":    "0x2222222222222222222222222222222222222222222222222222222222222222",
	}

	// Polygon USDC
	mintParams3 := map[string]string{
		"Recipient":       "user-3",
		"Quantity":        "150000000", // 150 USDC
		"SourceChainType": "polygon",
		"SourceTokenId":   "usdc-polygon-0x2791bca1f2de4661ed88a30c99a7a9449aa84174",
		"X-MintTxHash":    "0x3333333333333333333333333333333333333333333333333333333333333333",
	}

	// Execute multiple Mints
	mints := []map[string]string{mintParams1, mintParams2, mintParams3}

	for i, mintParams := range mints {
		_, err := multiToken.HandleCrossChainMint("mint-owner-address", mintParams)
		if err != nil {
			fmt.Printf("User %d Mint failed: %v\n", i+1, err)
		} else {
			fmt.Printf("User %d Mint successful!\n", i+1)
		}
	}

	// Scenario 2: Complex transfer network
	fmt.Println("\n--- Scenario 2: Complex Transfer Network ---")

	// user-1 transfers to user-4
	transfer1 := map[string]string{
		"Recipient": "user-4",
		"Quantity":  "30000000", // 30 mUSD
	}
	multiToken.HandleTransfer("tx-1", "user-1", transfer1)

	// user-2 transfers to user-5
	transfer2 := map[string]string{
		"Recipient": "user-5",
		"Quantity":  "50000000", // 50 mUSD
	}
	multiToken.HandleTransfer("tx-2", "user-2", transfer2)

	// user-3 transfers to user-1
	transfer3 := map[string]string{
		"Recipient": "user-1",
		"Quantity":  "20000000", // 20 mUSD
	}
	multiToken.HandleTransfer("tx-3", "user-3", transfer3)

	// Scenario 3: Multi-chain withdrawals
	fmt.Println("\n--- Scenario 3: Multi-chain Withdrawals ---")

	// user-4 withdraws to Ethereum
	burn1 := map[string]string{
		"Recipient":     "user-4-ethereum",
		"Quantity":      "30000000",
		"TargetTokenId": "usdc-ethereum-0xa0b86c33c6b7c8c8c8c8c8c8c8c8c8c8c8c8c8c8",
	}
	multiToken.HandleCrossChainBurn("user-4", burn1)

	// user-5 withdraws to BSC
	burn2 := map[string]string{
		"Recipient":     "user-5-bsc",
		"Quantity":      "50000000",
		"TargetTokenId": "usdt-bsc-0x55d398326f99059ff775485246999027b3197955",
	}
	multiToken.HandleCrossChainBurn("user-5", burn2)

	// Scenario 4: Verify final state
	fmt.Println("\n--- Scenario 4: Verify Final State ---")

	// Query all user balances
	for i, user := range []string{"user-1", "user-2", "user-3", "user-4", "user-5"} {
		balanceResult, err := multiToken.HandleBalanceOf(user, map[string]string{})
		if err == nil {
			for _, msg := range balanceResult.Messages {
				fmt.Printf("User %d (%s) balance: %s\n", i+1, user, msg.Data)
			}
		}
	}

	// Query token information
	infoResult, err := multiToken.HandleInfo("owner-address")
	if err == nil {
		fmt.Println("\nToken information:")
		for _, msg := range infoResult.Messages {
			for _, tag := range msg.Tags {
				if tag.Name == "SourceTokenChains" || tag.Name == "SourceLockAmounts" {
					fmt.Printf("%s: %s\n", tag.Name, tag.Value)
				}
			}
		}
	}

	fmt.Println("\n=== Complex Scenario Test Completed ===")
}
