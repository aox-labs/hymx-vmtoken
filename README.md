## hymx-vmtoken

A generic Token VM implementation for Hymatrix Hymx (ModuleFormat: `aox.token.0.0.1`).

This project provides:
- vmToken standard and action interfaces: mint, transfer, burn, query info and balances, and parameter updates.
- Server integration: mount vmToken as a VM on a Hymx node and expose token parameter cache reading.
- SDK examples: use `github.com/hymatrix/hymx/sdk` to deploy and interact with the token (see examples for end-to-end tests).

Best suited for developers who want to issue and manage fungible tokens on Hymx, or implement custom token logic by reference.

### Module Format
- `schema.VmTokenModuleFormat = "aox.token.0.0.1"`

When a Hymx node starts, it binds the ModuleFormat to `SpawnVmToken`:
- `s.Mount("aox.token.0.0.1", vmtoken.SpawnVmToken)`

### Spawn Parameters
Provide the following tags when instantiating a vmToken:
- Name: token name (required)
- Ticker: token symbol (required)
- Decimals: precision (required, decimal string)
- Logo: Arweave resource identifier for the logo (optional)

After spawning:
- `owner` is the spawner account (`env.AccId`).
- Initial state: `totalSupply = 0`, `balances = {}`, `burnFee = 0`, `feeRecipient = owner`.

### Actions and Parameters

- Info
  - Returns basic token info.
  - Params: none.
  - Return tags: `Name`, `Ticker`, `Logo`, `Denomination(=Decimals)`.
  - On first call, initializes and writes cache (see Cache Keys).

- Set-Params (owner-only)
  - Updates token and account parameters.
  - Any of: `Owner`, `FeeRecipient`, `Name`, `Ticker`, `Decimals`, `Logo`, `BurnFee` (decimal string).
  - Return tags: `Set-Params-Notice = success`.
  - Cache: refreshes `TokenInfo`.

- Total-Supply / TotalSupply
  - Queries total supply.
  - Params: none.
  - Return: message Data is the total as a decimal string; tags include `Action=Total-Supply`, `Ticker`.

- Balances
  - Returns a snapshot of all balances.
  - Params: none.
  - Return: message Data is a JSON `map[string]string` (address -> balance); tags include `Ticker`.

- Balance
  - Queries balance for an account.
  - Params: either `Recipient` or `Target`. If neither provided, defaults to the caller account.
  - Return tags: `Balance`, `Ticker`, `Account` (queried account). Data is the balance string.

- Transfer
  - Transfers from caller to `Recipient`.
  - Params: `Recipient` (required; supports EVM/Arweave addresses), `Quantity` (required; decimal string).
  - Validation: quantity format; sufficient balance; `Recipient` passes ID check.
  - Return: two messages:
    - Caller’s `Debit-Notice` with tags `Ticker`, `Action=Debit-Notice`, `Recipient`, `Quantity`.
    - Recipient’s `Credit-Notice` with tags `Ticker`, `Action=Credit-Notice`, `Sender`, `Quantity`.
  - Cache: updates `Balances:<sender>`, `Balances:<recipient>`, `Balances`, `TotalSupply`.

- Mint (owner-only)
  - Mints `Quantity` to `Recipient`.
  - Params: `Recipient`, `Quantity` (decimal string).
  - Return: two `Mint-Notice` messages (to owner and recipient), tags include `Recipient`, `Quantity`, `Ticker`.
  - Cache: updates `totalSupply` and affected account caches.

- Burn
  - Burns `Quantity` from caller; `burnFee` is transferred to `feeRecipient`. Net burn = `Quantity - burnFee`.
  - Params: `Quantity` (required, decimal string); `Recipient` or `X-Recipient` (optional; defaults to `from`).
  - Validation: `Quantity >= burnFee`; address passes ID check; sufficient balance.
  - Return: one `Burn-Notice` with tags `Sender`, `X-Recipient`, `Quantity` (=net burn), `Ticker`, `TokenId`, `Fee`, `FeeRecipient`.
  - Cache: updates `totalSupply` (minus net burn) and caches for `from` and `feeRecipient`.

### Cache Keys (via Hymx node HTTP)
- `TokenInfo`: stringified JSON with `Name`, `Ticker`, `Denomination`, `Logo`, `Owner`, `BurnFee`, `FeeRecipient`.
- `TotalSupply`: total supply as a string.
- `Balances`: stringified JSON (address -> balance string).
- `Balances:<Account>`: balance string of an account.

Read example (see `example/vmtoken_test.go`):
```go
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
```

## Run the Service (CLI)

### Config example (`cmd/config.yaml`)
- port: listen address, e.g. `:8080`
- ginMode: `debug` or `release`
- redisURL: Redis connection string
- arweaveURL: Arweave gateway
- hymxURL: Hymx node RPC (used by the SDK to send messages)
- prvKey: EVM private key (optional; alternative to `keyfilePath`)
- keyfilePath: path to Arweave keyfile
- nodeName/nodeDesc/nodeURL: node metadata
- joinNetwork: whether to join the Hymx network

### Start commands
```bash
# Run in foreground
go build -o hymx ./cmd && ./hymx --config ./cmd/config.yaml

# Daemon mode
./hymx start --config ./cmd/config.yaml

# Stop daemon
./hymx stop
```

On startup the service will:
- Build a Bundler from the provided wallet;
- Mount `aox.token.0.0.1` to `SpawnVmToken`;
- Expose cache endpoint `/cache/<pid>/<key>`.

## Using the SDK

### Spawn a token
```go
res, err := hySdk.SpawnAndWait(
    MODULE,    // module Id; sample Ids available under cmd/mod/*.json
    SCHEDULER, // scheduler AccId (usually this node's address)
    []goarSchema.Tag{
        {Name: "Name", Value: "a Token"},
        {Name: "Ticker", Value: "aToken"},
        {Name: "Decimals", Value: "12"},
        {Name: "Logo", Value: "<arweave-txid>"},
    },
)
tokenPid := res.Id
```

### Query info and balances
```go
// Info
_, _ = hySdk.SendMessageAndWait(tokenPid, "", []goarSchema.Tag{{Name: "Action", Value: "Info"}})

// Total supply
_, _ = hySdk.SendMessageAndWait(tokenPid, "", []goarSchema.Tag{{Name: "Action", Value: "Total-Supply"}})

// All balances
_, _ = hySdk.SendMessageAndWait(tokenPid, "", []goarSchema.Tag{{Name: "Action", Value: "Balances"}})

// Account balance (choose one: Recipient/Target)
_, _ = hySdk.SendMessageAndWait(tokenPid, "", []goarSchema.Tag{
    {Name: "Action", Value: "Balance"},
    {Name: "Recipient", Value: "0x..."},
})
```

### Transfer, Mint, Burn
```go
// Transfer
_, _ = hySdk.SendMessageAndWait(tokenPid, "", []goarSchema.Tag{
    {Name: "Action", Value: "Transfer"},
    {Name: "Recipient", Value: "0x... or ar..."},
    {Name: "Quantity", Value: "100000"},
})

// Mint (owner-only)
_, _ = hySdk.SendMessageAndWait(tokenPid, "", []goarSchema.Tag{
    {Name: "Action", Value: "Mint"},
    {Name: "Recipient", Value: hySdk.GetAddress()},
    {Name: "Quantity", Value: "50000000"},
})

// Burn (from caller; net burn = Quantity - BurnFee)
_, _ = hySdk.SendMessageAndWait(tokenPid, "", []goarSchema.Tag{
    {Name: "Action", Value: "Burn"},
    {Name: "Quantity", Value: "1000"},
    // optionally specify X-Recipient or Recipient for off-chain settlement hints
})
```

### Update parameters (owner-only)
```go
_, _ = hySdk.SendMessageAndWait(tokenPid, "", []goarSchema.Tag{
    {Name: "Action", Value: "Set-Params"},
    {Name: "BurnFee", Value: "10"},        // burn fee
    {Name: "FeeRecipient", Value: "0x..."},
    {Name: "Name", Value: "NewName"},
})
```

## Address and Amount Rules
- Addresses: `Recipient/Target` supports EVM or Arweave addresses; normalized via `IDCheck` internally.
- Amounts: all amounts are big integers represented as decimal strings.
- Burn: requires `Quantity >= BurnFee`, otherwise `err_incorrect_quantity` is returned.

## Examples and Tests
- See `example/vmtoken_test.go` for a full, end-to-end example from spawn to action calls and cache reads.
- Sample module Ids are provided under `cmd/mod/*.json`. Your environment may differ; use your actual deployed Ids.

## Development and Build
```bash
go mod tidy
go build -o hymx ./cmd
```

## License
See `LICENSE` in the repository.