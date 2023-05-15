package order

type Marketplace string

const (
	MarketplaceApecoin Marketplace = "0xa0fc0f7b9714afb29f672bd38a96650a12ff32b9c3fe7884a3416b47961d334d" // keccak256(apecoin)
)

func (m Marketplace) String() string {
	return string(m)
}
