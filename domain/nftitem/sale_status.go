package nftitem

type SaleStatus uint8

type SaleStatusType string

const (
	SaleStatusDefault SaleStatus = 0
	SaleStatusBuyNow  SaleStatus = 1 << iota
	SaleStatusHasBid
	SaleStatusHasOffer
	SaleStatusOnAuction
	SaleStatusHasTraded
	SaleStatusHasOfferWithExpired

	SaleStatusTypeBuyNow              = "buynow"
	SaleStatusTypeHasBid              = "hasbid"
	SaleStatusTypeHasOffer            = "hasoffer"
	SaleStatusTypeOnAuction           = "onauction"
	SaleStatusTypeHasTraded           = "hastraded"
	SaleStatusTypeHasOfferWithExpired = "hasofferwithexpired"
)

func HasSaleStatus(status, flag SaleStatus) bool {
	return status&flag != 0
}

func SetSaleStatus(status, flag SaleStatus) SaleStatus {
	return status | flag
}

func ParseSaleStatusType(types ...SaleStatusType) (saleStatus SaleStatus) {
	saleStatus = SaleStatusDefault
	for _, item := range types {
		switch item {
		case SaleStatusTypeBuyNow:
			saleStatus = SetSaleStatus(saleStatus, SaleStatusBuyNow)
		case SaleStatusTypeHasBid:
			saleStatus = SetSaleStatus(saleStatus, SaleStatusHasBid)
		case SaleStatusTypeHasOffer:
			saleStatus = SetSaleStatus(saleStatus, SaleStatusHasOffer)
		case SaleStatusTypeOnAuction:
			saleStatus = SetSaleStatus(saleStatus, SaleStatusOnAuction)
		case SaleStatusTypeHasTraded:
			saleStatus = SetSaleStatus(saleStatus, SaleStatusHasTraded)
		case SaleStatusTypeHasOfferWithExpired:
			saleStatus = SetSaleStatus(saleStatus, SaleStatusHasOfferWithExpired)
		}
	}
	return
}
