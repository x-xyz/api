package metadata_parser

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	bCtx "github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/domain"
	webresource_repository "github.com/x-xyz/goapi/stores/web_resource/repository"
	webresource_usecase "github.com/x-xyz/goapi/stores/web_resource/usecase"
)

func Test_otherdeedParser(t *testing.T) {
	ctx := bCtx.Background()
	httpRepo := webresource_repository.NewHttpReaderRepo(http.Client{}, 2*time.Second, nil)

	webResourceUseCase := webresource_usecase.NewWebResourceUseCase(&webresource_usecase.WebResourceUseCaseCfg{
		HttpReader: httpRepo,
	})
	p := NewOtherdeedParser(webResourceUseCase)
	tests := []struct {
		name    string
		tokenId string
		want    map[string]string
	}{
		{
			name:    "koda",
			tokenId: "53690",
			want: map[string]string{
				"Category":               "\"Decay\"",
				"Sediment":               "\"Infinite Expanse\"",
				"Sediment Tier":          "2",
				"Environment":            "\"Bog\"",
				"Environment Tier":       "3",
				"Northern Resource":      "\"Psychosilk\"",
				"Northern Resource Tier": "1",
				"Artifact":               "\"Sands of Time\"",
				"Koda":                   "7221",
				"Plot":                   "53690",
				"Obelisk Piece":          "\"First Trip\"",
				"Artifact?":              "Yes",
				"Koda?":                  "Yes",
				"Koda - Head":            "\"Viridescent Hero\"",
				"Koda - Eyes":            "\"Release Me\"",
				"Koda - Core":            "\"Cushioned\"",
				"Koda - Clothing":        "\"Daybreak Armor\"",
				"Koda - Clothing?":       "Yes",
				"Koda - Mega?":           "No",
				"Koda - Weapon?":         "No",
			},
		}, {
			name:    "non-koda",
			tokenId: "39596",
			want: map[string]string{
				"Category":         "\"Harsh\"",
				"Sediment":         "\"Cosmic Dream\"",
				"Sediment Tier":    "3",
				"Environment":      "\"Splinter\"",
				"Environment Tier": "5",
				"Plot":             "39596",
				"Artifact?":        "No",
				"Koda?":            "No",
			},
		}, {
			name:    "mega koda",
			tokenId: "55297",
			want: map[string]string{
				"Category":              "\"Spirit\"",
				"Sediment":              "\"Infinite Expanse\"",
				"Sediment Tier":         "1",
				"Environment":           "\"Veldan\"",
				"Environment Tier":      "1",
				"Eastern Resource":      "\"Brimstone\"",
				"Eastern Resource Tier": "2",
				"Western Resource":      "\"Psychosilk\"",
				"Western Resource Tier": "1",
				"Artifact":              "\"Camp Fire\"",
				"Koda":                  "9977",
				"Plot":                  "55297",
				"Artifact?":             "Yes",
				"Koda?":                 "Yes",
				"Koda - Head":           "\"Shrine Guardian\"",
				"Koda - Eyes":           "\"Eyevestigation\"",
				"Koda - Core":           "\"Spirit Slab\"",
				"Koda - Clothing":       "\"Ancient Orb Armor\"",
				"Koda - Weapon":         "\"Blade of Great Heights\"",
				"Koda - Clothing?":      "Yes",
				"Koda - Weapon?":        "Yes",
				"Koda - Mega?":          "Yes",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := require.New(t)
			url := fmt.Sprintf("https://api.otherside.xyz/lands/%s", tt.tokenId)
			data, err := webResourceUseCase.GetJson(ctx, url)
			req.NoError(err)
			attrs, err := p.Parse(ctx, 1, domain.Address(""), domain.TokenId(tt.tokenId), data)
			req.NoError(err)
			for _, attr := range attrs {
				req.Equal(tt.want[attr.TraitType], attr.Value)
			}
		})
	}
}
