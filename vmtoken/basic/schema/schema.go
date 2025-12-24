package schema

type Info struct {
	Id          string
	Name        string
	Ticker      string
	Decimals    string
	Logo        string
	Description string
}

type CacheInfo struct {
	Name        string
	Ticker      string
	Decimals    string
	Logo        string
	Description string
	Owner       string
	MintOwner   string
	MaxSupply   string
}
