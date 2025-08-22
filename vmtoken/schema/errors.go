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
)
