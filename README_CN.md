# hymx-vmtoken

基于 Hymatrix Hymx 的通用 Token VM 实现，支持基础代币和跨链代币两种类型。

## 项目概述

此项目提供了两种代币 VM 实现：

- **基础代币 VM**：标准代币功能，支持铸造、转账、查询等基础操作（模块格式：`hymx.basic.token.0.0.1`）
- **跨链代币 VM**：扩展代币功能，在基础代币基础上增加了跨链铸造和销毁功能（模块格式：`hymx.crosschain.token.0.0.1`）

所有代币类型都可以作为 VM 挂载到 Hymx 节点，并通过缓存系统提供快速的状态查询。

## 代币类型

### 基础代币（Basic Token）

基础代币提供标准的同质化代币功能，适用于大多数代币发行场景。

**模块格式**：`hymx.basic.token.0.0.1`

**功能特性**：
- ✅ 代币信息查询（Info）
- ✅ 参数设置（Set-Params）
- ✅ 总供应量查询（Total-Supply）
- ✅ 余额查询（Balance）
- ✅ 转账（Transfer）
- ✅ 铸造（Mint）
- ❌ 不支持销毁（Burn）

**适用场景**：
- 简单的代币发行
- 不需要销毁功能的场景
- 轻量级代币需求

### 跨链代币（Cross-Chain Token）

跨链代币在基础代币功能基础上，增加了跨链铸造和销毁功能，支持多链资产桥接。

**模块格式**：`hymx.crosschain.token.0.0.1`

**功能特性**：
- ✅ 所有基础代币功能
- ✅ 跨链铸造（Cross-Chain Mint）
- ✅ 跨链销毁（Cross-Chain Burn）
- ✅ 多链支持（支持不同链类型的源代币）
- ✅ 锁定数量跟踪（SourceLockAmounts）
- ✅ 销毁手续费（BurnFees，按链类型配置）
- ✅ 手续费接收者（FeeRecipient）
- ✅ 销毁处理器（BurnProcessor）

**适用场景**：
- 跨链资产桥接
- 多链代币发行
- 需要销毁功能的场景
- 跨链 DeFi 应用

## 操作说明

### 基础代币操作

#### 1. 实例化（Spawn）

创建基础代币实例时，需要提供以下参数：

**必需参数**：
- `Name`：代币名称
- `Ticker`：代币符号
- `Decimals`：小数位数（十进制字符串）

**可选参数**：
- `Logo`：代币 Logo（Arweave 交易 ID）
- `Description`：代币描述
- `MintOwner`：铸造权限所有者（默认为创建者）
- `MaxSupply`：最大供应量（十进制字符串，默认为 "0" 表示无限制）

**示例**：
```go
res, err := hySdk.SpawnAndWait(
    BASIC_MODULE,    // 基础代币模块 ID
    SCHEDULER,       // 调度器地址
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

#### 2. Info 操作

查询代币信息。

**参数**：无

**返回标签**：
- `Name`：代币名称
- `Ticker`：代币符号
- `Decimals`：小数位数
- `Logo`：Logo
- `Description`：描述
- `Owner`：所有者
- `MintOwner`：铸造权限所有者
- `MaxSupply`：最大供应量

**示例**：
```go
_, _ = hySdk.SendMessageAndWait(tokenId, "", []goarSchema.Tag{
    {Name: "Action", Value: "Info"},
})
```

#### 3. Set-Params 操作

更新代币参数（仅限 Owner）。

**可更新参数**：
- `TokenOwner`：新的代币所有者
- `MintOwner`：新的铸造权限所有者
- `Name`：代币名称
- `Ticker`：代币符号
- `Decimals`：小数位数
- `Logo`：Logo
- `Description`：描述
- `MaxSupply`：最大供应量（十进制字符串）

**示例**：
```go
_, _ = hySdk.SendMessageAndWait(tokenId, "", []goarSchema.Tag{
    {Name: "Action", Value: "Set-Params"},
    {Name: "Name", Value: "New Name"},
    {Name: "MintOwner", Value: "0x..."},
    {Name: "MaxSupply", Value: "2000000000"},
})
```

#### 4. Total-Supply 操作

查询代币总供应量。

**参数**：无

**返回**：总供应量（十进制字符串）

**示例**：
```go
_, _ = hySdk.SendMessageAndWait(tokenId, "", []goarSchema.Tag{
    {Name: "Action", Value: "Total-Supply"},
})
```

#### 5. Balance 操作

查询账户余额。

**参数**：
- `Recipient` 或 `Target`：账户地址（可选，默认为调用者）

**返回标签**：
- `Balance`：余额
- `Account`：账户地址
- `Ticker`：代币符号

**示例**：
```go
_, _ = hySdk.SendMessageAndWait(tokenId, "", []goarSchema.Tag{
    {Name: "Action", Value: "Balance"},
    {Name: "Recipient", Value: "0x..."},
})
```

#### 6. Transfer 操作

转账代币。

**参数**：
- `Recipient`：接收者地址（必需）
- `Quantity`：转账数量（十进制字符串，必需）
- `X-*`：任意以 `X-` 开头的标签会被转发到通知消息中

**通知消息**：
- 发送者收到 `Debit-Notice` 消息
- 接收者收到 `Credit-Notice` 消息

**示例**：
```go
_, _ = hySdk.SendMessageAndWait(tokenId, "", []goarSchema.Tag{
    {Name: "Action", Value: "Transfer"},
    {Name: "Recipient", Value: "0x..."},
    {Name: "Quantity", Value: "1000"},
})
```

#### 7. Mint 操作

铸造代币（仅限 MintOwner）。

**参数**：
- `Recipient`：接收者地址（必需）
- `Quantity`：铸造数量（十进制字符串，必需）

**验证**：
- 如果设置了 `MaxSupply`，铸造后总供应量不能超过最大供应量

**通知消息**：
- MintOwner 和接收者都会收到 `Mint-Notice` 消息

**示例**：
```go
_, _ = hySdk.SendMessageAndWait(tokenId, "", []goarSchema.Tag{
    {Name: "Action", Value: "Mint"},
    {Name: "Recipient", Value: "0x..."},
    {Name: "Quantity", Value: "10000"},
})
```

### 跨链代币操作

跨链代币支持所有基础代币操作，并额外提供以下操作：

#### 1. 实例化（Spawn）

创建跨链代币实例时，需要提供以下参数：

**必需参数**：
- `Name`：代币名称
- `Ticker`：代币符号
- `Decimals`：小数位数（十进制字符串）

**可选参数**：
- `Logo`：代币 Logo
- `Description`：代币描述
- `MintOwner`：铸造权限所有者（默认为创建者）
- `BurnFees`：销毁手续费（JSON 格式，例如：`{"ethereum":"100","bsc":"50"}`）
- `FeeRecipient`：手续费接收者（默认为创建者）
- `BurnProcessor`：销毁处理器（可选，用于接收销毁通知）

**示例**：
```go
res, err := hySdk.SpawnAndWait(
    CROSSCHAIN_MODULE, // 跨链代币模块 ID
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

#### 2. Info 操作

查询跨链代币信息。

**返回标签**（包含基础代币的所有标签，以及）：
- `BurnFees`：销毁手续费（JSON 字符串）
- `FeeRecipient`：手续费接收者
- `BurnProcessor`：销毁处理器
- `SourceTokenChains`：源代币链映射（JSON 字符串，格式：`{"sourceTokenId":"chainType"}`）
- `SourceLockAmounts`：源链锁定数量（JSON 字符串，格式：`{"chainType:sourceTokenId":"amount"}`）

#### 3. Set-Params 操作

更新跨链代币参数（仅限 Owner）。

**可更新参数**（包含基础代币的所有参数，以及）：
- `BurnFees`：销毁手续费（JSON 格式，例如：`{"ethereum":"200","bsc":"100"}`）
- `FeeRecipient`：手续费接收者
- `BurnProcessor`：销毁处理器

#### 4. Mint 操作（跨链铸造）

跨链铸造代币（仅限 MintOwner）。

**参数**：
- `Recipient`：接收者地址（必需）
- `Quantity`：铸造数量（十进制字符串，必需）
- `SourceChainType`：源链类型（必需，例如："ethereum", "bsc"）
- `SourceTokenId`：源代币 ID（必需）
- `X-MintTxHash`：铸造交易哈希（可选，用于防止重复铸造）

**功能说明**：
1. 验证 `X-MintTxHash` 是否已使用（如果提供）
2. 建立或验证源代币链映射（`SourceTokenChain`）
3. 增加接收者余额
4. 增加总供应量
5. 增加对应源链的锁定数量（`SourceLockAmount`）
6. 记录铸造交易哈希（如果提供）

**示例**：
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

#### 5. Burn 操作（跨链销毁）

跨链销毁代币。

**参数**：
- `Quantity`：销毁数量（十进制字符串，必需）
- `TargetTokenId`：目标代币 ID（必需，用于确定目标链）
- `Recipient` 或 `X-Recipient`：接收者提示（可选，默认为调用者）

**功能说明**：
1. 根据 `TargetTokenId` 查找对应的链类型
2. 验证该链类型的销毁手续费已设置
3. 验证锁定数量是否足够（需要 `Quantity - BurnFee`）
4. 从调用者扣除全部数量（`Quantity`）
5. 将销毁手续费转给 `FeeRecipient`
6. 减少总供应量（减少 `Quantity - BurnFee`）
7. 减少对应源链的锁定数量（减少 `Quantity - BurnFee`）
8. 向 `BurnProcessor` 发送销毁通知

**验证规则**：
- `Quantity >= BurnFee`（否则返回 `err_incorrect_quantity`）
- 锁定数量 >= `Quantity - BurnFee`（否则返回 `err_insufficient_lock_amount`）

**通知消息**：
- `BurnProcessor` 收到 `Burn-Notice` 消息，包含：
  - `Sender`：销毁者地址
  - `X-Recipient`：接收者提示
  - `Quantity`：净销毁数量（`Quantity - BurnFee`）
  - `Fee`：手续费
  - `FeeRecipient`：手续费接收者
  - `TargetChainType`：目标链类型
  - `TargetTokenId`：目标代币 ID

**示例**：
```go
_, _ = hySdk.SendMessageAndWait(tokenId, "", []goarSchema.Tag{
    {Name: "Action", Value: "Burn"},
    {Name: "Quantity", Value: "1000"},
    {Name: "TargetTokenId", Value: "0xa0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48"},
    {Name: "X-Recipient", Value: "0x..."},
})
```

## 缓存系统

所有代币状态都通过缓存系统提供快速查询。

### 基础代币缓存键

- `token-info`：代币信息的 JSON 字符串，包含：
  - `Name`：代币名称
  - `Ticker`：代币符号
  - `Decimals`：小数位数
  - `Logo`：Logo
  - `Description`：描述
  - `Owner`：所有者
  - `MintOwner`：铸造权限所有者
  - `MaxSupply`：最大供应量
- `total-supply`：总供应量（十进制字符串）
- `balances:<Account>`：账户余额（十进制字符串）
- `Balances`：完整余额映射的 JSON 字符串

### 跨链代币缓存键

- `token-info`：代币信息的 JSON 字符串，包含：
  - 基础代币的所有字段
  - `BurnFees`：销毁手续费（JSON 字符串）
  - `FeeRecipient`：手续费接收者
  - `BurnProcessor`：销毁处理器
  - `SourceTokenChains`：源代币链映射（JSON 字符串）
  - `SourceLockAmounts`：源链锁定数量（JSON 字符串）
- `total-supply`：总供应量（十进制字符串）
- `balances:<Account>`：账户余额（十进制字符串）
- `Balances`：完整余额映射的 JSON 字符串

### 缓存查询示例

```go
// 查询代币信息
info, err := hySdk.Client.GetCache(tokenId, "token-info")

// 查询总供应量
totalSupply, err := hySdk.Client.GetCache(tokenId, "total-supply")

// 查询账户余额
balance, err := hySdk.Client.GetCache(tokenId, "balances:"+accountId)
```

## 地址和数量规则

### 地址格式

- 支持 EVM 地址（0x 开头）和 Arweave 地址
- 所有地址在内部通过 `IDCheck` 进行标准化处理
- `Recipient`/`Target` 参数支持两种地址格式

### 数量格式

- 所有数量都是大整数，以十进制字符串表示
- 例如：`"1000000"`、`"1000000000000000000"`（18 位小数）

### 销毁规则

- 销毁数量必须 >= 销毁手续费，否则返回 `err_incorrect_quantity`
- 跨链销毁时，锁定数量必须 >= 净销毁数量（`Quantity - BurnFee`）

## 错误码

| 错误码 | 说明 |
|--------|------|
| `err_insufficient_balance` | 余额不足 |
| `err_insufficient_max_supply` | 超过最大供应量 |
| `err_invalid_from` | 无效的发送者地址 |
| `err_missing_recipient` | 缺少接收者参数 |
| `err_missing_quantity` | 缺少数量参数 |
| `err_invalid_quantity_format` | 无效的数量格式 |
| `err_incorrect_owner` | 权限不足（需要 Owner 或 MintOwner） |
| `err_repeat_mint` | 重复铸造（相同的 X-MintTxHash） |
| `err_incorrect_quantity` | 数量不正确（销毁数量 < 手续费） |
| `err_missing_source_chain` | 缺少源链类型 |
| `err_missing_source_token_id` | 缺少源代币 ID |
| `err_missing_target_token_id` | 缺少目标代币 ID |
| `err_incorrect_target_token_id` | 目标代币 ID 不存在 |
| `err_lock_amount_empty` | 锁定数量为空 |
| `err_insufficient_lock_amount` | 锁定数量不足 |
| `err_missing_burn_fee` | 缺少销毁手续费配置 |

## 代币类型选择指南

### 选择基础代币当：

- ✅ 只需要简单的代币功能（铸造、转账、查询）
- ✅ 不需要销毁功能
- ✅ 想要更轻量级的解决方案
- ✅ 在 Hymx 上发行基础资产

### 选择跨链代币当：

- ✅ 需要销毁功能用于跨链操作
- ✅ 需要实现代币桥接或跨链结算
- ✅ 需要销毁手续费和手续费接收者
- ✅ 需要自定义销毁处理器
- ✅ 需要跟踪多链锁定数量
- ✅ 需要基于不同链的资产进行铸造
- ✅ 构建跨链应用

## 服务器集成

项目提供了服务器集成，可以将代币 VM 挂载到 Hymx 节点：

```go
s := server.New(node, nil)
s.Mount(schema.VmTokenBasicModuleFormat, basic.Spawn)
s.Mount(schema.VmTokenCrossChainModuleFormat, crosschain.Spawn)
s.Run(port)
```

## 测试和示例

项目提供了完整的测试用例：

- `test/basic_token_test.go`：基础代币的完整测试用例
- `test/crosschain_token_test.go`：跨链代币的完整测试用例

测试用例展示了：
- 代币创建和初始化
- 各种操作的调用方式
- 缓存查询和验证
- 跨链铸造和销毁的完整流程

## 许可证

详见 LICENSE 文件。
