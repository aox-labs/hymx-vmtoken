# Owner balance removal

`Remove-Balance` deletes the specified account's entire current balance from a
cross-chain token ledger. It does not freeze the account or perform a withdrawal.

Required message tags:

```text
Action: Remove-Balance
Recipient: <account whose balance is to be removed>
SourceTokenId: <source token whose internal lock accounting is reduced>
```

The current token Owner must sign directly; `From-Process` is rejected. No
quantity is needed: the action removes the entire current balance, which must
be positive. The source token must already be configured on this wrapped token.
The Owner is responsible for selecting the correct source token; balances are
fungible, so the module cannot infer their source-chain provenance.

All checks precede mutation. The action deletes the balance entry and reduces
total supply and the selected source token's internal lock amount by the same
amount. It updates existing balances, individual balance, supply and Info cache
entries. It produces no Messages, Spawns or Assignments and never calls Burn.
The result Output reports Account, Removed (base units), and SourceTokenId.

There is no incident state, one-time marker, extra cache key, or checkpoint
schema change. Normal `basic` and `cc` snapshots already contain the balances,
supply and lock accounting. Replaying the removal messages reconstructs the
same state. Repeating removal of an empty account fails without changing state.
If the account later receives funds, the Owner may remove that new balance too.

For the two USDC addresses in this incident, send one message per address with
`SourceTokenId = 0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48`.
At the previously audited balances, the combined reduction is 1,990.000001
wUSDC. Each message is atomic individually; the two messages are not a batch.

Tests use in-memory state only:

```sh
go test -mod=readonly -race ./crosschain
go build -mod=readonly ./basic ./crosschain ./db/cache ./schema
```

This change is maintained in hymx-vmtoken. Publish a reviewed version before
updating hymx-bridge. No deployment, restart or live message is part of this
change. All nodes executing this token must support the new Action: old
binaries silently ignore unknown Actions. Owner key custody remains the
existing authority model.
