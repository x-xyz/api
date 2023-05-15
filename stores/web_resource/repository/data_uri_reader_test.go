package repository

import (
	"reflect"
	"testing"

	bCtx "github.com/x-xyz/goapi/base/ctx"
)

func Test_dataUriReaderRepo_Get(t *testing.T) {
	tests := []struct {
		name    string
		uri     string
		want    []byte
		wantErr bool
	}{
		{
			name:    "invalid schema",
			uri:     "https://url",
			wantErr: true,
		},
		{
			name:    "no data part",
			uri:     "data:application/json;base64,",
			wantErr: true,
		},
		{
			name:    "no data part",
			uri:     "data:application/json;base64",
			wantErr: true,
		},
		{
			name: "eth:0x892848074ddea461a15f337250da3ce55580ca85:1 cyberbrokers",
			uri:  `data:application/json;utf8,{"name":"Enticing Delightful","description":"Look around you. Everything in TPL is made from code. Hackers hunt from far corner to far corner, querying rough edges for exploits. Sometimes, they'll take commission, but, more than likely, if they're talking to you, be polite.  There's a very expensive reason why you should.","external_url":"https://cyberbrokers.io/","image":"ipfs://QmcsrQJMKA9qC9GcEMgdjb9LPN99iDNAg8aQQJLJGpkHxk/1.svg","attributes":[{"trait_type": "Talent", "value": "Hacker"},{"trait_type": "Species", "value": "Human"},{"trait_type": "Class", "value": "Explorers"},{"trait_type": "Mind", "value": 14},{"trait_type": "Body", "value": 16},{"trait_type": "Soul", "value": 14},{"trait_type": "Underwear", "value": "Dress Blue"},{"trait_type": "Pants", "value": "Check Engine Light"},{"trait_type": "Outerwear", "value": "Cyber"}]}`,
			want: []byte(`{"name":"Enticing Delightful","description":"Look around you. Everything in TPL is made from code. Hackers hunt from far corner to far corner, querying rough edges for exploits. Sometimes, they'll take commission, but, more than likely, if they're talking to you, be polite.  There's a very expensive reason why you should.","external_url":"https://cyberbrokers.io/","image":"ipfs://QmcsrQJMKA9qC9GcEMgdjb9LPN99iDNAg8aQQJLJGpkHxk/1.svg","attributes":[{"trait_type": "Talent", "value": "Hacker"},{"trait_type": "Species", "value": "Human"},{"trait_type": "Class", "value": "Explorers"},{"trait_type": "Mind", "value": 14},{"trait_type": "Body", "value": 16},{"trait_type": "Soul", "value": 14},{"trait_type": "Underwear", "value": "Dress Blue"},{"trait_type": "Pants", "value": "Check Engine Light"},{"trait_type": "Outerwear", "value": "Cyber"}]}`),
		},
		{
			name:    "ftm:0xf41270836df4db1d28f7fd0935270e3a603e78cc:26393 raritytools",
			uri:     "data:application/json;base64,eyJuYW1lIjogIml0ZW0gIzI2MzkzIiwgImRlc2NyaXB0aW9uIjogIlJhcml0eSB0aWVyIDEsIG5vbiBtYWdpY2FsLCBpdGVtIGNyYWZ0aW5nLiIsICJpbWFnZSI6ICJkYXRhOmltYWdlL3N2Zyt4bWw7YmFzZTY0LFBITjJaeUI0Yld4dWN6MGlhSFIwY0RvdkwzZDNkeTUzTXk1dmNtY3ZNakF3TUM5emRtY2lJSEJ5WlhObGNuWmxRWE53WldOMFVtRjBhVzg5SW5oTmFXNVpUV2x1SUcxbFpYUWlJSFpwWlhkQ2IzZzlJakFnTUNBek5UQWdNelV3SWo0OGMzUjViR1UrTG1KaGMyVWdleUJtYVd4c09pQjNhR2wwWlRzZ1ptOXVkQzFtWVcxcGJIazZJSE5sY21sbU95Qm1iMjUwTFhOcGVtVTZJREUwY0hnN0lIMDhMM04wZVd4bFBqeHlaV04wSUhkcFpIUm9QU0l4TURBbElpQm9aV2xuYUhROUlqRXdNQ1VpSUdacGJHdzlJbUpzWVdOcklpQXZQangwWlhoMElIZzlJakV3SWlCNVBTSXlNQ0lnWTJ4aGMzTTlJbUpoYzJVaVBtTmhkR1ZuYjNKNUlFZHZiMlJ6UEM5MFpYaDBQangwWlhoMElIZzlJakV3SWlCNVBTSTBNQ0lnWTJ4aGMzTTlJbUpoYzJVaVBtNWhiV1VnVEdGdWRHVnliaXdnU0c5dlpHVmtQQzkwWlhoMFBqeDBaWGgwSUhnOUlqRXdJaUI1UFNJMk1DSWdZMnhoYzNNOUltSmhjMlVpUG1OdmMzUWdOMmR3UEM5MFpYaDBQangwWlhoMElIZzlJakV3SWlCNVBTSTRNQ0lnWTJ4aGMzTTlJbUpoYzJVaVBuZGxhV2RvZENBeWJHSThMM1JsZUhRK1BIUmxlSFFnZUQwaU1UQWlJSGs5SWpFd01DSWdZMnhoYzNNOUltSmhjMlVpUG1SbGMyTnlhWEIwYVc5dUlFRWdhRzl2WkdWa0lHeGhiblJsY200Z1kyeGxZWEpzZVNCcGJHeDFiV2x1WVhSbGN5QmhJRE13TFdadmIzUWdjbUZrYVhWeklHRnVaQ0J3Y205MmFXUmxjeUJ6YUdGa2IzZDVJR2xzYkhWdGFXNWhkR2x2YmlCcGJpQmhJRFl3TFdadmIzUWdjbUZrYVhWekxpQkpkQ0JpZFhKdWN5Qm1iM0lnTmlCb2IzVnljeUJ2YmlCaElIQnBiblFnYjJZZ2IybHNMaUJaYjNVZ1kyRnVJR05oY25KNUlHRWdhRzl2WkdWa0lHeGhiblJsY200Z2FXNGdiMjVsSUdoaGJtUXVQQzkwWlhoMFBqeDBaWGgwSUhnOUlqRXdJaUI1UFNJeE1qQWlJR05zWVhOelBTSmlZWE5sSWo1amNtRm1kR1ZrSUdKNUlESXhOVGcyT1RJOEwzUmxlSFErUEhSbGVIUWdlRDBpTVRBaUlIazlJakUwTUNJZ1kyeGhjM005SW1KaGMyVWlQbU55WVdaMFpXUWdZWFFnTVRZek5Ua3hORFExTWp3dmRHVjRkRDQ4TDNOMlp6ND0ifQ==",
			want:    []byte(`{"name": "item #26393", "description": "Rarity tier 1, non magical, item crafting.", "image": "data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHByZXNlcnZlQXNwZWN0UmF0aW89InhNaW5ZTWluIG1lZXQiIHZpZXdCb3g9IjAgMCAzNTAgMzUwIj48c3R5bGU+LmJhc2UgeyBmaWxsOiB3aGl0ZTsgZm9udC1mYW1pbHk6IHNlcmlmOyBmb250LXNpemU6IDE0cHg7IH08L3N0eWxlPjxyZWN0IHdpZHRoPSIxMDAlIiBoZWlnaHQ9IjEwMCUiIGZpbGw9ImJsYWNrIiAvPjx0ZXh0IHg9IjEwIiB5PSIyMCIgY2xhc3M9ImJhc2UiPmNhdGVnb3J5IEdvb2RzPC90ZXh0Pjx0ZXh0IHg9IjEwIiB5PSI0MCIgY2xhc3M9ImJhc2UiPm5hbWUgTGFudGVybiwgSG9vZGVkPC90ZXh0Pjx0ZXh0IHg9IjEwIiB5PSI2MCIgY2xhc3M9ImJhc2UiPmNvc3QgN2dwPC90ZXh0Pjx0ZXh0IHg9IjEwIiB5PSI4MCIgY2xhc3M9ImJhc2UiPndlaWdodCAybGI8L3RleHQ+PHRleHQgeD0iMTAiIHk9IjEwMCIgY2xhc3M9ImJhc2UiPmRlc2NyaXB0aW9uIEEgaG9vZGVkIGxhbnRlcm4gY2xlYXJseSBpbGx1bWluYXRlcyBhIDMwLWZvb3QgcmFkaXVzIGFuZCBwcm92aWRlcyBzaGFkb3d5IGlsbHVtaW5hdGlvbiBpbiBhIDYwLWZvb3QgcmFkaXVzLiBJdCBidXJucyBmb3IgNiBob3VycyBvbiBhIHBpbnQgb2Ygb2lsLiBZb3UgY2FuIGNhcnJ5IGEgaG9vZGVkIGxhbnRlcm4gaW4gb25lIGhhbmQuPC90ZXh0Pjx0ZXh0IHg9IjEwIiB5PSIxMjAiIGNsYXNzPSJiYXNlIj5jcmFmdGVkIGJ5IDIxNTg2OTI8L3RleHQ+PHRleHQgeD0iMTAiIHk9IjE0MCIgY2xhc3M9ImJhc2UiPmNyYWZ0ZWQgYXQgMTYzNTkxNDQ1MjwvdGV4dD48L3N2Zz4="}`),
			wantErr: false,
		},
		{
			name:    "ftm:0xf41270836df4db1d28f7fd0935270e3a603e78cc:26393 raritytools (svg image)",
			uri:     "data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHByZXNlcnZlQXNwZWN0UmF0aW89InhNaW5ZTWluIG1lZXQiIHZpZXdCb3g9IjAgMCAzNTAgMzUwIj48c3R5bGU+LmJhc2UgeyBmaWxsOiB3aGl0ZTsgZm9udC1mYW1pbHk6IHNlcmlmOyBmb250LXNpemU6IDE0cHg7IH08L3N0eWxlPjxyZWN0IHdpZHRoPSIxMDAlIiBoZWlnaHQ9IjEwMCUiIGZpbGw9ImJsYWNrIiAvPjx0ZXh0IHg9IjEwIiB5PSIyMCIgY2xhc3M9ImJhc2UiPmNhdGVnb3J5IEdvb2RzPC90ZXh0Pjx0ZXh0IHg9IjEwIiB5PSI0MCIgY2xhc3M9ImJhc2UiPm5hbWUgTGFudGVybiwgSG9vZGVkPC90ZXh0Pjx0ZXh0IHg9IjEwIiB5PSI2MCIgY2xhc3M9ImJhc2UiPmNvc3QgN2dwPC90ZXh0Pjx0ZXh0IHg9IjEwIiB5PSI4MCIgY2xhc3M9ImJhc2UiPndlaWdodCAybGI8L3RleHQ+PHRleHQgeD0iMTAiIHk9IjEwMCIgY2xhc3M9ImJhc2UiPmRlc2NyaXB0aW9uIEEgaG9vZGVkIGxhbnRlcm4gY2xlYXJseSBpbGx1bWluYXRlcyBhIDMwLWZvb3QgcmFkaXVzIGFuZCBwcm92aWRlcyBzaGFkb3d5IGlsbHVtaW5hdGlvbiBpbiBhIDYwLWZvb3QgcmFkaXVzLiBJdCBidXJucyBmb3IgNiBob3VycyBvbiBhIHBpbnQgb2Ygb2lsLiBZb3UgY2FuIGNhcnJ5IGEgaG9vZGVkIGxhbnRlcm4gaW4gb25lIGhhbmQuPC90ZXh0Pjx0ZXh0IHg9IjEwIiB5PSIxMjAiIGNsYXNzPSJiYXNlIj5jcmFmdGVkIGJ5IDIxNTg2OTI8L3RleHQ+PHRleHQgeD0iMTAiIHk9IjE0MCIgY2xhc3M9ImJhc2UiPmNyYWZ0ZWQgYXQgMTYzNTkxNDQ1MjwvdGV4dD48L3N2Zz4=",
			want:    []byte(`<svg xmlns="http://www.w3.org/2000/svg" preserveAspectRatio="xMinYMin meet" viewBox="0 0 350 350"><style>.base { fill: white; font-family: serif; font-size: 14px; }</style><rect width="100%" height="100%" fill="black" /><text x="10" y="20" class="base">category Goods</text><text x="10" y="40" class="base">name Lantern, Hooded</text><text x="10" y="60" class="base">cost 7gp</text><text x="10" y="80" class="base">weight 2lb</text><text x="10" y="100" class="base">description A hooded lantern clearly illuminates a 30-foot radius and provides shadowy illumination in a 60-foot radius. It burns for 6 hours on a pint of oil. You can carry a hooded lantern in one hand.</text><text x="10" y="120" class="base">crafted by 2158692</text><text x="10" y="140" class="base">crafted at 1635914452</text></svg>`),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewDataUriReaderRepo()
			ctx := bCtx.Background()
			got, err := r.Get(ctx, tt.uri)
			if (err != nil) != tt.wantErr {
				t.Errorf("dataUriReaderRepo.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("dataUriReaderRepo.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}
