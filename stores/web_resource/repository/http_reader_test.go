package repository

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	bCtx "github.com/x-xyz/goapi/base/ctx"
)

func Test_httpReaderRepo_Get(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	req := require.New(t)
	c := http.Client{}
	timeout := 10 * time.Second

	// SuperRare 0xb932a70A57673d89f4acfFBE830E8ed7f75Fb9e0 #26653
	url := "https://ipfs.pixura.io/ipfs/QmQqzKQQmwt5sxygmKdNDUj9XD5FmgELLaQ72h2tFdgeBV/metadata.json"
	expectedStr := `{"name":"REBIRTH","createdBy":"CACHOU","yearCreated":"2021","description":"No. 6 \"Rebirth\"\nA female high school student secretly being created for the provisional government.\nHowever something went wrong during manufacture.\nEverything went out of control...","image":"https://ipfs.pixura.io/ipfs/QmaByv7H1UCwpDpgSeMqga3hMGmuGzsrgyq9FU3S9JkkF5/srt.gif","media":{"uri":"https://ipfs.pixura.io/ipfs/QmezN1AvA7vzk4VCn6NnTLDQjnrvAUhPVw42riT7CftYPS/CACHOURebirth.mp4","dimensions":"2188x2188","size":"50353036","mimeType":"video/mp4"},"tags":["animation","art","digital","nft","superrare"]}`
	expected := []byte(expectedStr)
	ctx := bCtx.Background()

	r := NewHttpReaderRepo(c, timeout, nil)
	b, err := r.Get(ctx, url)
	req.NoError(err)
	req.Equal(expected, b)
}
