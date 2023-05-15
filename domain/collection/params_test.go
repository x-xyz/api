package collection

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSortCollectionWithHoldingCount(t *testing.T) {
	cases := []struct {
		name       string
		sortOption SearchSortOption
		data       []*CollectionWithHoldingCount
		want       []*CollectionWithHoldingCount
	}{
		{
			name:       "without sort",
			sortOption: "",
			data: []*CollectionWithHoldingCount{
				{
					Collection: Collection{
						CollectionName: "collection1",
					},
				},
				{
					Collection: Collection{
						CollectionName: "collection2",
					},
				},
			},
			want: []*CollectionWithHoldingCount{
				{
					Collection: Collection{
						CollectionName: "collection1",
					},
				},
				{
					Collection: Collection{
						CollectionName: "collection2",
					},
				},
			},
		},
		{
			name:       "sort by SearchSortOptionHoldingAsc",
			sortOption: SearchSortOptionHoldingAsc,
			data: []*CollectionWithHoldingCount{
				{
					Collection: Collection{
						CollectionName: "collection2",
					},
					HoldingCount:   200,
					HoldingBalance: 200,
				},
				{
					Collection: Collection{
						CollectionName: "collection1",
					},
					HoldingCount:   100,
					HoldingBalance: 100,
				},
			},
			want: []*CollectionWithHoldingCount{
				{
					Collection: Collection{
						CollectionName: "collection1",
					},
					HoldingCount:   100,
					HoldingBalance: 100,
				},
				{
					Collection: Collection{
						CollectionName: "collection2",
					},
					HoldingCount:   200,
					HoldingBalance: 200,
				},
			},
		},
		{
			name:       "sort by SearchSortOptionHoldingDesc",
			sortOption: SearchSortOptionHoldingDesc,
			data: []*CollectionWithHoldingCount{
				{
					Collection: Collection{
						CollectionName: "collection1",
					},
					HoldingCount:   100,
					HoldingBalance: 100,
				},
				{
					Collection: Collection{
						CollectionName: "collection2",
					},
					HoldingCount:   200,
					HoldingBalance: 200,
				},
			},
			want: []*CollectionWithHoldingCount{
				{
					Collection: Collection{
						CollectionName: "collection2",
					},
					HoldingCount:   200,
					HoldingBalance: 200,
				},
				{
					Collection: Collection{
						CollectionName: "collection1",
					},
					HoldingCount:   100,
					HoldingBalance: 100,
				},
			},
		},
	}

	for _, c := range cases {
		cpy := make([]*CollectionWithHoldingCount, len(c.data))
		copy(cpy, c.data)

		SortCollectionWithHoldingCount(cpy, c.sortOption)
		assert.Equal(t, c.want, cpy)
	}
}
