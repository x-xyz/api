package nft_indexer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractMimeTypeFromDataUri(t *testing.T) {
	cases := []struct {
		data string
		want string
		err  error
	}{
		{
			data: "data:image/png;base64,abcdefghijklm..",
			want: "image/png",
			err:  nil,
		},
		{
			data: "data:image/svg+xml;base64,abcdefghijklm..",
			want: "image/svg+xml",
			err:  nil,
		},
		{
			// base64 is not required
			data: "data:image/svg+xml,abcdefghijklm..",
			want: "image/svg+xml",
			err:  nil,
		},
		{
			data: "data:image/svg+xml#abcdefghijklm..",
			want: "",
			err:  errBadDataUriFormat,
		},
	}

	for _, c := range cases {
		output, err := extractMimeTypeFromDataUri(c.data)
		assert.Equal(t, c.want, output)
		assert.Equal(t, c.err, err)
	}
}
