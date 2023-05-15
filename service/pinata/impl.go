package pinata

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/x-xyz/goapi/base/ctx"
)

const (
	endpoint    = "https://api.pinata.cloud"
	pinPath     = "/pinning/pinFileToIPFS"
	pinJsonPath = "/pinning/pinJSONToIPFS"
)

type pinataImpl struct {
	apiKey    string
	apiSecret string
}

func New(apiKey, apiSecret string) Service {
	return &pinataImpl{apiKey, apiSecret}
}

func (im *pinataImpl) Pin(c ctx.Ctx, file io.Reader, extension string, optFns ...Options) (string, error) {
	opts, err := GetPinOptions(optFns...)

	if err != nil {
		c.WithField("err", err).Error("GetPinOptions failed")
		return "", err
	}

	var b bytes.Buffer

	w := multipart.NewWriter(&b)
	if fw, err := w.CreateFormFile("file", "file."+extension); err != nil {
		c.WithField("err", err).Error("w.CreateFormField failed")
		return "", err
	} else if _, err := io.Copy(fw, file); err != nil {
		c.WithField("err", err).Error("io.Copy failed")
		return "", err
	}

	if opts.Metadata != nil {
		if b, err := json.Marshal(opts.Metadata); err != nil {
			c.WithField("err", err).Error("json.Marshal failed")
			return "", err
		} else {
			w.WriteField("pinataMetadata", string(b))
		}
	}

	if opts.Options != nil {
		if b, err := json.Marshal(opts.Options); err != nil {
			c.WithField("err", err).Error("json.Marshal failed")
			return "", err
		} else {
			w.WriteField("pinataOptions", string(b))
		}
	}

	w.Close()

	url := fmt.Sprintf("%s%s", endpoint, pinPath)

	req, err := http.NewRequest("POST", url, &b)
	if err != nil {
		c.WithField("err", err).Error("http.NewRequest failed")
		return "", err
	}

	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("pinata_api_key", im.apiKey)
	req.Header.Set("pinata_secret_api_key", im.apiSecret)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		c.WithField("err", err).Error("DefaultClient.Do failed")
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errorBody, _ := io.ReadAll(resp.Body)
		c.WithField("errorBody", string(errorBody)).Error("Request failed")
		return "", ErrRequestFailed
	}

	type payload struct {
		IpfsHash string `json:"IpfsHash"`
	}

	p := &payload{}

	if err := json.NewDecoder(resp.Body).Decode(p); err != nil {
		c.WithField("err", err).Error("json.NewDecoder.Decode failed")
		return "", err
	}

	return p.IpfsHash, nil
}

func (im *pinataImpl) PinJson(c ctx.Ctx, value interface{}, optFns ...Options) (string, error) {
	opts, err := GetPinOptions(optFns...)

	if err != nil {
		c.WithField("err", err).Error("GetPinOptions failed")
		return "", err
	}

	opts.PinataContent = value

	body, err := json.Marshal(opts)
	if err != nil {
		c.WithField("err", err).Error("json.Marshal failed")
		return "", err
	}

	url := fmt.Sprintf("%s%s", endpoint, pinJsonPath)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		c.WithField("err", err).Error("http.NewRequest failed")
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("pinata_api_key", im.apiKey)
	req.Header.Set("pinata_secret_api_key", im.apiSecret)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		c.WithField("err", err).Error("DefaultClient.Do failed")
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errorBody, _ := io.ReadAll(resp.Body)
		c.WithField("errorBody", string(errorBody)).Error("Request failed")
		return "", ErrRequestFailed
	}

	type payload struct {
		IpfsHash string `json:"IpfsHash"`
	}

	p := &payload{}

	if err := json.NewDecoder(resp.Body).Decode(p); err != nil {
		c.WithField("err", err).Error("json.NewDecoder.Decode failed")
		return "", err
	}

	return p.IpfsHash, nil
}
