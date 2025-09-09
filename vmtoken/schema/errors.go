package schema

import "errors"

var (
	ErrMissingRecipient      = errors.New("err_missing_recipient")
	ErrMissingQuantity       = errors.New("err_missing_quantity")
	ErrInvalidQuantityFormat = errors.New("err_invalid_quantity_format")
	ErrIncorrectOwner        = errors.New("err_incorrect_owner")
	ErrInsufficientBalance   = errors.New("err_insufficient_balance")
	ErrIncorrectQuantity     = errors.New("err_incorrect_quantity")
	ErrIncorrectTokenInfo    = errors.New("err_incorrect_token_info")
	ErrUnsupportedAction     = errors.New("err_unsupported_action")
	ErrInvalidFeeRecipient   = errors.New("err_invalid_fee_recipient")
	ErrInvalidMintOwner      = errors.New("err_invalid_mint_owner")

	// Multi-chain specific errors
	ErrMissingSourceChain       = errors.New("err_missing_source_chain")
	ErrIncorrectSourceChainType = errors.New("err_incorrect_source_chain_type")

	ErrMissingSourceTokenId   = errors.New("err_missing_source_token_id")
	ErrMissingTargetChain     = errors.New("err_missing_target_chain")
	ErrMissingTargetTokenId   = errors.New("err_missing_target_token_id")
	ErrIncorrectTargetTokenId = errors.New("err_incorrect_target_token_id")
	ErrLockAmountEmpty        = errors.New("err_lock_amount_empty")
	ErrInsufficientLockAmount = errors.New("err_insufficient_lock_amount")
	ErrInvalidChainType       = errors.New("err_invalid_chain_type")
	ErrMissingBurnFee         = errors.New("err_missing_burn_fee")
)
