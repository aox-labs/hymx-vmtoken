package test

import (
	"encoding/json"
	"fmt"
	bSchema "github.com/aox-labs/hymx-vmtoken/vmtoken/basic/schema"
	cSchema "github.com/aox-labs/hymx-vmtoken/vmtoken/crosschain/schema"
	goarSchema "github.com/permadao/goar/schema"
	"github.com/tidwall/gjson"
	"math/big"
)

var (
	TAxModule      = "1i03Vpe8DljkUMBEEEvR0VmbJjvgZtP_ytZdThkVSMw"
	RegistryModule = "MVTil0kn5SRiJELW7W2jLZ6cBr3QUGj1nJ67I2Wi4Ps"

	BasicTokenMod = "9bQh650l10NZ7GHUvj1L_kIIiivp9Zj7kJNY3CLEcRM" // Token token module format
	CcTokenMod    = "QW_l2HiEgurKA-_gxe5JXYYuqpkFwyps_V1RjGO86-c" // Cross-chain token module format

)

func basicToken(name, symbol, decimals, maxSupply string) string {
	res, err := hysdk.SpawnAndWait(BasicTokenMod, nodeInfo.Node.AccId,
		[]goarSchema.Tag{
			{Name: "Name", Value: name},
			{Name: "Ticker", Value: symbol},
			{Name: "Decimals", Value: decimals},
			{Name: "MaxSupply", Value: maxSupply},
		})
	if err != nil {
		panic(err)
	}
	return res.Id
}

func crosschainToken(name, symbol, decimals string) string {
	res, err := hysdk.SpawnAndWait(CcTokenMod, nodeInfo.Node.AccId,
		[]goarSchema.Tag{
			{Name: "Name", Value: name},
			{Name: "Ticker", Value: symbol},
			{Name: "Decimals", Value: decimals},
		})
	if err != nil {
		panic(err)
	}
	return res.Id
}

func tokenInfo(tokenId string) {
	tags := []goarSchema.Tag{
		{Name: "Action", Value: "Info"},
	}

	resp, err := hysdk.SendMessageAndWait(tokenId, "", tags)
	if err != nil {
		panic(err)
	}
	vmErr := gjson.Get(resp.Message, "Error").Str
	if vmErr != "" {
		panic(vmErr)
	}
	fmt.Println("get token info complete")
}

func setTokenOwner(tokenId, newOwner string) {
	tags := []goarSchema.Tag{
		{Name: "Action", Value: "Set-Params"},
		{Name: "TokenOwner", Value: newOwner},
	}

	resp, err := hysdk.SendMessageAndWait(tokenId, "", tags)
	if err != nil {
		panic(err)
	}
	vmErr := gjson.Get(resp.Message, "Error").Str
	if vmErr != "" {
		panic(vmErr)
	}
}

func basicTokenMint(tokenId string, recipient string, quantity string) {
	tags := []goarSchema.Tag{
		{Name: "Action", Value: "Mint"},
		{Name: "Recipient", Value: recipient},
		{Name: "Quantity", Value: quantity},
	}

	resp, err := hysdk.SendMessageAndWait(tokenId, "", tags)
	if err != nil {
		panic(err)
	}
	vmErr := gjson.Get(resp.Message, "Error").Str
	if vmErr != "" {
		panic(vmErr)
	}
}

func transfer(tokenId, to, amt string) {
	_, err := hysdk.SendMessageAndWait(tokenId, "",
		[]goarSchema.Tag{
			{Name: "Action", Value: "Transfer"},
			{Name: "Recipient", Value: to},
			{Name: "Quantity", Value: amt},
		},
	)
	if err != nil {
		panic(err)
	}
}

func getBasicTokenInfoByCache(tokenId string) bSchema.CacheInfo {
	infoJs, err := hysdk.Client.GetCache(tokenId, "token-info")
	if err != nil {
		panic(fmt.Sprintf("failed to get info: %v", err))
	}

	info := bSchema.CacheInfo{}
	if err = json.Unmarshal([]byte(infoJs), &info); err != nil {
		panic(fmt.Sprintf("failed to unmarshal infoJs: %v", err))
	}

	return info
}

func getCcTokenInfoByCache(tokenId string) cSchema.CacheInfo {
	infoJs, err := hysdk.Client.GetCache(tokenId, "token-info")
	if err != nil {
		panic(fmt.Sprintf("failed to get amm info: %v", err))
	}

	info := cSchema.CacheInfo{}
	if err = json.Unmarshal([]byte(infoJs), &info); err != nil {
		panic(fmt.Sprintf("failed to unmarshal infoJs: %v", err))
	}

	return info
}

func getBalanceByCache(tokenId, accId string) *big.Int {
	bal, err := hysdk.Client.GetCache(tokenId, "balances:"+accId)
	if err != nil {
		panic(fmt.Sprintf("failed to get balance: %v", err))
	}
	return mustParseBigInt(bal)
}

func getTotalSupplyByCache(tokenId string) *big.Int {
	amt, err := hysdk.Client.GetCache(tokenId, "total-supply")
	if err != nil {
		panic(fmt.Sprintf("failed to get total supply: %v", err))
	}
	return mustParseBigInt(amt)
}

func mustParseBigInt(s string) *big.Int {
	if s == "" {
		s = "0"
	}
	val, ok := new(big.Int).SetString(s, 10)
	if !ok {
		panic(fmt.Sprintf("invalid big.Int string: %s", s))
	}
	return val
}
