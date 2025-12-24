package schema

const (
	VmTokenBasicModuleFormat           = "hymx.basic.token.0.0.1"
	VmTokenCrossChainMultiModuleFormat = "hymx.cross.chain.multi.token.0.0.1"
)

type Info struct {
	Id          string
	Name        string
	Ticker      string
	Decimals    string
	Logo        string
	Description string
}

type BasicCacheInfo struct {
	Name        string
	Ticker      string
	Decimals    string
	Logo        string
	Description string
	Owner       string
	MintOwner   string
	MaxSupply   string
}

type CrossChainCacheInfo struct {
	Name              string
	Ticker            string
	Decimals          string
	Logo              string
	Description       string
	Owner             string
	MintOwner         string
	BurnFees          string
	FeeRecipient      string
	BurnProcessor     string
	SourceTokenChains string
	SourceLockAmounts string
}
