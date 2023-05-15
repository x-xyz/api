package usecase

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/mock"
	bCtx "github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/mocks"
)

func Test_metadataUseCase_GetFromUrl(t *testing.T) {
	tests := []struct {
		name         string
		calledReader string
		url          string
		calledUrl    string
		want         *domain.Metadata
		wantErr      bool
	}{
		{
			name:    "invalid schema",
			url:     "http://url",
			wantErr: true,
		},
		{
			name:         "datauri ftm:0xf41270836df4db1d28f7fd0935270e3a603e78cc:26393 raritytools",
			calledReader: "datauri",
			url:          "data:application/json;base64,eyJuYW1lIjogIml0ZW0gIzI2MzkzIiwgImRlc2NyaXB0aW9uIjogIlJhcml0eSB0aWVyIDEsIG5vbiBtYWdpY2FsLCBpdGVtIGNyYWZ0aW5nLiIsICJpbWFnZSI6ICJkYXRhOmltYWdlL3N2Zyt4bWw7YmFzZTY0LFBITjJaeUI0Yld4dWN6MGlhSFIwY0RvdkwzZDNkeTUzTXk1dmNtY3ZNakF3TUM5emRtY2lJSEJ5WlhObGNuWmxRWE53WldOMFVtRjBhVzg5SW5oTmFXNVpUV2x1SUcxbFpYUWlJSFpwWlhkQ2IzZzlJakFnTUNBek5UQWdNelV3SWo0OGMzUjViR1UrTG1KaGMyVWdleUJtYVd4c09pQjNhR2wwWlRzZ1ptOXVkQzFtWVcxcGJIazZJSE5sY21sbU95Qm1iMjUwTFhOcGVtVTZJREUwY0hnN0lIMDhMM04wZVd4bFBqeHlaV04wSUhkcFpIUm9QU0l4TURBbElpQm9aV2xuYUhROUlqRXdNQ1VpSUdacGJHdzlJbUpzWVdOcklpQXZQangwWlhoMElIZzlJakV3SWlCNVBTSXlNQ0lnWTJ4aGMzTTlJbUpoYzJVaVBtTmhkR1ZuYjNKNUlFZHZiMlJ6UEM5MFpYaDBQangwWlhoMElIZzlJakV3SWlCNVBTSTBNQ0lnWTJ4aGMzTTlJbUpoYzJVaVBtNWhiV1VnVEdGdWRHVnliaXdnU0c5dlpHVmtQQzkwWlhoMFBqeDBaWGgwSUhnOUlqRXdJaUI1UFNJMk1DSWdZMnhoYzNNOUltSmhjMlVpUG1OdmMzUWdOMmR3UEM5MFpYaDBQangwWlhoMElIZzlJakV3SWlCNVBTSTRNQ0lnWTJ4aGMzTTlJbUpoYzJVaVBuZGxhV2RvZENBeWJHSThMM1JsZUhRK1BIUmxlSFFnZUQwaU1UQWlJSGs5SWpFd01DSWdZMnhoYzNNOUltSmhjMlVpUG1SbGMyTnlhWEIwYVc5dUlFRWdhRzl2WkdWa0lHeGhiblJsY200Z1kyeGxZWEpzZVNCcGJHeDFiV2x1WVhSbGN5QmhJRE13TFdadmIzUWdjbUZrYVhWeklHRnVaQ0J3Y205MmFXUmxjeUJ6YUdGa2IzZDVJR2xzYkhWdGFXNWhkR2x2YmlCcGJpQmhJRFl3TFdadmIzUWdjbUZrYVhWekxpQkpkQ0JpZFhKdWN5Qm1iM0lnTmlCb2IzVnljeUJ2YmlCaElIQnBiblFnYjJZZ2IybHNMaUJaYjNVZ1kyRnVJR05oY25KNUlHRWdhRzl2WkdWa0lHeGhiblJsY200Z2FXNGdiMjVsSUdoaGJtUXVQQzkwWlhoMFBqeDBaWGgwSUhnOUlqRXdJaUI1UFNJeE1qQWlJR05zWVhOelBTSmlZWE5sSWo1amNtRm1kR1ZrSUdKNUlESXhOVGcyT1RJOEwzUmxlSFErUEhSbGVIUWdlRDBpTVRBaUlIazlJakUwTUNJZ1kyeGhjM005SW1KaGMyVWlQbU55WVdaMFpXUWdZWFFnTVRZek5Ua3hORFExTWp3dmRHVjRkRDQ4TDNOMlp6ND0ifQ==",
			calledUrl:    "data:application/json;base64,eyJuYW1lIjogIml0ZW0gIzI2MzkzIiwgImRlc2NyaXB0aW9uIjogIlJhcml0eSB0aWVyIDEsIG5vbiBtYWdpY2FsLCBpdGVtIGNyYWZ0aW5nLiIsICJpbWFnZSI6ICJkYXRhOmltYWdlL3N2Zyt4bWw7YmFzZTY0LFBITjJaeUI0Yld4dWN6MGlhSFIwY0RvdkwzZDNkeTUzTXk1dmNtY3ZNakF3TUM5emRtY2lJSEJ5WlhObGNuWmxRWE53WldOMFVtRjBhVzg5SW5oTmFXNVpUV2x1SUcxbFpYUWlJSFpwWlhkQ2IzZzlJakFnTUNBek5UQWdNelV3SWo0OGMzUjViR1UrTG1KaGMyVWdleUJtYVd4c09pQjNhR2wwWlRzZ1ptOXVkQzFtWVcxcGJIazZJSE5sY21sbU95Qm1iMjUwTFhOcGVtVTZJREUwY0hnN0lIMDhMM04wZVd4bFBqeHlaV04wSUhkcFpIUm9QU0l4TURBbElpQm9aV2xuYUhROUlqRXdNQ1VpSUdacGJHdzlJbUpzWVdOcklpQXZQangwWlhoMElIZzlJakV3SWlCNVBTSXlNQ0lnWTJ4aGMzTTlJbUpoYzJVaVBtTmhkR1ZuYjNKNUlFZHZiMlJ6UEM5MFpYaDBQangwWlhoMElIZzlJakV3SWlCNVBTSTBNQ0lnWTJ4aGMzTTlJbUpoYzJVaVBtNWhiV1VnVEdGdWRHVnliaXdnU0c5dlpHVmtQQzkwWlhoMFBqeDBaWGgwSUhnOUlqRXdJaUI1UFNJMk1DSWdZMnhoYzNNOUltSmhjMlVpUG1OdmMzUWdOMmR3UEM5MFpYaDBQangwWlhoMElIZzlJakV3SWlCNVBTSTRNQ0lnWTJ4aGMzTTlJbUpoYzJVaVBuZGxhV2RvZENBeWJHSThMM1JsZUhRK1BIUmxlSFFnZUQwaU1UQWlJSGs5SWpFd01DSWdZMnhoYzNNOUltSmhjMlVpUG1SbGMyTnlhWEIwYVc5dUlFRWdhRzl2WkdWa0lHeGhiblJsY200Z1kyeGxZWEpzZVNCcGJHeDFiV2x1WVhSbGN5QmhJRE13TFdadmIzUWdjbUZrYVhWeklHRnVaQ0J3Y205MmFXUmxjeUJ6YUdGa2IzZDVJR2xzYkhWdGFXNWhkR2x2YmlCcGJpQmhJRFl3TFdadmIzUWdjbUZrYVhWekxpQkpkQ0JpZFhKdWN5Qm1iM0lnTmlCb2IzVnljeUJ2YmlCaElIQnBiblFnYjJZZ2IybHNMaUJaYjNVZ1kyRnVJR05oY25KNUlHRWdhRzl2WkdWa0lHeGhiblJsY200Z2FXNGdiMjVsSUdoaGJtUXVQQzkwWlhoMFBqeDBaWGgwSUhnOUlqRXdJaUI1UFNJeE1qQWlJR05zWVhOelBTSmlZWE5sSWo1amNtRm1kR1ZrSUdKNUlESXhOVGcyT1RJOEwzUmxlSFErUEhSbGVIUWdlRDBpTVRBaUlIazlJakUwTUNJZ1kyeGhjM005SW1KaGMyVWlQbU55WVdaMFpXUWdZWFFnTVRZek5Ua3hORFExTWp3dmRHVjRkRDQ4TDNOMlp6ND0ifQ==",
			want:         &domain.Metadata{RawMessage: []byte(`{"name": "item #26393", "description": "Rarity tier 1, non magical, item crafting.", "image": "data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHByZXNlcnZlQXNwZWN0UmF0aW89InhNaW5ZTWluIG1lZXQiIHZpZXdCb3g9IjAgMCAzNTAgMzUwIj48c3R5bGU+LmJhc2UgeyBmaWxsOiB3aGl0ZTsgZm9udC1mYW1pbHk6IHNlcmlmOyBmb250LXNpemU6IDE0cHg7IH08L3N0eWxlPjxyZWN0IHdpZHRoPSIxMDAlIiBoZWlnaHQ9IjEwMCUiIGZpbGw9ImJsYWNrIiAvPjx0ZXh0IHg9IjEwIiB5PSIyMCIgY2xhc3M9ImJhc2UiPmNhdGVnb3J5IEdvb2RzPC90ZXh0Pjx0ZXh0IHg9IjEwIiB5PSI0MCIgY2xhc3M9ImJhc2UiPm5hbWUgTGFudGVybiwgSG9vZGVkPC90ZXh0Pjx0ZXh0IHg9IjEwIiB5PSI2MCIgY2xhc3M9ImJhc2UiPmNvc3QgN2dwPC90ZXh0Pjx0ZXh0IHg9IjEwIiB5PSI4MCIgY2xhc3M9ImJhc2UiPndlaWdodCAybGI8L3RleHQ+PHRleHQgeD0iMTAiIHk9IjEwMCIgY2xhc3M9ImJhc2UiPmRlc2NyaXB0aW9uIEEgaG9vZGVkIGxhbnRlcm4gY2xlYXJseSBpbGx1bWluYXRlcyBhIDMwLWZvb3QgcmFkaXVzIGFuZCBwcm92aWRlcyBzaGFkb3d5IGlsbHVtaW5hdGlvbiBpbiBhIDYwLWZvb3QgcmFkaXVzLiBJdCBidXJucyBmb3IgNiBob3VycyBvbiBhIHBpbnQgb2Ygb2lsLiBZb3UgY2FuIGNhcnJ5IGEgaG9vZGVkIGxhbnRlcm4gaW4gb25lIGhhbmQuPC90ZXh0Pjx0ZXh0IHg9IjEwIiB5PSIxMjAiIGNsYXNzPSJiYXNlIj5jcmFmdGVkIGJ5IDIxNTg2OTI8L3RleHQ+PHRleHQgeD0iMTAiIHk9IjE0MCIgY2xhc3M9ImJhc2UiPmNyYWZ0ZWQgYXQgMTYzNTkxNDQ1MjwvdGV4dD48L3N2Zz4="}`)},
			wantErr:      false,
		},
		{
			name:         "http eth:0xb932a70A57673d89f4acfFBE830E8ed7f75Fb9e0:26653 SuperRare",
			calledReader: "http",
			url:          "https://ipfs.pixura.io/ipfs/QmQqzKQQmwt5sxygmKdNDUj9XD5FmgELLaQ72h2tFdgeBV/metadata.json",
			calledUrl:    "https://ipfs.pixura.io/ipfs/QmQqzKQQmwt5sxygmKdNDUj9XD5FmgELLaQ72h2tFdgeBV/metadata.json",
			want:         &domain.Metadata{RawMessage: []byte(`{"name":"REBIRTH","createdBy":"CACHOU","yearCreated":"2021","description":"No. 6 \"Rebirth\"\nA female high school student secretly being created for the provisional government.\nHowever something went wrong during manufacture.\nEverything went out of control...","image":"https://ipfs.pixura.io/ipfs/QmaByv7H1UCwpDpgSeMqga3hMGmuGzsrgyq9FU3S9JkkF5/srt.gif","media":{"uri":"https://ipfs.pixura.io/ipfs/QmezN1AvA7vzk4VCn6NnTLDQjnrvAUhPVw42riT7CftYPS/CACHOURebirth.mp4","dimensions":"2188x2188","size":"50353036","mimeType":"video/mp4"},"tags":["animation","art","digital","nft","superrare"]}`)},
			wantErr:      false,
		},
		{
			name:         "ipfs eth:0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d:0 BAYC",
			calledReader: "ipfs",
			url:          "ipfs://QmeSjSinHpPnmXmspMjwiXyN6zS4E9zccariGR3jxcaWtq/0",
			calledUrl:    "QmeSjSinHpPnmXmspMjwiXyN6zS4E9zccariGR3jxcaWtq/0",
			want:         &domain.Metadata{RawMessage: []byte(fmt.Sprintf("%s\n", `{"image":"ipfs://QmRRPWG96cmgTn2qSzjwr2qvfNEuhunv6FNeMFGa9bx6mQ","attributes":[{"trait_type":"Earring","value":"Silver Hoop"},{"trait_type":"Background","value":"Orange"},{"trait_type":"Fur","value":"Robot"},{"trait_type":"Clothes","value":"Striped Tee"},{"trait_type":"Mouth","value":"Discomfort"},{"trait_type":"Eyes","value":"X Eyes"}]}`))},
			wantErr:      false,
		},
	}
	// url := "localhost:5001"
	// s := ipfsapi.NewShell(url)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			readers := map[string]domain.WebResourceReaderRepository{
				"http":    &mocks.WebResourceReaderRepository{},
				"ipfs":    &mocks.WebResourceReaderRepository{},
				"datauri": &mocks.WebResourceReaderRepository{},
			}
			if len(tt.calledReader) > 0 {
				b := []byte(tt.want.RawMessage)
				readers[tt.calledReader].(*mocks.WebResourceReaderRepository).
					On("Get", mock.Anything, tt.calledUrl).
					Return(b, nil)
			}
			u := NewMetadataUseCase(&MetadataUseCaseCfg{
				HttpReader:    readers["http"],
				IpfsReader:    readers["ipfs"],
				DataUriReader: readers["datauri"],
				// HttpReader:    repository.NewHttpReaderRepo(http.Client{}, 10*time.Second),
				// IpfsReader:    repository.NewIpfsNodeApiReaderRepo(s, 10*time.Second),
				// DataUriReader: repository.NewDataUriReaderRepo(),
			})
			ctx := bCtx.Background()
			got, err := u.GetFromUrl(ctx, tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("metadataUseCase.GetFromUrl() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("metadataUseCase.GetFromUrl() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_metadataUseCase_Store(t *testing.T) {
	chainId := domain.ChainId(1)
	contractAddress := "addr"
	tokenId := int32(5566)
	path := "1/addr/5566"
	metadata := &domain.Metadata{RawMessage: []byte(`{"name":"REBIRTH"}`)}
	tests := []struct {
		name    string
		want    string
		wantErr bool
	}{
		{
			name:    "error",
			want:    "",
			wantErr: true,
		},
		{
			name:    "stored",
			want:    "https://someurl",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		expectedErr := error(nil)
		if tt.wantErr {
			expectedErr = errors.New("err")
		}
		writer := &mocks.WebResourceWriterRepository{}
		writer.On("Store", mock.Anything, path, []byte(metadata.RawMessage)).Return(tt.want, expectedErr)
		u := NewMetadataUseCase(&MetadataUseCaseCfg{
			CloudStorageWriter: writer,
		})

		ctx := bCtx.Background()
		got, err := u.Store(ctx, chainId, contractAddress, tokenId, metadata)
		if (err != nil) != tt.wantErr {
			t.Errorf("metadataUseCase.Store() error = %v, wantErr %v", err, tt.wantErr)
			return
		}
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("metadataUseCase.Store() = %v, want %v", got, tt.want)
		}

	}
}
