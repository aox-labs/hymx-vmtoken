package example

import (
	"fmt"
	"github.com/hymatrix/hymx/sdk"
	goarSchema "github.com/permadao/goar/schema"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"testing"
)

var (
	MODULE    = "EwJd55Pntp1Ld-rZ2hfYTfVmZUyYTro32DLnmXi4UCs"
	SCHEDULER = "0x972AeD684D6f817e1b58AF70933dF1b4a75bfA51" // local hymx node accId
	hymxUrl   = "http://127.0.0.1:8080"                      // local hymx node rpc

	testArKeyFile = "./test-keyFile.json" // generate cmd: npx -y @permaweb/wallet > test-keyfile.json
	hySdk         = sdk.New(hymxUrl, testArKeyFile)
)

func Test_Vm_Token_Pid(t *testing.T) {
	res, err := hySdk.SpawnAndWait(
		MODULE,
		SCHEDULER,
		[]goarSchema.Tag{
			{Name: "Name", Value: "a Token"},
			{Name: "Ticker", Value: "aToken"},
			{Name: "Decimals", Value: "12"},
			{Name: "Logo", Value: "UkS-mdoiG8hcAClhKK8ch4ZhEzla0mCPDOix9hpdSFE"},
		})
	assert.NoError(t, err)
	t.Log("tokenPid: ", res.Id)
}

var tokenPid = "Ruh1qnTOwO86FtwlZn3qxwStO7pW11sVNN95GKJykEM" // copy Test_Vm_Token_Pid result tokenPid

func Test_Vm_token_Info(t *testing.T) {
	tags := []goarSchema.Tag{
		{Name: "Action", Value: "Info"},
	}

	resp, err := hySdk.SendMessageAndWait(tokenPid, "", tags)
	assert.NoError(t, err)
	t.Log(resp.Message)
}

func Test_Vm_Token_Mint(t *testing.T) {
	tags := []goarSchema.Tag{
		{Name: "Action", Value: "Mint"},
		{Name: "Recipient", Value: hySdk.GetAddress()},
		{Name: "Quantity", Value: "50000000"},
	}

	resp, err := hySdk.SendMessageAndWait(tokenPid, "", tags)
	assert.NoError(t, err)
	t.Log(resp)
}

func Test_VmToken_Balances(t *testing.T) {
	tags := []goarSchema.Tag{
		{Name: "Action", Value: "Balances"},
	}

	resp, err := hySdk.SendMessageAndWait(tokenPid, "", tags)
	assert.NoError(t, err)
	t.Log(resp.Message)
}

func Test_VmToken_Transfer(t *testing.T) {
	tokenId := tokenPid
	recipient := "0x6d2e03b7EfFEae98BD302A9F836D0d6Ab0002766" // must be evm address or ar address
	qty := "100000"
	err := transfer(tokenId, recipient, qty)
	assert.NoError(t, err)
}

func transfer(tokenId, recipient, qty string) error {
	tags := []goarSchema.Tag{
		{Name: "Action", Value: "Transfer"},
		{Name: "Recipient", Value: recipient},
		{Name: "Quantity", Value: qty},
	}

	_, err := hySdk.SendMessage(tokenId, "", tags)
	return err
}

// get token info and balance from cache
func Test_Cache_Token_Info(t *testing.T) {
	res, err := getCacheData(hymxUrl, tokenPid, "TokenInfo")
	assert.NoError(t, err)
	t.Log(res)
}

func Test_Cache_Balances(t *testing.T) {
	res, err := getCacheData(hymxUrl, tokenPid, "Balances")
	assert.NoError(t, err)
	t.Log(res)
}

func Test_Cache_BalanceOf(t *testing.T) {
	userAddress := "0x6d2e03b7EfFEae98BD302A9F836D0d6Ab0002766"
	res, err := getCacheData(hymxUrl, tokenPid, "Balances:"+userAddress)
	assert.NoError(t, err)
	t.Log(res)
}

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
