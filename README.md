## hymx-vmtoken

A generic Token VM implementation for Hymatrix Hymx with support for both basic and cross-chain tokens.

This project provides:
- **Basic Token VM**: Standard token functionality without burn support (ModuleFormat: `hymx.basic.token.0.0.1`)
- **Cross-Chain Token VM**: Extended token functionality with burn support for cross-chain operations (ModuleFormat: `hymx.crosschain.token.0.0.1`)
- **Server integration**: Mount both token types as VMs on a Hymx node and expose token parameter cache reading
- **SDK examples**: Use `github.com/hymatrix/hymx/sdk` to deploy and interact with tokens (see examples for end-to-end tests)

Best suited for developers who want to issue and manage fungible tokens on Hymx, with options for basic tokens or cross-chain tokens with burn functionality.

### Token Types and Module Formats

#### Basic Token (Recommended for simple use cases)
- **Module Format**: `schema.VmTokenBasicModuleFormat = "hymx.basic.token.0.0.1"`
- **Features**: Info, Set-Params, Total-Supply, Balance, Transfer, Mint
- **No Burn Support**: Lighter weight, suitable for basic asset issuance
- **Spawn Function**: `vmtoken.SpawnBasicToken`

#### Cross-Chain Token (For cross-chain operations)
- **Module Format**: `schema.VmTokenCrossChainModuleFormat = "hymx.crosschain.token.0.0.1"`
- **Features**: All basic features + Burn functionality
- **Burn Support**: Includes burn fees and fee recipients for cross-chain settlements
- **Burn Processor**: Supports custom burn notification processor, defaults to MintOwner
- **Spawn Function**: `vmtoken.SpawnCrossChainToken`

### Server Mounting

When a Hymx node starts, it can mount both token types:

```go
// Mount basic token
s.Mount("hymx.basic.token.0.0.1", vmtoken.SpawnBasicToken)

// Mount cross-chain token  
s.Mount("hymx.crosschain.token.0.0.1", vmtoken.SpawnCrossChainToken)
```

### Spawn Parameters

Both token types require the following tags when instantiating:
- **Name**: token name (required)
- **Ticker**: token symbol (required)  
- **Decimals**: precision (required, decimal string)
- **Logo**: Arweave resource identifier for the logo (optional)

#### Cross-Chain Token Specific Parameters:
- **BurnFee**: burn fee (optional, defaults to "0")
- **FeeRecipient**: fee recipient (optional, defaults to owner)
- **BurnProcessor**: burn transaction processor (optional, defaults to owner)

After spawning:
- `owner` is the spawner account (`env.AccId`)
- Initial state: `totalSupply = 0`, `balances = {}`
- `mintOwner = owner` (the account allowed to call Mint; can be changed by `owner`)
- **Cross-chain tokens only**: `burnFee = 0`, `feeRecipient = owner`, `burnProcessor = owner`

### Actions and Parameters

#### Common Actions (Both Token Types)

- **Info**
  - Returns basic token info.
  - Params: none.
  - Return tags: `Name`, `Ticker`, `Logo`, `Denomination(=Decimals)`, `Owner`, `MintOwner`.
  - **Cross-chain tokens**: Additional tags include `BurnFee`, `FeeRecipient`, `BurnProcessor`.
  - On first call, initializes and writes cache (see Cache Keys).

- **Set-Params** (owner-only)
  - Updates token and account parameters.
  - **Basic tokens**: `Owner`, `MintOwner`, `Name`, `Ticker`, `Decimals`, `Logo`.
  - **Cross-chain tokens**: All basic params + `FeeRecipient`, `BurnFee` (decimal string), `BurnProcessor`.
  - Return tags: `Set-Params-Notice = success`.
  - Cache: refreshes `TokenInfo`.

- **Total-Supply / TotalSupply**
  - Queries total supply.
  - Params: none.
  - Return: message Data is the total as a decimal string; tags include `Action=Total-Supply`, `Ticker`.

- **Balance**
  - Queries balance for an account.
  - Params: either `Recipient` or `Target`. If neither provided, defaults to the caller account.
  - Return tags: `Balance`, `Ticker`, `Account` (queried account). Data is the balance string.

- **Transfer**
  - Transfers from caller to `Recipient`.
  - Params: `Recipient` (required; supports EVM/Arweave addresses), `Quantity` (required; decimal string).
  - Validation: quantity format; sufficient balance; `Recipient` passes ID check.
  - Return: two messages:
    - Caller's `Debit-Notice` with tags `Ticker`, `Action=Debit-Notice`, `Recipient`, `Quantity`, `TransactionId`.
    - Recipient's `Credit-Notice` with tags `Ticker`, `Action=Credit-Notice`, `Sender`, `Quantity`, `TransactionId`.
  - Cache: updates `Balances:<sender>`, `Balances:<recipient>`, `TotalSupply`.

- **Mint** (mintOwner-only)
  - Mints `Quantity` to `Recipient`.
  - Params: `Recipient`, `Quantity` (decimal string).
  - Permission: caller must equal `mintOwner` (configurable by `owner` via `Set-Params`).
  - Return: two `Mint-Notice` messages (to owner and recipient), tags include `Recipient`, `Quantity`, `Ticker`.
  - Cache: updates `totalSupply` and affected account caches.

#### Cross-Chain Token Only Actions

- **Burn**
  - Burns `Quantity` from caller; `burnFee` is transferred to `feeRecipient`. Net burn = `Quantity - burnFee`.
  - Params: `Quantity` (required, decimal string); `Recipient` or `X-Recipient` (optional; defaults to `from`).
  - Validation: `Quantity >= burnFee`; address passes ID check; sufficient balance.
  - Return: one `Burn-Notice` with tags `Sender`, `X-Recipient`, `Quantity` (=net burn), `Ticker`, `TokenId`, `Fee`, `FeeRecipient`.
  - Cache: updates `totalSupply` (minus net burn) and caches for `from` and `feeRecipient`.
  - **Special Note**: Burn notification messages are sent to the address specified by `BurnProcessor`.

### Cache Keys (via Hymx node HTTP)

#### Basic Token Cache Keys
- `TokenInfo`: stringified JSON with `Name`, `Ticker`, `Denomination`, `Logo`, `Owner`, `MintOwner`.
- `TotalSupply`: total supply as a string.
- `Balances:<Account>`: balance string of an account.

#### Cross-Chain Token Cache Keys
- `TokenInfo`: stringified JSON with `Name`, `Ticker`, `Denomination`, `Logo`, `Owner`, `MintOwner`, `BurnFee`, `FeeRecipient`, `BurnProcessor`.
- `TotalSupply`: total supply as a string.
- `Balances:<Account>`: balance string of an account.

### Usage Examples

#### Spawn a Basic Token
```go
res, err := hySdk.SpawnAndWait(
    BASIC_MODULE,    // Basic token module ID generated from Utils module.go
    SCHEDULER,                    // scheduler AccId (usually this node's address)
    []goarSchema.Tag{
        {Name: "Name", Value: "Basic Token"},
        {Name: "Ticker", Value: "bToken"},
        {Name: "Decimals", Value: "12"},
        {Name: "Logo", Value: "<arweave-txid>"},
    },
)
tokenPid := res.Id
```

#### Spawn a Cross-Chain Token
```go
res, err := hySdk.SpawnAndWait(
    CROSSCHAIN_MODULE, // Cross-chain token module ID generated from Utils module.go
    SCHEDULER,                      // scheduler AccId (usually this node's address)
    []goarSchema.Tag{
        {Name: "Name", Value: "Cross-Chain Token"},
        {Name: "Ticker", Value: "ccToken"},
        {Name: "Decimals", Value: "18"},
        {Name: "Logo", Value: "<arweave-txid>"},
        {Name: "BurnFee", Value: "100"},           // burn fee
        {Name: "FeeRecipient", Value: "0x..."},    // fee recipient
        {Name: "BurnProcessor", Value: "0x..."},   // burn processor
    },
)
tokenPid := res.Id
```

#### Query info and balances
```go
// Info
_, _ = hySdk.SendMessageAndWait(tokenPid, "", []goarSchema.Tag{{Name: "Action", Value: "Info"}})

// Total supply
_, _ = hySdk.SendMessageAndWait(tokenPid, "", []goarSchema.Tag{{Name: "Action", Value: "Total-Supply"}})

// Account balance (choose one: Recipient/Target)
_, _ = hySdk.SendMessageAndWait(tokenPid, "", []goarSchema.Tag{
    {Name: "Action", Value: "Balance"},
    {Name: "Recipient", Value: "0x..."},
})
```

#### Transfer, Mint, Burn
```go
// Transfer
_, _ = hySdk.SendMessageAndWait(tokenPid, "", []goarSchema.Tag{
    {Name: "Action", Value: "Transfer"},
    {Name: "Recipient", Value: "0x... or ar..."},
    {Name: "Quantity", Value: "100000"},
})

// Mint (mintOwner-only)
_, _ = hySdk.SendMessageAndWait(tokenPid, "", []goarSchema.Tag{
    {Name: "Action", Value: "Mint"},
    {Name: "Recipient", Value: "0x... or ar..."},
    {Name: "Quantity", Value: "50000000"},
})

// Burn (Cross-chain tokens only; from caller; net burn = Quantity - BurnFee)
_, _ = hySdk.SendMessageAndWait(tokenPid, "", []goarSchema.Tag{
    {Name: "Action", Value: "Burn"},
    {Name: "Quantity", Value: "1000"},
    // optionally specify X-Recipient or Recipient for off-chain settlement hints
})
```

#### Update parameters (owner-only)
```go
// Basic token parameters
_, _ = hySdk.SendMessageAndWait(tokenPid, "", []goarSchema.Tag{
    {Name: "Action", Value: "Set-Params"},
    {Name: "MintOwner", Value: "0x..."},   // account allowed to call Mint
    {Name: "Name", Value: "NewName"},
})

// Cross-chain token parameters (includes burn-related params)
_, _ = hySdk.SendMessageAndWait(tokenPid, "", []goarSchema.Tag{
    {Name: "Action", Value: "Set-Params"},
    {Name: "MintOwner", Value: "0x..."},   // account allowed to call Mint
    {Name: "BurnFee", Value: "10"},        // burn fee (cross-chain tokens only)
    {Name: "FeeRecipient", Value: "0x..."}, // fee recipient (cross-chain tokens only)
    {Name: "BurnProcessor", Value: "0x..."}, // burn processor (cross-chain tokens only)
    {Name: "Name", Value: "NewName"},
})
```

## Address and Amount Rules
- Addresses: `Recipient/Target` supports EVM or Arweave addresses; normalized via `IDCheck` internally.
- Amounts: all amounts are big integers represented as decimal strings.
- Burn: requires `Quantity >= BurnFee`, otherwise `err_incorrect_quantity` is returned.

## Token Type Selection Guide

### Choose Basic Token when:
- You need simple token functionality (mint, transfer, query)
- You don't require burn functionality
- You want a lighter weight solution
- You're building basic assets on Hymx

### Choose Cross-Chain Token when:
- You need burn functionality for cross-chain operations
- You want to implement token bridges or cross-chain settlements
- You need burn fees and fee recipients
- You need custom burn processors
- You're building cross-chain applications

## Examples and Tests
- See `example/basic_token_test.go` for a full, end-to-end example of basic token functionality
- See `example/crosschain_token_test.go` for a full, end-to-end example of cross-chain token functionality
- Sample module Ids are provided under `cmd/mod/*.json`. Your environment may differ; use your actual deployed Ids.

## Development and Build
```bash
go mod tidy
go build -o hymx ./cmd
```