package crosschain

import (
	"encoding/json"
	"math/big"
	"testing"

	"github.com/aox-labs/hymx-vmtoken/schema"
	vmmSchema "github.com/hymatrix/hymx/vmm/schema"
	"github.com/stretchr/testify/require"
)

const (
	testOwner    = "0x1111111111111111111111111111111111111111"
	testHolder   = "0x2222222222222222222222222222222222222222"
	testSource   = "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48"
	testAddressA = "0x074d257e22ECb024Ca44870B07922D0928E82dBA"
	testAddressB = "0xA3957Dad9e0e0A778556eBb2C17A2f64bb3b7503"
)

func newRemoveBalanceToken(t *testing.T) *Token {
	t.Helper()
	vm, err := Spawn(vmmSchema.Env{Meta: vmmSchema.Meta{
		ItemId: "test-token", AccId: testOwner,
		Params: map[string]string{"Name": "USDC", "Ticker": "USDC", "Decimals": "6"},
	}})
	require.NoError(t, err)
	token := vm.(*Token)
	for account, balance := range map[string]int64{testAddressA: 1, testAddressB: 1990000000, testHolder: 12538424619} {
		require.NoError(t, token.basic.DB.UpdateBalance(account, big.NewInt(balance)))
	}
	token.basic.DB.SetTotalSupply(big.NewInt(14528424620))
	token.db.SetSourceTokenChain(testSource, "ethereum")
	token.db.SetSourceLockAmount(testSource, "ethereum", big.NewInt(14528424620))
	token.db.SetMintedRecord("previous-mint", "ethereum")
	return token
}

func removeBalanceMeta(account string) vmmSchema.Meta {
	return vmmSchema.Meta{
		Action: "Remove-Balance", AccId: testOwner,
		Params: map[string]string{"Recipient": account, "SourceTokenId": testSource},
	}
}

func snapshot(t *testing.T, token *Token) string {
	t.Helper()
	data, err := token.Checkpoint()
	require.NoError(t, err)
	return data
}

func noOutbox(t *testing.T, result vmmSchema.Result) {
	t.Helper()
	require.Empty(t, result.Messages)
	require.Empty(t, result.Spawns)
	require.Empty(t, result.Assignments)
}

func TestRemoveBalanceUpdatesLedgerWithoutWithdrawal(t *testing.T) {
	token := newRemoveBalanceToken(t)
	for _, account := range []string{testAddressA, testAddressB} {
		res := token.Apply(testOwner, removeBalanceMeta(account))
		require.NoError(t, res.Error)
		noOutbox(t, res)
		require.Equal(t, "0", res.Cache["balances:"+account])
		var balances map[string]*big.Int
		require.NoError(t, json.Unmarshal([]byte(res.Cache["balances"]), &balances))
		require.NotContains(t, balances, account)
		require.Equal(t, token.basic.DB.GetTotalSupply().String(), res.Cache["total-supply"])
		var info map[string]string
		require.NoError(t, json.Unmarshal([]byte(res.Cache["info"]), &info))
		lock, ok := token.db.GetSourceLockAmount(testSource, "ethereum")
		require.True(t, ok)
		require.JSONEq(t, `{"ethereum:`+testSource+`":`+lock.String()+`}`, info["SourceLockAmounts"])
	}
	balances, err := token.basic.DB.Balances()
	require.NoError(t, err)
	require.Len(t, balances, 1)
	require.Equal(t, "12538424619", balances[testHolder].String())
	require.Equal(t, "12538424619", token.basic.DB.GetTotalSupply().String())
	lock, _ := token.db.GetSourceLockAmount(testSource, "ethereum")
	require.Equal(t, "12538424619", lock.String())
	_, exists := token.db.GetMintedRecord("previous-mint")
	require.True(t, exists)
	var parts map[string]string
	require.NoError(t, json.Unmarshal([]byte(snapshot(t, token)), &parts))
	require.Len(t, parts, 2, "checkpoint remains basic + cc only")
	require.Contains(t, parts, "basic")
	require.Contains(t, parts, "cc")
}

func TestRemoveBalanceRejectsInvalidRequestsWithoutStateChanges(t *testing.T) {
	tests := []struct {
		name   string
		change func(*Token, *vmmSchema.Meta, *string)
		want   error
	}{
		{"non-owner", func(_ *Token, m *vmmSchema.Meta, f *string) { *f = testHolder; m.AccId = testHolder }, schema.ErrIncorrectOwner},
		{"spoofed-owner", func(_ *Token, m *vmmSchema.Meta, _ *string) { m.AccId = testHolder }, schema.ErrIncorrectOwner},
		{"from-process", func(_ *Token, m *vmmSchema.Meta, _ *string) { m.FromProcess = testOwner }, schema.ErrIncorrectOwner},
		{"missing-account", func(_ *Token, m *vmmSchema.Meta, _ *string) { delete(m.Params, "Recipient") }, schema.ErrMissingRecipient},
		{"invalid-account", func(_ *Token, m *vmmSchema.Meta, _ *string) { m.Params["Recipient"] = "invalid" }, schema.ErrInvalidRecipient},
		{"missing-source", func(_ *Token, m *vmmSchema.Meta, _ *string) { delete(m.Params, "SourceTokenId") }, schema.ErrMissingSourceTokenId},
		{"invalid-source", func(_ *Token, m *vmmSchema.Meta, _ *string) { m.Params["SourceTokenId"] = "invalid" }, schema.ErrInvalidSourceTokenId},
		{"unknown-source", func(_ *Token, m *vmmSchema.Meta, _ *string) { m.Params["SourceTokenId"] = testHolder }, schema.ErrInvalidSourceTokenId},
		{"no-balance", func(_ *Token, m *vmmSchema.Meta, _ *string) { m.Params["Recipient"] = testOwner }, schema.ErrInsufficientBalance},
		{"insufficient-lock", func(v *Token, _ *vmmSchema.Meta, _ *string) {
			v.db.SetSourceLockAmount(testSource, "ethereum", big.NewInt(0))
		}, schema.ErrInsufficientLockAmount},
		{"insufficient-supply", func(v *Token, _ *vmmSchema.Meta, _ *string) { v.basic.DB.SetTotalSupply(big.NewInt(0)) }, schema.ErrIncorrectQuantity},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			token := newRemoveBalanceToken(t)
			meta := removeBalanceMeta(testAddressA)
			from := testOwner
			tc.change(token, &meta, &from)
			before := snapshot(t, token)
			res := token.Apply(from, meta)
			require.ErrorIs(t, res.Error, tc.want)
			noOutbox(t, res)
			require.Empty(t, res.Cache)
			require.Equal(t, before, snapshot(t, token))
		})
	}
}

func TestRemoveBalanceRestoreReplayAndRepeat(t *testing.T) {
	token := newRemoveBalanceToken(t)
	legacy := snapshot(t, token)
	for _, account := range []string{testAddressA, testAddressB} {
		res := token.Apply(testOwner, removeBalanceMeta(account))
		require.NoError(t, res.Error)
	}
	removed := snapshot(t, token)
	restored := newRemoveBalanceToken(t)
	require.NoError(t, restored.Restore(removed))
	for _, account := range []string{testAddressA, testAddressB} {
		res := restored.Apply(testOwner, removeBalanceMeta(account))
		require.ErrorIs(t, res.Error, schema.ErrInsufficientBalance)
		noOutbox(t, res)
	}
	require.Equal(t, removed, snapshot(t, restored))
	for _, mode := range []vmmSchema.ExecMode{vmmSchema.ExecModeDryRun, vmmSchema.ExecModeReplay} {
		require.NoError(t, restored.Restore(legacy))
		for _, account := range []string{testAddressA, testAddressB} {
			meta := removeBalanceMeta(account)
			meta.Mode = mode
			res := restored.Apply(testOwner, meta)
			require.NoError(t, res.Error)
			noOutbox(t, res)
		}
		require.Equal(t, removed, snapshot(t, restored))
	}
}

func TestOwnerCanRemoveAnySpecifiedAccount(t *testing.T) {
	token := newRemoveBalanceToken(t)
	res := token.Apply(testOwner, removeBalanceMeta(testHolder))
	require.NoError(t, res.Error)
	noOutbox(t, res)
	balances, err := token.basic.DB.Balances()
	require.NoError(t, err)
	require.NotContains(t, balances, testHolder)
	require.Equal(t, "1", balances[testAddressA].String())
	require.Equal(t, "1990000000", balances[testAddressB].String())
}
