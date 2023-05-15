package validator

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type ValidatorTestSuite struct {
	suite.Suite
}

func (s *ValidatorTestSuite) SetupTest() {
}

func (s *ValidatorTestSuite) TearDownTest() {
}

func (s *ValidatorTestSuite) SetupSuite() {
}

func (s *ValidatorTestSuite) TearDownSuite() {
}

func (s *ValidatorTestSuite) TestIsValidAddress() {
	tests := []struct {
		desc       string
		address    string
		expIsValid bool
	}{
		{
			desc:       "invalid address",
			address:    "0x000",
			expIsValid: false,
		},
		{
			desc:       "valid address - real address",
			address:    "0x939ae6A4C8dfDBB1f7085189574F0A938013952A",
			expIsValid: true,
		},
		{
			desc:       "valid address - lower case",
			address:    "0x939ae6a4c8dfdbb1f7085189574f0a938013952b",
			expIsValid: true,
		},
	}
	for _, t := range tests {
		s.Equal(t.expIsValid, IsValidAddress(t.address), t.desc)
	}
}

func TestValidatorTestSuite(t *testing.T) {
	suite.Run(t, new(ValidatorTestSuite))
}
