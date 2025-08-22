## hymx-vmtoken

基于 Hymatrix Hymx 的通用 Token VM 实现，支持基础代币和跨链代币两种类型。

此项目提供：
- **基础代币 VM**：标准代币功能，无销毁支持（模块格式：`hymx.basic.token.0.0.1`）
- **跨链代币 VM**：扩展代币功能，包含销毁功能，适用于跨链操作（模块格式：`hymx.crosschain.token.0.0.1`）
- **服务端集成**：将两种代币类型作为 VM 挂载到 Hymx 节点，并实现代币参数 cache 读取
- **SDK 使用示例**：使用 `github.com/hymatrix/hymx/sdk` 进行部署与调用代币 (详细测试请参考 example)

适合需要在 Hymx 上发行和管理同质化代币的开发者，可选择基础代币或具有销毁功能的跨链代币。

### 代币类型与模块格式

#### 基础代币（推荐用于简单场景）
- **模块格式**：`schema.VmTokenBasicModuleFormat = "hymx.basic.token.0.0.1"`
- **功能特性**：Info, Set-Params, Total-Supply, Balance, Transfer, Mint
- **无销毁支持**：更轻量，适合基础资产发行
- **生成函数**：`vmtoken.SpawnBasicToken`

#### 跨链代币（适用于跨链操作）
- **模块格式**：`schema.VmTokenCrossChainModuleFormat = "hymx.crosschain.token.0.0.1"`
- **功能特性**：包含所有基础功能 + 销毁功能
- **销毁支持**：包含销毁手续费和手续费接收者，适用于跨链结算
- **销毁处理器**：支持自定义销毁通知处理器，默认与 MintOwner 一致
- **生成函数**：`vmtoken.SpawnCrossChainToken`

### 服务端挂载

当 Hymx 节点启动时，可以挂载两种代币类型：

```go
// 挂载基础代币
s.Mount("hymx.basic.token.0.0.1", vmtoken.SpawnBasicToken)

// 挂载跨链代币
s.Mount("hymx.crosschain.token.0.0.1", vmtoken.SpawnCrossChainToken)
```

### 实例化（Spawn）参数

两种代币类型都需要在实例化时传入以下标签（Tags）：
- **Name**: 代币名称（必填）
- **Ticker**: 代币符号（必填）
- **Decimals**: 精度（必填，字符串数值）
- **Logo**: Logo 的 Arweave 资源标识（可选）

#### 跨链代币特有参数：
- **BurnFee**: 销毁手续费（可选，默认为 "0"）
- **FeeRecipient**: 手续费接收者（可选，默认为 owner）
- **BurnProcessor**: 销毁交易后处理器（可选，默认为 owner）

实例化后：
- `owner` 是实例化账户（`env.AccId`）
- 初始状态：`totalSupply = 0`，`balances = {}`
- `mintOwner = owner`（允许调用 Mint 的账户；可由 `owner` 更改）
- **跨链代币**：`burnFee = 0`，`feeRecipient = owner`，`burnProcessor = owner`

### 操作和参数

#### 通用操作（两种代币类型）

- **Info**
  - 返回基本代币信息
  - 参数：无
  - 返回标签：`Name`、`Ticker`、`Logo`、`Denomination(=Decimals)`、`Owner`、`MintOwner`
  - **跨链代币**：额外标签包括 `BurnFee`、`FeeRecipient`、`BurnProcessor`
  - 首次调用时，初始化并写入缓存（见缓存键）

- **Set-Params**（仅限 owner）
  - 更新代币和账户参数
  - **基础代币**：`Owner`、`MintOwner`、`Name`、`Ticker`、`Decimals`、`Logo`
  - **跨链代币**：所有基础参数 + `FeeRecipient`、`BurnFee`（十进制字符串）、`BurnProcessor`
  - 返回标签：`Set-Params-Notice = success`
  - 缓存：刷新 `TokenInfo`

- **Total-Supply / TotalSupply**
  - 查询总供应量
  - 参数：无
  - 返回：消息 Data 是总供应量的十进制字符串；标签包括 `Action=Total-Supply`、`Ticker`

- **Balance**
  - 查询账户余额
  - 参数：`Recipient` 或 `Target`。如果都未提供，默认为调用者账户
  - 返回标签：`Balance`、`Ticker`、`Account`（查询的账户）。Data 是余额字符串

- **Transfer**
  - 从调用者转账到 `Recipient`
  - 参数：`Recipient`（必需；支持 EVM/Arweave 地址）、`Quantity`（必需；十进制字符串）
  - 验证：数量格式；足够余额；`Recipient` 通过 ID 检查
  - 返回：两个消息：
    - 调用者的 `Debit-Notice`，标签包括 `Ticker`、`Action=Debit-Notice`、`Recipient`、`Quantity`、`TransactionId`
    - 接收者的 `Credit-Notice`，标签包括 `Ticker`、`Action=Credit-Notice`、`Sender`、`Quantity`、`TransactionId`
  - 缓存：更新 `Balances:<sender>`、`Balances:<recipient>`、`TotalSupply`

- **Mint**（仅限 mintOwner）
  - 向 `Recipient` 增发 `Quantity`
  - 参数：`Recipient`、`Quantity`（十进制字符串）
  - 权限：调用者必须等于 `mintOwner`（可通过 `Set-Params` 由 `owner` 配置）
  - 返回：两个 `Mint-Notice` 消息（给 owner 和 recipient），标签包括 `Recipient`、`Quantity`、`Ticker`
  - 缓存：更新 `totalSupply` 和受影响的账户缓存

#### 跨链代币特有操作

- **Burn**
  - 从调用者销毁 `Quantity`；`burnFee` 转移给 `feeRecipient`。净销毁量 = `Quantity - burnFee`
  - 参数：`Quantity`（必需，十进制字符串）；`Recipient` 或 `X-Recipient`（可选；默认为 `from`）
  - 验证：`Quantity >= burnFee`；地址通过 ID 检查；足够余额
  - 返回：一个 `Burn-Notice`，标签包括 `Sender`、`X-Recipient`、`Quantity`（=净销毁量）、`Ticker`、`TokenId`、`Fee`、`FeeRecipient`
  - 缓存：更新 `totalSupply`（减去净销毁量）和 `from`、`feeRecipient` 的缓存
  - **特殊说明**：销毁通知消息会发送到 `BurnProcessor` 指定的地址

### 缓存键（通过 Hymx 节点 HTTP）

#### 基础代币缓存键
- `TokenInfo`：包含 `Name`、`Ticker`、`Denomination`、`Logo`、`Owner`、`MintOwner` 的 JSON 字符串
- `TotalSupply`：总供应量字符串
- `Balances:<Account>`：账户余额字符串

#### 跨链代币缓存键
- `TokenInfo`：包含 `Name`、`Ticker`、`Denomination`、`Logo`、`Owner`、`MintOwner`、`BurnFee`、`FeeRecipient`、`BurnProcessor` 的 JSON 字符串
- `TotalSupply`：总供应量字符串
- `Balances:<Account>`：账户余额字符串

### 使用示例

#### 生成基础代币
```go
res, err := hySdk.SpawnAndWait(
    BASIC_MODULE,    // 通过 Utils 中 module.go 生成的 基础代币的模块 Id
    SCHEDULER,                    // 调度器 AccId（通常是本节点地址）
    []goarSchema.Tag{
        {Name: "Name", Value: "基础代币"},
        {Name: "Ticker", Value: "bToken"},
        {Name: "Decimals", Value: "12"},
        {Name: "Logo", Value: "<arweave-txid>"},
    },
)
tokenPid := res.Id
```

#### 生成跨链代币
```go
res, err := hySdk.SpawnAndWait(
    CROSSCHAIN_MODULE, // 通过 Utils 中 module.go 生成的跨链代币的模块 Id
    SCHEDULER,                      // 调度器 AccId（通常是本节点地址）
    []goarSchema.Tag{
        {Name: "Name", Value: "跨链代币"},
        {Name: "Ticker", Value: "ccToken"},
        {Name: "Decimals", Value: "18"},
        {Name: "Logo", Value: "<arweave-txid>"},
        {Name: "BurnFee", Value: "100"},           // 销毁手续费
        {Name: "FeeRecipient", Value: "0x..."},    // 手续费接收者
        {Name: "BurnProcessor", Value: "0x..."},   // 销毁处理器
    },
)
tokenPid := res.Id
```

#### 查询信息和余额
```go
// 信息查询
_, _ = hySdk.SendMessageAndWait(tokenPid, "", []goarSchema.Tag{{Name: "Action", Value: "Info"}})

// 总供应量
_, _ = hySdk.SendMessageAndWait(tokenPid, "", []goarSchema.Tag{{Name: "Action", Value: "Total-Supply"}})

// 账户余额（选择其一：Recipient/Target）
_, _ = hySdk.SendMessageAndWait(tokenPid, "", []goarSchema.Tag{
    {Name: "Action", Value: "Balance"},
    {Name: "Recipient", Value: "0x..."},
})
```

#### 转账、增发、销毁
```go
// 转账
_, _ = hySdk.SendMessageAndWait(tokenPid, "", []goarSchema.Tag{
    {Name: "Action", Value: "Transfer"},
    {Name: "Recipient", Value: "0x... 或 ar..."},
    {Name: "Quantity", Value: "100000"},
})

// 增发（仅限 mintOwner）
_, _ = hySdk.SendMessageAndWait(tokenPid, "", []goarSchema.Tag{
    {Name: "Action", Value: "Mint"},
    {Name: "Recipient", Value: "0x... 或 ar..."},
    {Name: "Quantity", Value: "50000000"},
})

// 销毁（仅跨链代币；从调用者；净销毁量 = Quantity - BurnFee）
_, _ = hySdk.SendMessageAndWait(tokenPid, "", []goarSchema.Tag{
    {Name: "Action", Value: "Burn"},
    {Name: "Quantity", Value: "1000"},
    // 可选指定 X-Recipient 或 Recipient 用于链下结算提示
})
```

#### 更新参数（仅限 owner）
```go
// 基础代币参数
_, _ = hySdk.SendMessageAndWait(tokenPid, "", []goarSchema.Tag{
    {Name: "Action", Value: "Set-Params"},
    {Name: "MintOwner", Value: "0x..."},   // 允许调用 Mint 的账户
    {Name: "Name", Value: "NewName"},
})

// 跨链代币参数（包含销毁相关参数）
_, _ = hySdk.SendMessageAndWait(tokenPid, "", []goarSchema.Tag{
    {Name: "Action", Value: "Set-Params"},
    {Name: "MintOwner", Value: "0x..."},   // 允许调用 Mint 的账户
    {Name: "BurnFee", Value: "10"},        // 销毁手续费（仅跨链代币）
    {Name: "FeeRecipient", Value: "0x..."}, // 手续费接收者（仅跨链代币）
    {Name: "BurnProcessor", Value: "0x..."}, // 销毁处理器（仅跨链代币）
    {Name: "Name", Value: "NewName"},
})
```

## 地址和数量规则
- 地址：`Recipient/Target` 支持 EVM 或 Arweave 地址；内部通过 `IDCheck` 标准化
- 数量：所有数量都是表示为十进制字符串的大整数
- 销毁：要求 `Quantity >= BurnFee`，否则返回 `err_incorrect_quantity`

## 代币类型选择指南

### 选择基础代币当：
- 您需要简单的代币功能（增发、转账、查询）
- 您不需要销毁功能
- 您想要更轻量的解决方案
- 您在 Hymx 上构建基础资产

### 选择跨链代币当：
- 您需要销毁功能用于跨链操作
- 您想要实现代币桥接或跨链结算
- 您需要销毁手续费和手续费接收者
- 您需要自定义销毁处理器
- 您在构建跨链应用

## 测试示例
- 参见 `example/basic_token_test.go` 了解基础代币的完整端到端测试
- 参见 `example/crosschain_token_test.go` 了解跨链代币的完整端到端测试
- 示例模块 ID 在 `cmd/mod/*.json` 下提供。您的环境可能不同；使用您实际部署的 ID

## 开发和构建
```bash
go mod tidy
go build -o hymx ./cmd
```


