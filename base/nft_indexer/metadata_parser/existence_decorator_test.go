package metadata_parser

import (
	"testing"

	"github.com/stretchr/testify/require"
	bCtx "github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/domain/nftitem"
)

func Test_existenceDecorator(t *testing.T) {
	ctx := bCtx.Background()
	d := NewExistenceDecorator([]string{"Koda", "Artifact"})
	tests := []struct {
		name       string
		attributes nftitem.Attributes
		appended   map[string]string
	}{
		{
			name: "none",
			attributes: []nftitem.Attribute{
				{TraitType: "Category", Value: "Spirit"},
			},
			appended: map[string]string{
				"Koda?":     "No",
				"Artifact?": "No",
			},
		},
		{
			name: "has artifact",
			attributes: []nftitem.Attribute{
				{TraitType: "Category", Value: "Spirit"},
				{TraitType: "Environment", Value: "Veldan"},
				{TraitType: "Artifact", Value: "Celestial Orb"},
				{TraitType: "Plot", Value: "92534"},
			},
			appended: map[string]string{
				"Koda?":     "No",
				"Artifact?": "Yes",
			},
		},
		{
			name: "has Koda",
			attributes: []nftitem.Attribute{
				{TraitType: "Category", Value: "Spirit"},
				{TraitType: "Environment", Value: "Veldan"},
				{TraitType: "Plot", Value: "92534"},
				{TraitType: "Koda", Value: "7221"},
			},
			appended: map[string]string{
				"Koda?":     "Yes",
				"Artifact?": "No",
			},
		},
		{
			name: "has both",
			attributes: []nftitem.Attribute{
				{TraitType: "Category", Value: "Spirit"},
				{TraitType: "Environment", Value: "Veldan"},
				{TraitType: "Artifact", Value: "Celestial Orb"},
				{TraitType: "Plot", Value: "92534"},
				{TraitType: "Koda", Value: "7221"},
			},
			appended: map[string]string{
				"Koda?":     "Yes",
				"Artifact?": "Yes",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := require.New(t)
			attrs, err := d.Decorate(ctx, tt.attributes)
			req.NoError(err)
			req.Equal(tt.attributes, attrs[:len(tt.attributes)])
			for _, attr := range attrs[len(tt.attributes):] {
				req.Equal(tt.appended[attr.TraitType], attr.Value)
			}
		})
	}
}
