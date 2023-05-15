package order

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/x-xyz/goapi/domain"
)

func init() {
	var err error
	boolTy, err = abi.NewType("bool", "", nil)
	if err != nil {
		panic(err)
	}
	addressTy, err = abi.NewType("address", "", nil)
	if err != nil {
		panic(err)
	}
	uintTy, err = abi.NewType("uint256", "", nil)
	if err != nil {
		panic(err)
	}
	bytesTy, err = abi.NewType("bytes", "", nil)
	if err != nil {
		panic(err)
	}
	tupleTy, err = abi.NewType("tuple", "", []abi.ArgumentMarshaling{
		{Name: "collection", Type: "address"},
		{Name: "tokenId", Type: "uint256"},
		{Name: "amount", Type: "uint256"},
		{Name: "price", Type: "uint256"},
	})
	if err != nil {
		panic(err)
	}
}

var (
	boolTy    abi.Type
	addressTy abi.Type
	uintTy    abi.Type
	bytesTy   abi.Type
	tupleTy   abi.Type
)

const (
	PrimaryType      = "MakerOrder"
	Eip712DomainName = "EIP712Domain"
)

func GetDomainSeperator(chainId domain.ChainId, address domain.Address) apitypes.TypedDataDomain {
	return apitypes.TypedDataDomain{
		Name:              "XExchange",
		Version:           "1",
		ChainId:           math.NewHexOrDecimal256(int64(chainId)),
		VerifyingContract: address.ToLowerStr(),
	}
}

var OrderTypes = apitypes.Types{
	"MakerOrder": {
		{Name: "isAsk", Type: "bool"},
		{Name: "signer", Type: "address"},
		{Name: "items", Type: "OrderItem[]"},
		{Name: "strategy", Type: "address"},
		{Name: "currency", Type: "address"},
		{Name: "nonce", Type: "uint256"},
		{Name: "startTime", Type: "uint256"},
		{Name: "endTime", Type: "uint256"},
		{Name: "minPercentageToAsk", Type: "uint256"},
		{Name: "marketplace", Type: "bytes32"},
		{Name: "params", Type: "bytes"},
	},
	"OrderItem": {
		{Name: "collection", Type: "address"},
		{Name: "tokenId", Type: "uint256"},
		{Name: "amount", Type: "uint256"},
		{Name: "price", Type: "uint256"},
	},
	"EIP712Domain": {
		{Name: "name", Type: "string"},
		{Name: "version", Type: "string"},
		{Name: "chainId", Type: "uint256"},
		{Name: "verifyingContract", Type: "address"},
	},
}

func (i *Item) ToMessage() apitypes.TypedDataMessage {
	return apitypes.TypedDataMessage{
		"collection": i.Collection.ToLowerStr(),
		"tokenId":    i.TokenId.String(),
		"amount":     i.Amount,
		"price":      i.Price,
	}
}

func (o *Order) ToMessage() apitypes.TypedDataMessage {
	items := []interface{}{}
	for _, item := range o.Items {
		items = append(items, item.ToMessage())
	}
	return apitypes.TypedDataMessage{
		"isAsk":              o.IsAsk,
		"signer":             o.Signer.ToLowerStr(),
		"items":              items,
		"strategy":           o.Strategy.ToLowerStr(),
		"currency":           o.Currency.ToLowerStr(),
		"nonce":              o.Nonce,
		"startTime":          o.StartTime,
		"endTime":            o.EndTime,
		"minPercentageToAsk": o.MinPercentageToAsk,
		"marketplace":        o.Marketplace,
		"params":             o.Params,
	}
}

func (o *Order) Hash() ([]byte, error) {
	typedData := apitypes.TypedData{
		Types:       OrderTypes,
		PrimaryType: PrimaryType,
		Domain:      GetDomainSeperator(1, domain.EmptyAddress), // dummy
		Message:     o.ToMessage(),
	}

	return typedData.HashStruct(typedData.PrimaryType, typedData.Message)
}

func (o *Order) HashOrderItem(idx int) ([]byte, error) {
	item := o.Items[idx]
	nums, err := domain.ToBigInt([]string{o.Nonce, o.StartTime, o.EndTime, o.MinPercentageToAsk, item.TokenId.String(), item.Amount, item.Price})
	if err != nil {
		return nil, err
	}
	orderItem := struct {
		Collection common.Address
		TokenId    *big.Int
		Amount     *big.Int
		Price      *big.Int
	}{
		common.HexToAddress(item.Collection.ToLowerStr()),
		nums[4],
		nums[5],
		nums[6],
	}
	params, err := hexutil.Decode(o.Params)
	if err != nil {
		return nil, err
	}
	args := abi.Arguments{
		{Type: boolTy},
		{Type: addressTy},
		{Type: uintTy},
		{Type: tupleTy},
		{Type: addressTy},
		{Type: addressTy},
		{Type: uintTy},
		{Type: uintTy},
		{Type: uintTy},
		{Type: uintTy},
		{Type: bytesTy},
	}
	encoded, err := args.Pack(
		o.IsAsk,
		common.HexToAddress(o.Signer.ToLowerStr()),
		big.NewInt(int64(idx)),
		orderItem,
		common.HexToAddress(o.Strategy.ToLowerStr()),
		common.HexToAddress(o.Currency.ToLowerStr()),
		nums[0], // nonce
		nums[1], // startTime
		nums[2], // endTime
		nums[3], // minPercentageToAsk
		params,
	)
	if err != nil {
		return nil, err
	}
	return crypto.Keccak256(encoded), nil
}
