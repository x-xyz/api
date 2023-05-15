package usecase

import "testing"

func Test_getIpfsUrl(t *testing.T) {
	type args struct {
		url string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "pinata",
			args: args{
				url: "https://gateway.pinata.cloud/ipfs/QmVVutd4A4i1jCQnJXR49miQdXLNLVeGwyo5wWznpgRGeH",
			},
			want: "ipfs://QmVVutd4A4i1jCQnJXR49miQdXLNLVeGwyo5wWznpgRGeH",
		},
		{
			name: "pinata dedicated",
			args: args{
				url: "https://womenandweapons.mypinata.cloud/ipfs/QmTeTTMFgPYULCNkfxLcJSu5KByxDWh6JA4HFZY4CQnxdS",
			},
			want: "ipfs://QmTeTTMFgPYULCNkfxLcJSu5KByxDWh6JA4HFZY4CQnxdS",
		},
		{
			name: "ipfs.io",
			args: args{
				url: "https://ipfs.io/ipfs/QmRM6jM1Agru6fgm9aae1oFukwSi5d3Kk71Lue2rYznEYm/0.png",
			},
			want: "ipfs://QmRM6jM1Agru6fgm9aae1oFukwSi5d3Kk71Lue2rYznEYm/0.png",
		},
		{
			name: "cloudflare",
			args: args{
				url: "https://cloudflare-ipfs.com/ipfs/QmSddkqicov3HC1Urzv5AKPy2S7KqcnMQR5fjBnrFs2Z7A",
			},
			want: "ipfs://QmSddkqicov3HC1Urzv5AKPy2S7KqcnMQR5fjBnrFs2Z7A",
		},
		{
			name: "noop",
			args: args{
				url: "https://some.url",
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getIpfsUrl(tt.args.url); got != tt.want {
				t.Errorf("getIpfsUrl() = %v, want %v", got, tt.want)
			}
		})
	}
}
