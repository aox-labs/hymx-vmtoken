package example

import (
	"testing"

	goarSchema "github.com/permadao/goar/schema"
	"github.com/stretchr/testify/assert"
)

var (
	CROSSCHAIN_MODULE = "QW_l2HiEgurKA-_gxe5JXYYuqpkFwyps_V1RjGO86-c" // Cross-chain token module format
)

// Test_CrossChain_Token_Spawn tests spawning a cross-chain token
func Test_CrossChain_Token_Spawn(t *testing.T) {
	res, err := hySdk.SpawnAndWait(
		CROSSCHAIN_MODULE,
		SCHEDULER,
		[]goarSchema.Tag{
			{Name: "Name", Value: "Cross-Chain Token"},
			{Name: "Ticker", Value: "ccToken"},
			{Name: "Decimals", Value: "18"},
			{Name: "Logo", Value: "UkS-mdoiG8hcAClhKK8ch4ZhEzla0mCPDOix9hpdSFE"},
			{Name: "MintOwner", Value: "0xf4FfBA30E4A427E4c743A9142dEDba284487c75F"},      // Custom mint owner
			{Name: "BurnFee", Value: "100"},                                               // Cross-chain specific: burn fee
			{Name: "FeeRecipient", Value: "0xf4FfBA30E4A427E4c743A9142dEDba284487c75F"},   // Cross-chain specific: fee recipient
			{Name: "BurnProcessor", Value: "MysFttDUI1YJKcFwYIyqVWGfFGnetcCp_5TGjdhVgS4"}, // Cross-chain specific: burn processor
		})
	assert.NoError(t, err)
	t.Log("Cross-chain token PID: ", res.Id)
}

var crosschainTokenPid = "Ix2df23IjisdYxrTO7xrWgM0hwCU9DPgvjZPZrx1DF4" // copy Test_CrossChain_Token_Spawn result tokenPid

// ===== Basic Token Functionality Tests (same as basic_token_test.go) =====

// Test_CrossChain_Token_Info tests getting cross-chain token info
func Test_CrossChain_Token_Info(t *testing.T) {
	tags := []goarSchema.Tag{
		{Name: "Action", Value: "Info"},
	}

	resp, err := hySdk.SendMessageAndWait(crosschainTokenPid, "", tags)
	assert.NoError(t, err)
	t.Log("Cross-chain token info: ", resp.Message)
}

// Test_CrossChain_Token_Mint tests minting tokens
func Test_CrossChain_Token_Mint(t *testing.T) {
	tags := []goarSchema.Tag{
		{Name: "Action", Value: "Mint"},
		{Name: "Recipient", Value: hySdk.GetAddress()},
		{Name: "Quantity", Value: "50000000"},
	}

	resp, err := hySdk.SendMessageAndWait(crosschainTokenPid, "", tags)
	assert.NoError(t, err)
	t.Log("Mint result: ", resp)
}

// Test_CrossChain_Token_TotalSupply tests getting total supply
func Test_CrossChain_Token_TotalSupply(t *testing.T) {
	tags := []goarSchema.Tag{
		{Name: "Action", Value: "Total-Supply"},
	}

	resp, err := hySdk.SendMessageAndWait(crosschainTokenPid, "", tags)
	assert.NoError(t, err)
	t.Log("Total supply: ", resp.Message)
}

// Test_CrossChain_Token_Balance tests getting account balance
func Test_CrossChain_Token_Balance(t *testing.T) {
	tags := []goarSchema.Tag{
		{Name: "Action", Value: "Balance"},
		{Name: "Recipient", Value: hySdk.GetAddress()},
	}

	resp, err := hySdk.SendMessageAndWait(crosschainTokenPid, "", tags)
	assert.NoError(t, err)
	t.Log("Balance: ", resp.Message)
}

// Test_CrossChain_Token_Transfer tests transferring tokens
func Test_CrossChain_Token_Transfer(t *testing.T) {
	tokenId := crosschainTokenPid
	recipient := "0x6d2e03b7EfFEae98BD302A9F836D0d6Ab0002766" // must be evm address or ar address
	quantity := "100000"
	err := transfer(tokenId, recipient, quantity)
	assert.NoError(t, err)
}

// Test_CrossChain_Token_SetParams tests setting token parameters
func Test_CrossChain_Token_SetParams(t *testing.T) {
	tags := []goarSchema.Tag{
		{Name: "Action", Value: "Set-Params"},
		{Name: "Name", Value: "Updated Cross-Chain Token"},
		{Name: "Ticker", Value: "uccToken"},
		{Name: "BurnFee", Value: "200"},                   // Cross-chain specific: update burn fee
		{Name: "FeeRecipient", Value: hySdk.GetAddress()}, // Cross-chain specific: update fee recipient
	}

	resp, err := hySdk.SendMessageAndWait(crosschainTokenPid, "", tags)
	assert.NoError(t, err)
	t.Log("Set params result: ", resp)
}

// ===== Cross-Chain Specific Functionality Tests =====

// Test_CrossChain_Token_Burn tests burning tokens with fee
func Test_CrossChain_Token_Burn(t *testing.T) {
	tags := []goarSchema.Tag{
		{Name: "Action", Value: "Burn"},
		{Name: "Quantity", Value: "1000"},
		{Name: "X-Recipient", Value: "0x6d2e03b7EfFEae98BD302A9F836D0d6Ab0002766"}, // Cross-chain recipient hint
	}

	resp, err := hySdk.SendMessageAndWait(crosschainTokenPid, "", tags)
	assert.NoError(t, err)
	t.Log("Burn result: ", resp)
}

// Test_CrossChain_Token_Burn_WithFee tests burning tokens and verifying fee deduction
func Test_CrossChain_Token_Burn_WithFee(t *testing.T) {
	// First get current balance
	balanceTags := []goarSchema.Tag{
		{Name: "Action", Value: "Balance"},
		{Name: "Recipient", Value: hySdk.GetAddress()},
	}

	balanceResp, err := hySdk.SendMessageAndWait(crosschainTokenPid, "", balanceTags)
	assert.NoError(t, err)
	t.Log("Balance before burn: ", balanceResp.Message)

	// Perform burn operation
	burnTags := []goarSchema.Tag{
		{Name: "Action", Value: "Burn"},
		{Name: "Quantity", Value: "5000"},
	}

	burnResp, err := hySdk.SendMessageAndWait(crosschainTokenPid, "", burnTags)
	assert.NoError(t, err)
	t.Log("Burn operation result: ", burnResp)

	// Get balance after burn
	balanceAfterResp, err := hySdk.SendMessageAndWait(crosschainTokenPid, "", balanceTags)
	assert.NoError(t, err)
	t.Log("Balance after burn: ", balanceAfterResp.Message)
}

// Test_CrossChain_Token_BurnProcessor tests setting and using custom burn processor
func Test_CrossChain_Token_BurnProcessor(t *testing.T) {
	// Set custom burn processor
	setProcessorTags := []goarSchema.Tag{
		{Name: "Action", Value: "Set-Params"},
		{Name: "BurnProcessor", Value: "0x6d2e03b7EfFEae98BD302A9F836D0d6Ab0002766"},
	}

	resp, err := hySdk.SendMessageAndWait(crosschainTokenPid, "", setProcessorTags)
	assert.NoError(t, err)
	t.Log("Set burn processor result: ", resp)

	// Perform burn operation to test custom processor
	burnTags := []goarSchema.Tag{
		{Name: "Action", Value: "Burn"},
		{Name: "Quantity", Value: "1000"},
	}

	burnResp, err := hySdk.SendMessageAndWait(crosschainTokenPid, "", burnTags)
	assert.NoError(t, err)
	t.Log("Burn with custom processor result: ", burnResp)
}

// Test_CrossChain_Token_Cache_TokenInfo tests getting cross-chain token info from cache
func Test_CrossChain_Token_Cache_TokenInfo(t *testing.T) {
	res, err := getCacheData(hymxUrl, crosschainTokenPid, "TokenInfo")
	assert.NoError(t, err)
	t.Log("Cached cross-chain token info: ", res)
}

// Test_CrossChain_Token_Cache_BalanceOf tests getting specific account balance from cache
func Test_CrossChain_Token_Cache_BalanceOf(t *testing.T) {
	userAddress := "0x6d2e03b7EfFEae98BD302A9F836D0d6Ab0002766"
	res, err := getCacheData(hymxUrl, crosschainTokenPid, "Balances:"+userAddress)
	assert.NoError(t, err)
	t.Log("Cached balance for ", userAddress, ": ", res)
}

// Test_CrossChain_Token_Cache_FeeRecipient tests getting fee recipient balance from cache
func Test_CrossChain_Token_Cache_FeeRecipient(t *testing.T) {
	feeRecipient := hySdk.GetAddress()
	res, err := getCacheData(hymxUrl, crosschainTokenPid, "Balances:"+feeRecipient)
	assert.NoError(t, err)
	t.Log("Cached balance for fee recipient ", feeRecipient, ": ", res)
}
