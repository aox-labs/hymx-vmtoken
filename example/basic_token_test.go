package example

import (
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/hymatrix/hymx/sdk"
	goarSchema "github.com/permadao/goar/schema"
	"github.com/stretchr/testify/assert"
)

var (
	BASIC_MODULE = "9bQh650l10NZ7GHUvj1L_kIIiivp9Zj7kJNY3CLEcRM" // Basic token module format
	SCHEDULER    = "0x972AeD684D6f817e1b58AF70933dF1b4a75bfA51"  // local hymx node accId
	// SCHEDULER = "0xCD1Ef67a57a7c03BFB05F175Be10e3eC79821138" // permadao node accId
	hymxUrl = "http://127.0.0.1:8080" // local hymx node rpc
	// hymxUrl = "https://hymx.permadao.io" // local hymx node rpc

	testArKeyFile = "./test-keyfile.json" // generate cmd: npx -y @permaweb/wallet > test-keyfile.json
	hySdk         = sdk.New(hymxUrl, testArKeyFile)
)

// Test_Basic_Token_Spawn tests spawning a basic token
func Test_Basic_Token_Spawn(t *testing.T) {
	res, err := hySdk.SpawnAndWait(
		BASIC_MODULE,
		SCHEDULER,
		[]goarSchema.Tag{
			{Name: "Name", Value: "Basic Token"},
			{Name: "Ticker", Value: "bToken"},
			{Name: "Decimals", Value: "12"},
			{Name: "Logo", Value: "UkS-mdoiG8hcAClhKK8ch4ZhEzla0mCPDOix9hpdSFE"},
			{Name: "MintOwner", Value: hySdk.GetAddress()}, // Custom mint owner
		})
	assert.NoError(t, err)
	t.Log("Basic token PID: ", res.Id)
}

var basicTokenPid = "vfB8_goyfsagvrkN0kxv90gRlUT58tyhMTNodj4g4FY" // copy Test_Basic_Token_Spawn result tokenPid

// Test_Basic_Token_Info tests getting basic token info
func Test_Basic_Token_Info(t *testing.T) {
	tags := []goarSchema.Tag{
		{Name: "Action", Value: "Info"},
	}

	resp, err := hySdk.SendMessageAndWait(basicTokenPid, "", tags)
	assert.NoError(t, err)
	t.Log("Basic token info: ", resp.Message)
}

// Test_Basic_Token_Mint tests minting tokens
func Test_Basic_Token_Mint(t *testing.T) {
	tags := []goarSchema.Tag{
		{Name: "Action", Value: "Mint"},
		{Name: "Recipient", Value: hySdk.GetAddress()},
		{Name: "Quantity", Value: "50000000"},
	}

	resp, err := hySdk.SendMessageAndWait(basicTokenPid, "", tags)
	assert.NoError(t, err)
	t.Log("Mint result: ", resp)
}

// Test_Basic_Token_TotalSupply tests getting total supply
func Test_Basic_Token_TotalSupply(t *testing.T) {
	tags := []goarSchema.Tag{
		{Name: "Action", Value: "Total-Supply"},
	}

	resp, err := hySdk.SendMessageAndWait(basicTokenPid, "", tags)
	assert.NoError(t, err)
	t.Log("Total supply: ", resp.Message)
}

// Test_Basic_Token_Balance tests getting account balance
func Test_Basic_Token_Balance(t *testing.T) {
	tags := []goarSchema.Tag{
		{Name: "Action", Value: "Balance"},
		{Name: "Recipient", Value: hySdk.GetAddress()},
	}

	resp, err := hySdk.SendMessageAndWait(basicTokenPid, "", tags)
	assert.NoError(t, err)
	t.Log("Balance: ", resp.Message)
}

// Test_Basic_Token_Transfer tests transferring tokens
func Test_Basic_Token_Transfer(t *testing.T) {
	tokenId := basicTokenPid
	recipient := "0x6d2e03b7EfFEae98BD302A9F836D0d6Ab0002766" // must be evm address or ar address
	quantity := "100000"
	err := transfer(tokenId, recipient, quantity)
	assert.NoError(t, err)
}

// Test_Basic_Token_SetParams tests setting token parameters
func Test_Basic_Token_SetParams(t *testing.T) {
	tags := []goarSchema.Tag{
		{Name: "Action", Value: "Set-Params"},
		{Name: "Name", Value: "Updated Basic Token"},
		{Name: "Ticker", Value: "ubToken"},
	}

	resp, err := hySdk.SendMessageAndWait(basicTokenPid, "", tags)
	assert.NoError(t, err)
	t.Log("Set params result: ", resp)
}

// Test_Basic_Token_Cache_TokenInfo tests getting token info from cache
func Test_Basic_Token_Cache_TokenInfo(t *testing.T) {
	res, err := getCacheData(hymxUrl, basicTokenPid, "TokenInfo")
	assert.NoError(t, err)
	t.Log("Cached token info: ", res)
}

// Test_Basic_Token_Cache_BalanceOf tests getting specific account balance from cache
func Test_Basic_Token_Cache_BalanceOf(t *testing.T) {
	userAddress := "0x6d2e03b7EfFEae98BD302A9F836D0d6Ab0002766"
	res, err := getCacheData(hymxUrl, basicTokenPid, "Balances:"+userAddress)
	assert.NoError(t, err)
	t.Log("Cached balance for ", userAddress, ": ", res)
}

// Helper function for transfer operations
func transfer(tokenId, recipient, quantity string) error {
	tags := []goarSchema.Tag{
		{Name: "Action", Value: "Transfer"},
		{Name: "Recipient", Value: recipient},
		{Name: "Quantity", Value: quantity},
	}

	_, err := hySdk.SendMessage(tokenId, "", tags)
	return err
}

// Helper function to get cache data
func getCacheData(hymxUrl, pid, key string) (string, error) {
	url := fmt.Sprintf("%s/cache/%s/%s", hymxUrl, pid, key)

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}
