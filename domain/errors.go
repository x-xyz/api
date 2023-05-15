package domain

import "errors"

var (
	// ErrInternalServerError will throw if any the Internal Server Error happen
	ErrInternalServerError = errors.New("Internal Server Error")
	// ErrNotFound will throw if the requested item is not exists
	ErrNotFound = errors.New("Your requested Item is not found")
	// ErrConflict will throw if the current action already exists
	ErrConflict = errors.New("Your Item already exist")
	// ErrBadParamInput will throw if the given request-body or params is not valid
	ErrBadParamInput               = errors.New("Given Param is not valid")
	ErrUnsupportedSchema           = errors.New("Unsupported schema")
	ErrUnimplemented               = errors.New("Unimplemented")
	ErrInvalidJsonFormat           = errors.New("invalid JSON format")
	ErrErc721InterfaceUnsupported  = errors.New("erc721 interface unsupported")
	ErrErc1155InterfaceUnsupported = errors.New("erc1155 interface unsupported")
	ErrInvalidNumberFormat         = errors.New("invalid number format")
	ErrInvalidChainId              = errors.New("invalid chain id")
	ErrInvalidStrategy             = errors.New("invalid strategy")
	ErrInvalidOrderNonce           = errors.New("invalid order nonce")
	ErrInvalidOrderSideForStrategy = errors.New("invalid order side for strategy")
	ErrInvalidCurrency             = errors.New("invalid currency")

	// request error
	ErrInvalidAddress   = errors.New("Invalid address")
	ErrInvalidSignature = errors.New("Invalid signature")

	ErrDeprecated     = errors.New("deprecated")
	ErrNotImplemented = errors.New("not implemented")
)
