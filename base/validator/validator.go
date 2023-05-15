package validator

import (
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

// IsValidAddress returns is an address valid or not
func IsValidAddress(address string) bool {
	checksum := common.HexToAddress(address).Hex()
	return strings.ToLower(checksum) == strings.ToLower(address)
}

func NewCustomValidator(v *validator.Validate) echo.Validator {
	return &CustomValidator{v}
}

type CustomValidator struct {
	validator *validator.Validate
}

func (v *CustomValidator) Validate(i interface{}) error {
	if err := v.validator.Struct(i); err != nil {
		return err
	}
	return nil
}
