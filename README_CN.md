## hymx-vmtoken

基于 Hymatrix Hymx 的通用 Token VM 实现（模块格式：`aox.token.0.0.1`）。

此项目提供：
- **vmToken 标准与动作接口**：铸造、转账、销毁、查询信息与余额、参数变更等。
- **服务端集成**：将 vmToken 作为 VM 挂载到 Hymx 节点，并实现 token 参数 cache 读取。
- **SDK 使用示例**：使用 `github.com/hymatrix/hymx/sdk` 进行部署与调用 Token (详细测试请参考 example)。

适合需要在 Hymx 上发行和管理同质化代币，或者参考实现自定义 token 的开发者。

### 模块格式（ModuleFormat）
- `schema.VmTokenModuleFormat = "aox.token.0.0.1"`

当 Hymx 节点启动时，会将该模块格式与 `SpawnVmToken` 绑定：
- `s.Mount("aox.token.0.0.1", vmtoken.SpawnVmToken)`

### 实例化（Spawn）参数
实例化 vmToken 时需要传入以下标签（Tags）：
- **Name**: 代币名称（必填）
- **Ticker**: 代币符号（必填）
- **Decimals**: 精度（必填，字符串数值）
- **Logo**: Logo 的 Arweave 资源标识（可选）

实例化后：
- `owner` 为发起者账户（`env.AccId`）。
- 初始 `totalSupply = 0`，`balances = {}`，`burnFee = 0`，`feeRecipient = owner`。
- `mintOwner = owner`（允许执行增发 Mint 的账户；可由 `owner` 修改）。

### Action 与参数规范

- **Info**
  - 描述：返回代币基础信息。
  - 参数：无。
  - 返回 Tags：`Name`, `Ticker`, `Logo`, `Denomination(=Decimals)`。
  - 首次调用会初始化并写入缓存（见「缓存键」）。

- **Set-Params**（仅限 owner）
  - 描述：更新代币与账户参数。
  - 参数（任填其一或多项）：`Owner`, `MintOwner`, `FeeRecipient`, `Name`, `Ticker`, `Decimals`, `Logo`, `BurnFee`（十进制字符串）。
  - 返回 Tags：`Set-Params-Notice = success`。
  - 缓存：更新后会刷新 `TokenInfo` 缓存。

- **Total-Supply / TotalSupply**
  - 描述：查询总供应量。
  - 参数：无。
  - 返回：消息 Data 为十进制字符串的总量；Tags 含 `Action=Total-Supply`, `Ticker`。

- **Balances**
  - 描述：返回全部账户余额快照。
  - 参数：无。
  - 返回：消息 Data 为 `map[string]string` 的 JSON（地址->余额）；Tags 含 `Ticker`。

- **Balance**
  - 描述：查询某账户余额。
  - 参数：可选其一 `Recipient` 或 `Target`，均不传则默认查询调用者账户。
  - 返回 Tags：`Balance`, `Ticker`, `Account`（被查询账户），Data 为余额字符串。

- **Transfer**
  - 描述：从调用者向 `Recipient` 转账。
  - 参数：`Recipient`（必须，支持 EVM/Arweave 地址），`Quantity`（必须，十进制字符串）。
  - 校验：`Quantity` 数字格式；调用者余额充足；`Recipient` 通过 ID 校验。
  - 返回：两条消息：
    - 发送者的 `Debit-Notice`，Tags：`Ticker`, `Action=Debit-Notice`, `Recipient`, `Quantity`；
    - 接收者的 `Credit-Notice`，Tags：`Ticker`, `Action=Credit-Notice`, `Sender`, `Quantity`。
  - 缓存：更新 `Balances:<sender>` 与 `Balances:<recipient>`、`Balances`、`TotalSupply`。

- **Mint**（仅限 mintOwner）
  - 描述：给 `Recipient` 增发 `Quantity`。
  - 参数：`Recipient`, `Quantity`（十进制字符串）。
  - 权限：调用者必须等于 `mintOwner`（可由 `owner` 通过 `Set-Params` 配置）。
  - 返回：两条 `Mint-Notice`（发送给 owner 与 recipient），Tags 包含 `Recipient`, `Quantity`, `Ticker`。
  - 缓存：更新 `totalSupply` 和受影响账户的缓存。

- **Burn**
  - 描述：从调用者余额中销毁 `Quantity`，其中 `burnFee` 会转给 `feeRecipient`，实际销毁量为 `Quantity - burnFee`。
  - 参数：`Quantity`（必须，十进制字符串）；`Recipient` 或 `X-Recipient`（可选，不传则默认 `from`）。
  - 校验：`Quantity >= burnFee`；地址通过 ID 校验；余额充足。
  - 返回：一条 `Burn-Notice`，Tags：`Sender`, `X-Recipient`, `Quantity`（=销毁净额），`Ticker`, `TokenId`, `Fee`, `FeeRecipient`。
  - 缓存：更新 `totalSupply`（减去净额）与缓存（`from`、`feeRecipient`）。


### 缓存键（通过 Hymx 节点 HTTP 暴露）
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

### 配置文件示例（`cmd/config.yaml`）
- **port**: 监听端口，如 `:8080`
- **ginMode**: `debug` 或 `release`
- **redisURL**: Redis 连接串
- **arweaveURL**: Arweave 网关
- **hymxURL**: HymX 节点 RPC（供 SDK 发送消息）
- **prvKey**: EVM 私钥（可选，与 `keyfilePath` 二选一）
- **keyfilePath**: Arweave keyfile 路径
- **nodeName/nodeDesc/nodeURL**: 节点元信息
- **joinNetwork**: 是否加入 Hymx 网络

### 启动命令
```bash
# 直接前台运行
go build -o hymx ./cmd && ./hymx --config ./cmd/config.yaml

# 守护进程模式
./hymx start --config ./cmd/config.yaml

# 停止守护进程
./hymx stop
```

服务启动后会：
- 用提供的钱包构造 Bundler；
- 挂载 `aox.token.0.0.1` 到 `SpawnVmToken`；
- 提供缓存读取接口 `/cache/<pid>/<key>`。

## 通过 SDK 使用

### 实例化 Token
```go
res, err := hySdk.SpawnAndWait(
    MODULE,   // 模块 Id，如示例中的 cmd/mod/*.json 提供的样例 Id
    SCHEDULER,// 调度者 AccId（通常为本节点地址）
    []goarSchema.Tag{
        {Name: "Name", Value: "a Token"},
        {Name: "Ticker", Value: "aToken"},
        {Name: "Decimals", Value: "12"},
        {Name: "Logo", Value: "<arweave-txid>"},
    },
)
tokenPid := res.Id
```

### 查询信息与余额
```go
// Info
_, _ = hySdk.SendMessageAndWait(tokenPid, "", []goarSchema.Tag{{Name: "Action", Value: "Info"}})

// 总供应量
_, _ = hySdk.SendMessageAndWait(tokenPid, "", []goarSchema.Tag{{Name: "Action", Value: "Total-Supply"}})

// 全量余额
_, _ = hySdk.SendMessageAndWait(tokenPid, "", []goarSchema.Tag{{Name: "Action", Value: "Balances"}})

// 指定账户余额（二选一：Recipient/Target）
_, _ = hySdk.SendMessageAndWait(tokenPid, "", []goarSchema.Tag{
    {Name: "Action", Value: "Balance"},
    {Name: "Recipient", Value: "0x..."},
})
```

### 转账、增发、销毁
```go
// Transfer
_, _ = hySdk.SendMessageAndWait(tokenPid, "", []goarSchema.Tag{
    {Name: "Action", Value: "Transfer"},
    {Name: "Recipient", Value: "0x... or ar..."},
    {Name: "Quantity", Value: "100000"},
})

// Mint（仅 mintOwner）
_, _ = hySdk.SendMessageAndWait(tokenPid, "", []goarSchema.Tag{
    {Name: "Action", Value: "Mint"},
    {Name: "Recipient", Value: hySdk.GetAddress()},
    {Name: "Quantity", Value: "50000000"},
})

// Burn（从调用者销毁，净额=Quantity-BurnFee）
_, _ = hySdk.SendMessageAndWait(tokenPid, "", []goarSchema.Tag{
    {Name: "Action", Value: "Burn"},
    {Name: "Quantity", Value: "1000"},
    // 可选指定 X-Recipient 或 Recipient，用于链下收款标识
})
```

### 变更参数（仅 owner）
```go
_, _ = hySdk.SendMessageAndWait(tokenPid, "", []goarSchema.Tag{
    {Name: "Action", Value: "Set-Params"},
    {Name: "MintOwner", Value: "0x..."},          // 配置允许执行 Mint 的账户
    {Name: "BurnFee", Value: "10"},             // 设置燃烧手续费
    {Name: "FeeRecipient", Value: "0x..."},      // 手续费接收者
    {Name: "Name", Value: "NewName"},            // 其他可选字段
})
```

## 地址与数量规则
- 地址：`Recipient/Target` 支持 EVM 或 Arweave 地址，内部通过 `IDCheck` 校验与标准化。
- 数量：所有数量均为十进制字符串表示的大整数。
- 销毁：`Quantity >= BurnFee`，否则会返回 `err_incorrect_quantity`。

## 示例与测试
- 参考 `example/vmtoken_test.go` 获取从实例化到调用各动作、以及通过 `/cache` 读取的完整示例。
- 本仓库 `cmd/mod/*.json` 提供了样例模块 Id，具体环境可能不同，请以实际部署为准。

## 开发与构建
```bash
go mod tidy
go build -o hymx ./cmd
```

## 许可证
遵循仓库内 `LICENSE` 文件。


