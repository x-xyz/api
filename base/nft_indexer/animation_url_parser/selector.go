package animation_url_parser

import (
	"github.com/x-xyz/goapi/domain/collection"
)

type Selector struct {
	defaultParser Parser
	mapping       map[collection.CollectionId]Parser
}

func NewSelector(defaultParser Parser) *Selector {
	return &Selector{
		defaultParser: defaultParser,
		mapping:       make(map[collection.CollectionId]Parser),
	}
}

func (s Selector) Add(collectionId collection.CollectionId, parser Parser) {
	s.mapping[collectionId] = parser
}

func (s *Selector) GetParser(collectionId collection.CollectionId) Parser {
	if parser, ok := s.mapping[collectionId]; ok {
		return parser
	}
	return s.defaultParser
}

func InitializeSelector(s *Selector) {
	collections := map[collection.CollectionId]Parser{
		{ChainId: 1, Address: "0xa1d4657e0e6507d5a94d06da93e94dc7c8c44b51"}: NewNullParser(),
		{ChainId: 1, Address: "0xeda3b617646b5fc8c9c696e0356390128ce900f8"}: NewNullParser(),
	}
	for id, parser := range collections {
		s.Add(id, parser)
	}
}
