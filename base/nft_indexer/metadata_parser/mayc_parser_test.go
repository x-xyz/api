package metadata_parser

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	bCtx "github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/apecoinstaking/mocks"
	webresource_repository "github.com/x-xyz/goapi/stores/web_resource/repository"
	webresource_usecase "github.com/x-xyz/goapi/stores/web_resource/usecase"
)

func Test_maycParser(t *testing.T) {
	ctx := bCtx.Background()
	httpRepo := webresource_repository.NewHttpReaderRepo(http.Client{}, 2*time.Second, nil)

	webResourceUseCase := webresource_usecase.NewWebResourceUseCase(&webresource_usecase.WebResourceUseCaseCfg{
		HttpReader: httpRepo,
	})
	apecoinStakingUC := mocks.NewUseCase(t)
	p := NewMAYCParser(apecoinStakingUC)
	tests := []struct {
		name    string
		tokenId string
		want    map[string]string
	}{
		{
			name:    "m1",
			tokenId: "24936",
			want: map[string]string{
				"Background":    "\"M1 Yellow\"",
				"Fur":           "\"M1 Dark Brown\"",
				"Eyes":          "\"M1 Hypnotized\"",
				"Hat":           "\"M1 Commie Hat\"",
				"Mouth":         "\"M1 Bored Unshaven\"",
				"Earring":       "\"M1 Silver Stud\"",
				"Mutation Type": "M1",
			},
		}, {
			name:    "m2",
			tokenId: "15485",
			want: map[string]string{
				"Background":    "\"M2 Blue\"",
				"Fur":           "\"M2 Trippy\"",
				"Eyes":          "\"M2 Scumbag\"",
				"Clothes":       "\"M2 Leather Punk Jacket\"",
				"Hat":           "\"M2 Short Mohawk\"",
				"Mouth":         "\"M2 Bored Cigarette\"",
				"Mutation Type": "M2",
			},
		}, {
			name:    "mutant",
			tokenId: "30004",
			want: map[string]string{
				"Name":          "\"Mega DMT\"",
				"Mutation Type": "Mega",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := require.New(t)
			url := fmt.Sprintf("https://boredapeyachtclub.com/api/mutants/%s", tt.tokenId)
			data, err := webResourceUseCase.GetJson(ctx, url)
			req.NoError(err)
			apecoinStakingUC.EXPECT().Get(mock.AnythingOfType("ctx.Ctx"), mock.AnythingOfType("apecoinstaking.Id")).Return(nil, domain.ErrNotFound)
			attrs, err := p.Parse(ctx, 1, "0x5954ab967bc958940b7eb73ee84797dc8a2afbb9", domain.TokenId(tt.tokenId), data)
			req.NoError(err)
			for _, attr := range attrs {
				fmt.Printf("%s: %s\n", attr.TraitType, attr.Value)
				// req.Equal(tt.want[attr.TraitType], attr.Value)
			}
		})
	}
}
