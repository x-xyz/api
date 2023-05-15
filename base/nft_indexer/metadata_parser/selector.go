package metadata_parser

import (
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/apecoinstaking"
	"github.com/x-xyz/goapi/domain/collection"
)

type Selector struct {
	defaultParser MetadataParser
	mapping       map[collection.CollectionId]MetadataParser
}

func NewSelector(defaultParser MetadataParser) *Selector {
	return &Selector{
		defaultParser: defaultParser,
		mapping:       make(map[collection.CollectionId]MetadataParser),
	}
}

func (s Selector) Add(collectionId collection.CollectionId, parser MetadataParser) {
	s.mapping[collectionId] = parser
}

func (s *Selector) GetParser(collectionId collection.CollectionId) MetadataParser {
	if parser, ok := s.mapping[collectionId]; ok {
		return parser
	}
	return s.defaultParser
}

func InitializeSelector(s *Selector, webResourceUC domain.WebResourceUseCase, apecoinStakingUC apecoinstaking.UseCase) {
	collections := map[collection.CollectionId]MetadataParser{
		{ChainId: 1, Address: "0x34d85c9cdeb23fa97cb08333b511ac86e1c4e258"}: NewOtherdeedParser(webResourceUC),
		{ChainId: 1, Address: "0x60e4d786628fea6478f785a6d7e704777c86a7c6"}: NewMAYCParser(apecoinStakingUC),
		{ChainId: 1, Address: "0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d"}: NewBAYCParser(apecoinStakingUC),
		{ChainId: 1, Address: "0xba30e5f9bb24caa003e9f2f0497ad287fdf95623"}: NewBAKCParser(apecoinStakingUC),
	}
	for id, parser := range collections {
		s.Add(id, parser)
	}
}
