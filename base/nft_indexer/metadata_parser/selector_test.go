package metadata_parser

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/x-xyz/goapi/domain/collection"
)

func TestSelector(t *testing.T) {
	defaultParser := NewDefaultParser()
	otherdeedParser := NewOtherdeedParser(nil)
	otherdeedColId := collection.CollectionId{ChainId: 1, Address: "0x34d85c9cdeb23fa97cb08333b511ac86e1c4e258"}

	s := NewSelector(defaultParser)
	s.Add(otherdeedColId, otherdeedParser)

	tests := []struct {
		name         string
		collectionId collection.CollectionId
		parser       MetadataParser
	}{
		{
			name:         "otherdeed",
			collectionId: otherdeedColId,
			parser:       otherdeedParser,
		},
		{
			name:         "some collection",
			collectionId: collection.CollectionId{ChainId: 1, Address: "0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d"},
			parser:       defaultParser,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := require.New(t)
			parser := s.GetParser(tt.collectionId)
			req.Equal(tt.parser, parser)
		})
	}

}
