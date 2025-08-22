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

实例化后：
- `owner` 为发起者账户（`env.AccId`）
- 初始 `totalSupply = 0`，`balances = {}`
- `mintOwner = owner`（允许执行增发 Mint 的账户；可由 `owner` 修改）
- **仅跨链代币**：`burnFee = 0`，`feeRecipient = owner`

### Action 与参数规范

#### 通用 Actions（两种代币类型都支持）

- **Info**
  - 描述：返回代币基础信息。
  - 参数：无。
  - 返回 Tags：`Name`, `Ticker`, `Logo`, `Denomination(=Decimals)`, `Owner`, `MintOwner`。
  - **跨链代币**：额外包含 `BurnFee`, `FeeRecipient` 标签。
  - 首次调用会初始化并写入缓存（见「缓存键」）。

- **Set-Params**（仅限 owner）
  - 描述：更新代币与账户参数。
  - **基础代币参数**：`Owner`, `MintOwner`, `Name`, `Ticker`, `Decimals`, `Logo`。
  - **跨链代币参数**：所有基础参数 + `FeeRecipient`, `BurnFee`（十进制字符串）。
  - 返回 Tags：`Set-Params-Notice = success`。
  - 缓存：更新后会刷新 `TokenInfo` 缓存。

- **Total-Supply / TotalSupply**
  - 描述：查询总供应量。
  - 参数：无。
  - 返回：消息 Data 为十进制字符串的总量；Tags 含 `Action=Total-Supply`, `Ticker`。

- **Balance**
  - 描述：查询某账户余额。
  - 参数：可选其一 `Recipient` 或 `Target`，均不传则默认查询调用者账户。
  - 返回 Tags：`Balance`, `Ticker`, `Account`（被查询账户），Data 为余额字符串。

- **Transfer**
  - 描述：从调用者向 `Recipient` 转账。
  - 参数：`Recipient`（必须，支持 EVM/Arweave 地址），`Quantity`（必须，十进制字符串）。
  - 校验：`Quantity` 数字格式；调用者余额充足；`Recipient` 通过 ID 校验。
  - 返回：两条消息：
    - 发送者的 `Debit-Notice`，Tags：`Ticker`, `Action=Debit-Notice`, `Recipient`, `Quantity`, `TransactionId`；
    - 接收者的 `Credit-Notice`，Tags：`Ticker`, `Action=Credit-Notice`, `Sender`, `Quantity`, `TransactionId`。
  - 缓存：更新 `Balances:<sender>` 与 `Balances:<recipient>`、`TotalSupply`。

- **Mint**（仅限 mintOwner）
  - 描述：给 `Recipient` 增发 `Quantity`。
  - 参数：`Recipient`, `Quantity`（十进制字符串）。
  - 权限：调用者必须等于 `mintOwner`（可由 `owner` 通过 `Set-Params` 配置）。
  - 返回：两条 `Mint-Notice`（发送给 owner 与 recipient），Tags 包含 `Recipient`, `Quantity`, `Ticker`。
  - 缓存：更新 `totalSupply` 和受影响账户的缓存。

#### 仅跨链代币支持的 Actions

- **Burn**
  - 描述：从调用者余额中销毁 `Quantity`，其中 `burnFee` 会转给 `feeRecipient`，实际销毁量为 `Quantity - burnFee`。
  - 参数：`Quantity`（必须，十进制字符串）；`Recipient` 或 `X-Recipient`（可选，不传则默认 `from`）。
  - 校验：`Quantity >= burnFee`；地址通过 ID 校验；余额充足。
  - 返回：一条 `Burn-Notice`，Tags：`Sender`, `X-Recipient`, `Quantity`（=销毁净额），`Ticker`, `TokenId`, `Fee`, `FeeRecipient`。
  - 缓存：更新 `totalSupply`（减去净额）与缓存（`from`、`feeRecipient`）。

### 缓存键（通过 Hymx 节点 HTTP 暴露）

#### 基础代币缓存键
- `TokenInfo`：字符串化 JSON，包含 `Name`, `Ticker`, `Denomination`, `Logo`, `Owner`, `MintOwner`。
- `TotalSupply`：总供应量字符串。
- `Balances`：字符串化 JSON（地址->余额字符串）。
- `Balances:<Account>`：某账户余额字符串。

#### 跨链代币缓存键
- `TokenInfo`：字符串化 JSON，包含 `Name`, `Ticker`, `Denomination`, `Logo`, `Owner`, `MintOwner`, `BurnFee`, `FeeRecipient`。
- `TotalSupply`：总供应量字符串。
- `Balances`：字符串化 JSON（地址->余额字符串）。
- `Balances:<Account>`：某账户余额字符串。

读取示例（见 `example/vmtoken_test.go`）：
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

## 运行服务（CLI）

### 配置示例（`cmd/config.yaml`）
- port: 监听地址，如 `:8080`
- ginMode: `debug` 或 `release`
- redisURL: Redis 连接字符串
- arweaveURL: Arweave 网关
- hymxURL: Hymx 节点 RPC（SDK 用于发送消息）
- prvKey: EVM 私钥（可选；替代 `keyfilePath`）
- keyfilePath: Arweave 密钥文件路径
- nodeName/nodeDesc/nodeURL: 节点元数据
- joinNetwork: 是否加入 Hymx 网络

### 启动命令
```bash
# 前台运行
go build -o hymx ./cmd && ./hymx --config ./cmd/config.yaml

# 守护进程模式
./hymx start --config ./cmd/config.yaml

# 停止守护进程
./hymx stop
```

服务启动后会：
- 用提供的钱包构造 Bundler；
- 挂载两种代币类型：`hymx.basic.token.0.0.1` 到 `SpawnBasicToken` 和 `hymx.crosschain.token.0.0.1` 到 `SpawnCrossChainToken`；
- 提供缓存读取接口 `/cache/<pid>/<key>`。

## 使用 SDK

### 生成基础代币
```go
res, err := hySdk.SpawnAndWait(
    "hymx.basic.token.0.0.1",    // 基础代币的模块格式
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

### 生成跨链代币
```go
res, err := hySdk.SpawnAndWait(
    "hymx.crosschain.token.0.0.1", // 跨链代币的模块格式
    SCHEDULER,                      // 调度器 AccId（通常是本节点地址）
    []goarSchema.Tag{
        {Name: "Name", Value: "跨链代币"},
        {Name: "Ticker", Value: "ccToken"},
        {Name: "Decimals", Value: "18"},
        {Name: "Logo", Value: "<arweave-txid>"},
    },
)
tokenPid := res.Id
```

### 查询信息和余额
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

### 转账、增发、销毁
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
    {Name: "Recipient", Value: hySdk.GetAddress()},
    {Name: "Quantity", Value: "50000000"},
})

// 销毁（仅跨链代币；从调用者；净销毁量 = Quantity - BurnFee）
_, _ = hySdk.SendMessageAndWait(tokenPid, "", []goarSchema.Tag{
    {Name: "Action", Value: "Burn"},
    {Name: "Quantity", Value: "1000"},
    // 可选指定 X-Recipient 或 Recipient 用于链下结算提示
})
```

### 更新参数（仅限 owner）
```go
// 基础代币参数
_, _ = hySdk.SendMessageAndWait(tokenPid, "", []goarSchema.Tag{
    {Name: "Action", Value: "Set-Params"},
    {Name: "MintOwner", Value: "0x..."},   // 允许调用 Mint 的账户
    {Name: "Name", Value: "新名称"},
})

// 跨链代币参数（包含销毁相关参数）
_, _ = hySdk.SendMessageAndWait(tokenPid, "", []goarSchema.Tag{
    {Name: "Action", Value: "Set-Params"},
    {Name: "MintOwner", Value: "0x..."},   // 允许调用 Mint 的账户
    {Name: "BurnFee", Value: "10"},        // 销毁手续费（仅跨链代币）
    {Name: "FeeRecipient", Value: "0x..."}, // 手续费接收者（仅跨链代币）
    {Name: "Name", Value: "新名称"},
})
```

## 地址和数量规则
- 地址：`Recipient/Target` 支持 EVM 或 Arweave 地址；内部通过 `IDCheck` 标准化。
- 数量：所有数量都是表示为十进制字符串的大整数。
- 销毁：要求 `Quantity >= BurnFee`，否则返回 `err_incorrect_quantity`。

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
- 您在构建跨链应用

## 示例和测试
- 见 `example/vmtoken_test.go` 获取从生成到动作调用和缓存读取的完整端到端示例。
- 示例模块 ID 在 `cmd/mod/*.json` 下提供。您的环境可能不同；请使用您实际部署的 ID。

## 开发和构建
```bash
go mod tidy
go build -o hymx ./cmd
```

## 许可证
见仓库中的 `LICENSE`。


