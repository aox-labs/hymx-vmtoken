package schema

import "errors"

var (
	ErrMissingRecipient      = errors.New("err_missing_recipient")
	ErrMissingQuantity       = errors.New("err_missing_quantity")
	ErrInvalidQuantityFormat = errors.New("err_invalid_quantity_format")
	ErrIncorrectOwner        = errors.New("err_incorrect_owner")
	ErrInsufficientBalance   = errors.New("err_insufficient_balance")
	ErrInsufficientMaxSupply = errors.New("err_insufficient_max_supply")
	ErrIncorrectTokenInfo    = errors.New("err_incorrect_token_info")
	ErrInvalidRecipient      = errors.New("err_invalid_recipient")
	ErrInvalidFrom           = errors.New("err_invalid_from")
	ErrInvalidMintOwner      = errors.New("err_invalid_mint_owner")
	ErrInvalidMaxSupply      = errors.New("err_invalid_max_supply")
	ErrInvalidOwner          = errors.New("err_invalid_owner")
)
