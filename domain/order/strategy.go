package order

type Strategy string

const (
	StrategyFixedPrice      Strategy = "fixedPrice"
	StrategyPrivateSale     Strategy = "privateSale"
	StrategyCollectionOffer Strategy = "collectionOffer"
	StrategyUnknown         Strategy = "unknown"
)

func ToStrategy(name string) Strategy {
	switch name {
	case string(StrategyFixedPrice):
		return StrategyFixedPrice
	case string(StrategyPrivateSale):
		return StrategyPrivateSale
	case string(StrategyCollectionOffer):
		return StrategyCollectionOffer
	}
	return StrategyUnknown
}
