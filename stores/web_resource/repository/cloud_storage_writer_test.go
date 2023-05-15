package repository

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"cloud.google.com/go/storage"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	bCtx "github.com/x-xyz/goapi/base/ctx"
	"golang.org/x/xerrors"
	"google.golang.org/api/iterator"
)

type cloudStorageTestSuite struct {
	suite.Suite
	client          *storage.Client
	bucketName      string
	bucketUrl       string
	testingMetadata []byte
	testingFolder   string
}

func (suite *cloudStorageTestSuite) SetupSuite() {
	ctx := bCtx.Background()
	metadataStr := `{"name":"REBIRTH","createdBy":"CACHOU","yearCreated":"2021","description":"No. 6 \"Rebirth\"\nA female high school student secretly being created for the provisional government.\nHowever something went wrong during manufacture.\nEverything went out of control...","image":"https://ipfs.pixura.io/ipfs/QmaByv7H1UCwpDpgSeMqga3hMGmuGzsrgyq9FU3S9JkkF5/srt.gif","media":{"uri":"https://ipfs.pixura.io/ipfs/QmezN1AvA7vzk4VCn6NnTLDQjnrvAUhPVw42riT7CftYPS/CACHOURebirth.mp4","dimensions":"2188x2188","size":"50353036","mimeType":"video/mp4"},"tags":["animation","art","digital","nft","superrare"]}`
	client, err := storage.NewClient(ctx)
	suite.NoError(err)

	suite.client = client
	suite.bucketName = "dev-storage.x.xyz"
	suite.bucketUrl = "https://dev-storage.x.xyz"
	suite.testingMetadata = []byte(metadataStr)
	suite.testingFolder = "testing"
}

func (suite *cloudStorageTestSuite) TearDownSuite() {
	ctx := bCtx.Background()
	query := &storage.Query{Prefix: suite.testingFolder}
	bucket := suite.client.Bucket(suite.bucketName)
	it := bucket.Objects(ctx, query)
	for {
		attr, err := it.Next()
		if err == iterator.Done {
			break
		}
		suite.NoError(err)
		err = bucket.Object(attr.Name).Delete(ctx)
		suite.NoError(err)
	}
	err := suite.client.Close()
	suite.NoError(err)
}

func TestCloudStorageWriterRepo(t *testing.T) {
	t.Skip("requires google cloud storage auth")
	suite.Run(t, new(cloudStorageTestSuite))
}

func (suite *cloudStorageTestSuite) Test_cloudStorageWriterRepo_Store() {
	req := require.New(suite.T())
	ctx := bCtx.Background()

	contentPath := fmt.Sprintf("%s/with/some/metadata.json", suite.testingFolder)
	expectedUrl := fmt.Sprintf("%s/%s", suite.bucketUrl, contentPath)
	cs, err := NewCloudStorageWriterRepo(&CloudStorageWriterRepoCfg{
		Client:     suite.client,
		BucketName: suite.bucketName,
		Timeout:    10 * time.Second,
		Url:        suite.bucketUrl,
	})
	req.NoError(err)
	url, err := cs.Store(ctx, contentPath, suite.testingMetadata, "")
	req.NoError(err)
	req.Equal(expectedUrl, url)

	body, err := httpGet(ctx, url)
	req.NoError(err)
	req.Equal(suite.testingMetadata, body)
}

func httpGet(ctx bCtx.Ctx, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, xerrors.Errorf("resp.StatusCode != 200")
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}
