package cache

import (
	"encoding/json"
	"math/big"
	"sync"

	dbSchema "github.com/aox-labs/hymx-vmtoken/db/cache/schema"
)

type CrossChainToken struct {
	mintedRecords     map[string]string   // key: X-MintTxHash val: chainType
	sourceTokenChains map[string]string   // key: sourceTokenId, val: sourceChainType
	sourceLockAmounts map[string]*big.Int // key: sourceChain:sourceTokenId, val: source chain locked amount
	burnFees          map[string]*big.Int // key: chainType, val: burn fee
	feeRecipient      string
	burnProcessor     string
	rwlock            sync.RWMutex
}

func NewCrossChainToken(burnFees map[string]*big.Int, feeRecipient string, burnProcessor string) *CrossChainToken {
	return &CrossChainToken{
		mintedRecords:     make(map[string]string),
		sourceTokenChains: make(map[string]string),
		sourceLockAmounts: make(map[string]*big.Int),
		burnFees:          burnFees,
		feeRecipient:      feeRecipient,
		burnProcessor:     burnProcessor,
		rwlock:            sync.RWMutex{},
	}
}

// GetMintedRecord gets the source chain type for a mint transaction hash
func (c *CrossChainToken) GetMintedRecord(mintTxHash string) (string, bool) {
	c.rwlock.RLock()
	defer c.rwlock.RUnlock()
	chainType, exists := c.mintedRecords[mintTxHash]
	return chainType, exists
}

// SetMintedRecord sets the source chain type for a mint transaction hash
func (c *CrossChainToken) SetMintedRecord(mintTxHash, sourceChainType string) {
	c.rwlock.Lock()
	defer c.rwlock.Unlock()
	if c.mintedRecords == nil {
		c.mintedRecords = make(map[string]string)
	}
	c.mintedRecords[mintTxHash] = sourceChainType
}

// GetSourceTokenChains returns a copy of all source token chains
func (c *CrossChainToken) GetSourceTokenChains() map[string]string {
	c.rwlock.RLock()
	defer c.rwlock.RUnlock()
	result := make(map[string]string, len(c.sourceTokenChains))
	for k, v := range c.sourceTokenChains {
		result[k] = v
	}
	return result
}

// GetSourceTokenChain gets the chain type for a source token ID
func (c *CrossChainToken) GetSourceTokenChain(tokenId string) (string, bool) {
	c.rwlock.RLock()
	defer c.rwlock.RUnlock()
	chainType, exists := c.sourceTokenChains[tokenId]
	return chainType, exists
}

// SetSourceTokenChain sets the chain type for a source token ID
func (c *CrossChainToken) SetSourceTokenChain(tokenId, chainType string) {
	c.rwlock.Lock()
	defer c.rwlock.Unlock()
	if c.sourceTokenChains == nil {
		c.sourceTokenChains = make(map[string]string)
	}
	c.sourceTokenChains[tokenId] = chainType
}

// GetSourceLockAmounts returns a copy of all source lock amounts
func (c *CrossChainToken) GetSourceLockAmounts() map[string]*big.Int {
	c.rwlock.RLock()
	defer c.rwlock.RUnlock()
	result := make(map[string]*big.Int, len(c.sourceLockAmounts))
	for k, v := range c.sourceLockAmounts {
		if v == nil {
			result[k] = big.NewInt(0)
		} else {
			result[k] = new(big.Int).Set(v)
		}
	}
	return result
}

// GetSourceLockAmount gets the locked amount for a source token ID and chain type
func (c *CrossChainToken) GetSourceLockAmount(tokenId, chainType string) (*big.Int, bool) {
	key := chainType + ":" + tokenId
	c.rwlock.RLock()
	defer c.rwlock.RUnlock()
	amount, exists := c.sourceLockAmounts[key]
	if !exists || amount == nil {
		return big.NewInt(0), exists
	}
	return new(big.Int).Set(amount), exists
}

// SetSourceLockAmount sets the locked amount for a source token ID and chain type
func (c *CrossChainToken) SetSourceLockAmount(tokenId, chainType string, amount *big.Int) {
	key := chainType + ":" + tokenId
	c.rwlock.Lock()
	defer c.rwlock.Unlock()
	if c.sourceLockAmounts == nil {
		c.sourceLockAmounts = make(map[string]*big.Int)
	}
	if amount == nil {
		amount = big.NewInt(0)
	}
	c.sourceLockAmounts[key] = new(big.Int).Set(amount)
}

// GetBurnFees returns a copy of all burn fees
func (c *CrossChainToken) GetBurnFees() map[string]*big.Int {
	c.rwlock.RLock()
	defer c.rwlock.RUnlock()
	result := make(map[string]*big.Int, len(c.burnFees))
	for k, v := range c.burnFees {
		if v == nil {
			result[k] = big.NewInt(0)
		} else {
			result[k] = new(big.Int).Set(v)
		}
	}
	return result
}

// GetBurnFee gets the burn fee for a chain type
func (c *CrossChainToken) GetBurnFee(chainType string) (*big.Int, bool) {
	c.rwlock.RLock()
	defer c.rwlock.RUnlock()
	fee, exists := c.burnFees[chainType]
	if !exists || fee == nil {
		return big.NewInt(0), exists
	}
	return new(big.Int).Set(fee), exists
}

// SetBurnFee sets the burn fee for a chain type
func (c *CrossChainToken) SetBurnFee(chainType string, amount *big.Int) {
	c.rwlock.Lock()
	defer c.rwlock.Unlock()
	if c.burnFees == nil {
		c.burnFees = make(map[string]*big.Int)
	}
	if amount == nil {
		amount = big.NewInt(0)
	}
	c.burnFees[chainType] = new(big.Int).Set(amount)
}

// GetFeeRecipient gets the fee recipient address
func (c *CrossChainToken) GetFeeRecipient() string {
	c.rwlock.RLock()
	defer c.rwlock.RUnlock()
	return c.feeRecipient
}

// SetFeeRecipient sets the fee recipient address
func (c *CrossChainToken) SetFeeRecipient(addr string) {
	c.rwlock.Lock()
	defer c.rwlock.Unlock()
	c.feeRecipient = addr
}

// GetBurnProcessor gets the burn processor address
func (c *CrossChainToken) GetBurnProcessor() string {
	c.rwlock.RLock()
	defer c.rwlock.RUnlock()
	return c.burnProcessor
}

// SetBurnProcessor sets the burn processor address
func (c *CrossChainToken) SetBurnProcessor(addr string) {
	c.rwlock.Lock()
	defer c.rwlock.Unlock()
	c.burnProcessor = addr
}

// Checkpoint creates a snapshot of the cross-chain token state
func (c *CrossChainToken) Checkpoint() (data string, err error) {
	c.rwlock.RLock()
	defer c.rwlock.RUnlock()
	snap := dbSchema.CrossChainMultiSnapshot{
		MintedRecords:     c.mintedRecords,
		SourceTokenChains: c.sourceTokenChains,
		SourceLockAmounts: c.sourceLockAmounts,
		BurnFees:          c.burnFees,
		FeeRecipient:      c.feeRecipient,
		BurnProcessor:     c.burnProcessor,
	}
	by, err := json.Marshal(snap)
	if err != nil {
		return "", err
	}
	return string(by), nil
}

// Restore restores the cross-chain token state from a snapshot
func (c *CrossChainToken) Restore(data string) error {
	snap := &dbSchema.CrossChainMultiSnapshot{}
	if err := json.Unmarshal([]byte(data), snap); err != nil {
		return err
	}
	c.rwlock.Lock()
	defer c.rwlock.Unlock()
	c.mintedRecords = snap.MintedRecords
	if c.mintedRecords == nil {
		c.mintedRecords = make(map[string]string)
	}
	c.sourceTokenChains = snap.SourceTokenChains
	if c.sourceTokenChains == nil {
		c.sourceTokenChains = make(map[string]string)
	}
	c.sourceLockAmounts = snap.SourceLockAmounts
	if c.sourceLockAmounts == nil {
		c.sourceLockAmounts = make(map[string]*big.Int)
	}
	c.burnFees = snap.BurnFees
	if c.burnFees == nil {
		c.burnFees = make(map[string]*big.Int)
	}
	c.feeRecipient = snap.FeeRecipient
	c.burnProcessor = snap.BurnProcessor
	return nil
}
