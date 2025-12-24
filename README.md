# hymx-vmtoken

A generic Token VM implementation for Hymatrix Hymx with support for basic and cross-chain tokens.

## Overview

This project provides two token VM implementations:

- **Basic Token VM**: Standard token functionality with mint, transfer, query, and other basic operations (Module Format: `hymx.basic.token.0.0.1`)
- **Cross-Chain Token VM**: Extended token functionality that adds cross-chain mint and burn operations on top of basic tokens (Module Format: `hymx.crosschain.token.0.0.1`)

All token types can be mounted as VMs on Hymx nodes and provide fast state queries through a caching system.

## Token Types

### Basic Token

Basic tokens provide standard fungible token functionality suitable for most token issuance scenarios.

**Module Format**: `hymx.basic.token.0.0.1`

**Features**:
- ✅ Token info query (Info)
- ✅ Parameter setting (Set-Params)
- ✅ Total supply query (Total-Supply)
- ✅ Balance query (Balance)
- ✅ Transfer
- ✅ Mint
- ❌ No burn support

**Use Cases**:
- Simple token issuance
- Scenarios that don't require burn functionality
- Lightweight token requirements

### Cross-Chain Token

Cross-chain tokens extend basic token functionality with cross-chain mint and burn operations, supporting multi-chain asset bridging.

**Module Format**: `hymx.crosschain.token.0.0.1`

**Features**:
- ✅ All basic token features
- ✅ Cross-chain mint
- ✅ Cross-chain burn
- ✅ Multi-chain support (supports source tokens from different chain types)
- ✅ Locked amount tracking (SourceLockAmounts)
- ✅ Burn fees (BurnFees, configurable per chain type)
- ✅ Fee recipient (FeeRecipient)
- ✅ Burn processor (BurnProcessor)

**Use Cases**:
- Cross-chain asset bridging
- Multi-chain token issuance
- Scenarios requiring burn functionality
- Cross-chain DeFi applications

## Operations

### Basic Token Operations

#### 1. Spawn

When creating a basic token instance, provide the following parameters:

**Required Parameters**:
- `Name`: Token name
- `Ticker`: Token symbol
- `Decimals`: Decimal places (decimal string)

**Optional Parameters**:
- `Logo`: Token logo (Arweave transaction ID)
- `Description`: Token description
- `MintOwner`: Mint permission owner (defaults to creator)
- `MaxSupply`: Maximum supply (decimal string, defaults to "0" meaning unlimited)

**Example**:
```go
res, err := hySdk.SpawnAndWait(
    BASIC_MODULE,    // Basic token module ID
    SCHEDULER,       // Scheduler address
    []goarSchema.Tag{
        {Name: "Name", Value: "My Token"},
        {Name: "Ticker", Value: "MTK"},
        {Name: "Decimals", Value: "18"},
        {Name: "MaxSupply", Value: "1000000000"},
        {Name: "MintOwner", Value: "0x..."},
    },
)
tokenId := res.Id
```

#### 2. Info Operation

Query token information.

**Parameters**: None

**Return Tags**:
- `Name`: Token name
- `Ticker`: Token symbol
- `Decimals`: Decimal places
- `Logo`: Logo
- `Description`: Description
- `Owner`: Owner
- `MintOwner`: Mint permission owner
- `MaxSupply`: Maximum supply

**Example**:
```go
_, _ = hySdk.SendMessageAndWait(tokenId, "", []goarSchema.Tag{
    {Name: "Action", Value: "Info"},
})
```

#### 3. Set-Params Operation

Update token parameters (Owner only).

**Updatable Parameters**:
- `TokenOwner`: New token owner
- `MintOwner`: New mint permission owner
- `Name`: Token name
- `Ticker`: Token symbol
- `Decimals`: Decimal places
- `Logo`: Logo
- `Description`: Description
- `MaxSupply`: Maximum supply (decimal string)

**Example**:
```go
_, _ = hySdk.SendMessageAndWait(tokenId, "", []goarSchema.Tag{
    {Name: "Action", Value: "Set-Params"},
    {Name: "Name", Value: "New Name"},
    {Name: "MintOwner", Value: "0x..."},
    {Name: "MaxSupply", Value: "2000000000"},
})
```

#### 4. Total-Supply Operation

Query token total supply.

**Parameters**: None

**Returns**: Total supply (decimal string)

**Example**:
```go
_, _ = hySdk.SendMessageAndWait(tokenId, "", []goarSchema.Tag{
    {Name: "Action", Value: "Total-Supply"},
})
```

#### 5. Balance Operation

Query account balance.

**Parameters**:
- `Recipient` or `Target`: Account address (optional, defaults to caller)

**Return Tags**:
- `Balance`: Balance
- `Account`: Account address
- `Ticker`: Token symbol

**Example**:
```go
_, _ = hySdk.SendMessageAndWait(tokenId, "", []goarSchema.Tag{
    {Name: "Action", Value: "Balance"},
    {Name: "Recipient", Value: "0x..."},
})
```

#### 6. Transfer Operation

Transfer tokens.

**Parameters**:
- `Recipient`: Recipient address (required)
- `Quantity`: Transfer amount (decimal string, required)
- `X-*`: Any tags starting with `X-` will be forwarded to notification messages

**Notification Messages**:
- Sender receives `Debit-Notice` message
- Recipient receives `Credit-Notice` message

**Example**:
```go
_, _ = hySdk.SendMessageAndWait(tokenId, "", []goarSchema.Tag{
    {Name: "Action", Value: "Transfer"},
    {Name: "Recipient", Value: "0x..."},
    {Name: "Quantity", Value: "1000"},
})
```

#### 7. Mint Operation

Mint tokens (MintOwner only).

**Parameters**:
- `Recipient`: Recipient address (required)
- `Quantity`: Mint amount (decimal string, required)

**Validation**:
- If `MaxSupply` is set, total supply after minting cannot exceed maximum supply

**Notification Messages**:
- Both MintOwner and recipient receive `Mint-Notice` messages

**Example**:
```go
_, _ = hySdk.SendMessageAndWait(tokenId, "", []goarSchema.Tag{
    {Name: "Action", Value: "Mint"},
    {Name: "Recipient", Value: "0x..."},
    {Name: "Quantity", Value: "10000"},
})
```

### Cross-Chain Token Operations

Cross-chain tokens support all basic token operations and additionally provide the following operations:

#### 1. Spawn

When creating a cross-chain token instance, provide the following parameters:

**Required Parameters**:
- `Name`: Token name
- `Ticker`: Token symbol
- `Decimals`: Decimal places (decimal string)

**Optional Parameters**:
- `Logo`: Token logo
- `Description`: Token description
- `MintOwner`: Mint permission owner (defaults to creator)
- `BurnFees`: Burn fees (JSON format, e.g., `{"ethereum":"100","bsc":"50"}`)
- `FeeRecipient`: Fee recipient (defaults to creator)
- `BurnProcessor`: Burn processor (optional, for receiving burn notifications)

**Example**:
```go
res, err := hySdk.SpawnAndWait(
    CROSSCHAIN_MODULE, // Cross-chain token module ID
    SCHEDULER,
    []goarSchema.Tag{
        {Name: "Name", Value: "Cross-Chain Token"},
        {Name: "Ticker", Value: "CCT"},
        {Name: "Decimals", Value: "18"},
        {Name: "BurnFees", Value: `{"ethereum":"100","bsc":"50"}`},
        {Name: "FeeRecipient", Value: "0x..."},
        {Name: "BurnProcessor", Value: "0x..."},
    },
)
tokenId := res.Id
```

#### 2. Info Operation

Query cross-chain token information.

**Return Tags** (includes all basic token tags, plus):
- `BurnFees`: Burn fees (JSON string)
- `FeeRecipient`: Fee recipient
- `BurnProcessor`: Burn processor
- `SourceTokenChains`: Source token chain mapping (JSON string, format: `{"sourceTokenId":"chainType"}`)
- `SourceLockAmounts`: Source chain locked amounts (JSON string, format: `{"chainType:sourceTokenId":"amount"}`)

#### 3. Set-Params Operation

Update cross-chain token parameters (Owner only).

**Updatable Parameters** (includes all basic token parameters, plus):
- `BurnFees`: Burn fees (JSON format, e.g., `{"ethereum":"200","bsc":"100"}`)
- `FeeRecipient`: Fee recipient
- `BurnProcessor`: Burn processor

#### 4. Mint Operation (Cross-Chain Mint)

Cross-chain mint tokens (MintOwner only).

**Parameters**:
- `Recipient`: Recipient address (required)
- `Quantity`: Mint amount (decimal string, required)
- `SourceChainType`: Source chain type (required, e.g., "ethereum", "bsc")
- `SourceTokenId`: Source token ID (required)
- `X-MintTxHash`: Mint transaction hash (optional, for preventing duplicate mints)

**Functionality**:
1. Verify if `X-MintTxHash` has been used (if provided)
2. Establish or verify source token chain mapping (`SourceTokenChain`)
3. Increase recipient balance
4. Increase total supply
5. Increase locked amount for corresponding source chain (`SourceLockAmount`)
6. Record mint transaction hash (if provided)

**Example**:
```go
_, _ = hySdk.SendMessageAndWait(tokenId, "", []goarSchema.Tag{
    {Name: "Action", Value: "Mint"},
    {Name: "Recipient", Value: "0x..."},
    {Name: "Quantity", Value: "1000"},
    {Name: "SourceChainType", Value: "ethereum"},
    {Name: "SourceTokenId", Value: "0xa0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48"},
    {Name: "X-MintTxHash", Value: "0x1234..."},
})
```

#### 5. Burn Operation (Cross-Chain Burn)

Cross-chain burn tokens.

**Parameters**:
- `Quantity`: Burn amount (decimal string, required)
- `TargetTokenId`: Target token ID (required, to determine target chain)
- `Recipient` or `X-Recipient`: Recipient hint (optional, defaults to caller)

**Functionality**:
1. Find corresponding chain type based on `TargetTokenId`
2. Verify that burn fee for that chain type is set
3. Verify that locked amount is sufficient (requires `Quantity - BurnFee`)
4. Deduct full amount from caller (`Quantity`)
5. Transfer burn fee to `FeeRecipient`
6. Decrease total supply (decrease by `Quantity - BurnFee`)
7. Decrease locked amount for corresponding source chain (decrease by `Quantity - BurnFee`)
8. Send burn notification to `BurnProcessor`

**Validation Rules**:
- `Quantity >= BurnFee` (otherwise returns `err_incorrect_quantity`)
- Locked amount >= `Quantity - BurnFee` (otherwise returns `err_insufficient_lock_amount`)

**Notification Messages**:
- `BurnProcessor` receives `Burn-Notice` message containing:
  - `Sender`: Burner address
  - `X-Recipient`: Recipient hint
  - `Quantity`: Net burn amount (`Quantity - BurnFee`)
  - `Fee`: Fee
  - `FeeRecipient`: Fee recipient
  - `TargetChainType`: Target chain type
  - `TargetTokenId`: Target token ID

**Example**:
```go
_, _ = hySdk.SendMessageAndWait(tokenId, "", []goarSchema.Tag{
    {Name: "Action", Value: "Burn"},
    {Name: "Quantity", Value: "1000"},
    {Name: "TargetTokenId", Value: "0xa0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48"},
    {Name: "X-Recipient", Value: "0x..."},
})
```

## Cache System

All token states are provided through a caching system for fast queries.

### Basic Token Cache Keys

- `token-info`: JSON string of token information, containing:
  - `Name`: Token name
  - `Ticker`: Token symbol
  - `Decimals`: Decimal places
  - `Logo`: Logo
  - `Description`: Description
  - `Owner`: Owner
  - `MintOwner`: Mint permission owner
  - `MaxSupply`: Maximum supply
- `total-supply`: Total supply (decimal string)
- `balances:<Account>`: Account balance (decimal string)
- `Balances`: Complete balance mapping JSON string

### Cross-Chain Token Cache Keys

- `token-info`: JSON string of token information, containing:
  - All basic token fields
  - `BurnFees`: Burn fees (JSON string)
  - `FeeRecipient`: Fee recipient
  - `BurnProcessor`: Burn processor
  - `SourceTokenChains`: Source token chain mapping (JSON string)
  - `SourceLockAmounts`: Source chain locked amounts (JSON string)
- `total-supply`: Total supply (decimal string)
- `balances:<Account>`: Account balance (decimal string)
- `Balances`: Complete balance mapping JSON string

### Cache Query Examples

```go
// Query token info
info, err := hySdk.Client.GetCache(tokenId, "token-info")

// Query total supply
totalSupply, err := hySdk.Client.GetCache(tokenId, "total-supply")

// Query account balance
balance, err := hySdk.Client.GetCache(tokenId, "balances:"+accountId)
```

## Address and Amount Rules

### Address Format

- Supports EVM addresses (starting with 0x) and Arweave addresses
- All addresses are normalized internally through `IDCheck`
- `Recipient`/`Target` parameters support both address formats

### Amount Format

- All amounts are big integers represented as decimal strings
- Examples: `"1000000"`, `"1000000000000000000"` (18 decimals)

### Burn Rules

- Burn amount must be >= burn fee, otherwise returns `err_incorrect_quantity`
- For cross-chain burns, locked amount must be >= net burn amount (`Quantity - BurnFee`)

## Error Codes

| Error Code | Description |
|------------|-------------|
| `err_insufficient_balance` | Insufficient balance |
| `err_insufficient_max_supply` | Exceeds maximum supply |
| `err_invalid_from` | Invalid sender address |
| `err_missing_recipient` | Missing recipient parameter |
| `err_missing_quantity` | Missing quantity parameter |
| `err_invalid_quantity_format` | Invalid quantity format |
| `err_incorrect_owner` | Insufficient permissions (requires Owner or MintOwner) |
| `err_repeat_mint` | Duplicate mint (same X-MintTxHash) |
| `err_incorrect_quantity` | Incorrect quantity (burn amount < fee) |
| `err_missing_source_chain` | Missing source chain type |
| `err_missing_source_token_id` | Missing source token ID |
| `err_missing_target_token_id` | Missing target token ID |
| `err_incorrect_target_token_id` | Target token ID does not exist |
| `err_lock_amount_empty` | Locked amount is empty |
| `err_insufficient_lock_amount` | Insufficient locked amount |
| `err_missing_burn_fee` | Missing burn fee configuration |

## Token Type Selection Guide

### Choose Basic Token when:

- ✅ You only need simple token functionality (mint, transfer, query)
- ✅ You don't need burn functionality
- ✅ You want a lighter-weight solution
- ✅ You're issuing basic assets on Hymx

### Choose Cross-Chain Token when:

- ✅ You need burn functionality for cross-chain operations
- ✅ You need to implement token bridging or cross-chain settlements
- ✅ You need burn fees and fee recipients
- ✅ You need custom burn processors
- ✅ You need to track multi-chain locked amounts
- ✅ You need to mint tokens based on assets from different chains
- ✅ You're building cross-chain applications

## Server Integration

The project provides server integration to mount token VMs on Hymx nodes:

```go
s := server.New(node, nil)
s.Mount(schema.VmTokenBasicModuleFormat, basic.Spawn)
s.Mount(schema.VmTokenCrossChainModuleFormat, crosschain.Spawn)
s.Run(port)
```

## Testing and Examples

The project provides complete test cases:

- `test/basic_token_test.go`: Complete test cases for basic tokens
- `test/crosschain_token_test.go`: Complete test cases for cross-chain tokens

Test cases demonstrate:
- Token creation and initialization
- How to call various operations
- Cache queries and validation
- Complete cross-chain mint and burn workflows

## License

See LICENSE file for details.
