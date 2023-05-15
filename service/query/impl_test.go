package query

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/bsonx"

	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/database/mongoclient"
	"github.com/x-xyz/goapi/domain"
)

var (
	mockCTX = ctx.Background()
)

const (
	mockTable = domain.TableAccounts
	dbName    = "testdb"
)

type querySuite struct {
	suite.Suite
	im       *impl
	mongoURI string
}

func (q *querySuite) SetupSuite() {
	q.mongoURI = "mongodb://xxyz:xxyz@localhost:28000/?retryWrites=true&w=majority"

}

func (q *querySuite) TearDownSuite() {
}

func (q *querySuite) SetupTest() {
	q.im = &impl{
		client:     mongoclient.MustConnectMongoClient(q.mongoURI, "admin", dbName, false, true, 1),
		checkIndex: false,
	}
	q.Require().NoError(q.im.client.Database(q.im.client.DbName).Collection(string(mockTable)).Drop(ctx.Background()))
}

func (q *querySuite) testFindOne() {
	type Dummy struct {
		Dummy  string `json:"dummy" bson:"dummy"`
		Update string `json:"updatekey" bson:"updatekey"`
	}

	mockDummyValue := Dummy{"test-value11155", "test-value222255"}

	// First set-insert
	err := q.im.Upsert(mockCTX, mockTable, bson.M{"dummy": "test-value11155"}, bson.M{"dummy": "test-value11155", "updatekey": "test-value222255"})
	q.NoError(err)

	result := &Dummy{}
	err = q.im.FindOne(mockCTX, mockTable, bson.M{"dummy": "test-value11155"}, result)
	q.Require().NoError(err)
	q.Equal(mockDummyValue, *result)

	err = q.im.FindOne(mockCTX, mockTable, bson.M{"dummy": "test-value11166"}, result)
	q.Error(ErrNotFound)
}

func (q *querySuite) TestFindOne() {
	q.testFindOne()
}

func (q *querySuite) TestInsert() {
	type Dummy struct {
		Dummy  string `json:"dummy" bson:"dummy"`
		Update string `json:"updatekey" bson:"updatekey"`
	}

	mockDummyValue := Dummy{"test-value113", "test-value2222"}

	err := q.im.Insert(
		mockCTX, mockTable,
		bson.M{"dummy": "test-value113", "updatekey": "test-value2222"},
	)
	q.NoError(err)

	client := q.im.getClient(mockCTX)

	v := &Dummy{}
	r := client.Database(dbName).Collection(string(mockTable)).FindOne(mockCTX, bson.M{"dummy": "test-value113"})
	err = r.Decode(&v)
	q.Require().NoError(err)
	q.Equal(mockDummyValue, *v)

	err = q.im.Insert(
		mockCTX, mockTable,
		bson.M{"dummy": "test-value113", "updatekey": "test-value22223"},
	)
	q.NoError(err)

	c, err := client.Database(dbName).Collection(string(mockTable)).CountDocuments(mockCTX, bson.M{"dummy": "test-value113"})
	q.Require().NoError(err)
	q.Equal(2, int(c))
}

func (q *querySuite) TestInsertShouldFailWithDuplicateKey() {
	type Dummy struct {
		Dummy  string `json:"dummy" bson:"dummy"`
		Update string `json:"updatekey" bson:"updatekey"`
	}

	mockDummyValue := Dummy{"test-value113", "test-value2222"}

	err := q.im.Insert(
		mockCTX, mockTable,
		bson.M{"dummy": "test-value113", "updatekey": "test-value2222"},
	)
	q.NoError(err)

	client := q.im.getClient(mockCTX)

	col := client.Database(dbName).Collection(string(mockTable))

	v := &Dummy{}
	r := col.FindOne(mockCTX, bson.M{"dummy": "test-value113"})
	err = r.Decode(&v)
	q.Require().NoError(err)
	q.Equal(mockDummyValue, *v)

	keys := bsonx.Doc{{Key: "dummy", Value: bsonx.Int32(1)}}
	unique := true
	index := mongo.IndexModel{
		Keys: keys,
		Options: &options.IndexOptions{
			Unique: &unique,
		},
	}
	_, err = col.Indexes().CreateOne(mockCTX, index)
	q.Require().NoError(err)

	err = q.im.Insert(
		mockCTX, mockTable,
		bson.M{"dummy": "test-value113", "updatekey": "test-value2222"},
	)
	q.Require().Equal(ErrDuplicateKey, err)

	err = q.im.Insert(
		mockCTX, mockTable,
		bson.M{"dummy": "test-value114", "updatekey": "test-value2222"},
	)
	q.Require().NoError(err)
}

func (q *querySuite) TestUpsert() {
	type Dummy struct {
		Dummy  string `json:"dummy" bson:"dummy"`
		Update string `json:"updatekey" bson:"updatekey"`
		Dummy2 string `json:"dummy2" bson:"dummy2"`
	}

	mockDummyValue := Dummy{"test-value113", "test-value2222", "test-value"}

	client := q.im.getClient(mockCTX)

	// First set-insert
	err := q.im.Upsert(
		mockCTX, mockTable,
		bson.M{"dummy": "test-value113"},
		bson.M{"dummy": "test-value113", "updatekey": "test-value2222", "dummy2": "test-value"},
	)
	q.Require().NoError(err)

	v := &Dummy{}
	res := client.Database(dbName).Collection(string(mockTable)).FindOne(mockCTX, bson.M{"dummy": "test-value113"})
	err = res.Decode(v)
	q.Require().NoError(err)
	q.Equal(mockDummyValue, *v)

	// Test update (Second upsert)
	mockDummyValue2 := Dummy{"test-value113", "test-value2222", ""}
	err = q.im.Upsert(mockCTX, mockTable, bson.M{"dummy": "test-value113"}, mockDummyValue2)
	q.Require().NoError(err)

	v = &Dummy{}
	res = client.Database(dbName).Collection(string(mockTable)).FindOne(mockCTX, bson.M{"dummy": "test-value113"})
	err = res.Decode(v)
	q.Require().NoError(err)
	q.Equal(mockDummyValue2, *v)
}

func (q *querySuite) TestCount() {
	type Dummy struct {
		Dummy  string `json:"dummy" bson:"dummy"`
		Update string `json:"updatekey" bson:"updatekey"`
	}

	// Should be 0 at first
	cnt, err := q.im.Count(mockCTX, mockTable, bson.M{"dummy": "test-value-count0"})
	q.NoError(err)
	q.Equal(0, cnt)
	// Test EstimateCount
	cnt, err = q.im.EstimateCount(mockCTX, mockTable, bson.M{"dummy": "test-value-count0"})
	q.NoError(err)
	q.Equal(0, cnt)

	// Insert one doc
	d := Dummy{"test-value-count0", "test-value-count0"}
	err = q.im.Upsert(mockCTX, mockTable, bson.M{"updatekey": "test-value-count0"}, d)
	q.NoError(err)

	// count should be 1
	cnt, err = q.im.Count(mockCTX, mockTable, bson.M{"dummy": "test-value-count0"})
	q.NoError(err)
	q.Equal(1, cnt)
	cnt, err = q.im.EstimateCount(mockCTX, mockTable, bson.M{"dummy": "test-value-count0"})
	q.NoError(err)
	q.Equal(1, cnt)

	// insert another one
	d = Dummy{"test-value-count0", "test-value-count1"}
	err = q.im.Upsert(mockCTX, mockTable, bson.M{"updatekey": "test-value-count1"}, d)
	q.NoError(err)

	// now count should be 2
	cnt, err = q.im.Count(mockCTX, mockTable, bson.M{"dummy": "test-value-count0"})
	q.NoError(err)
	q.Equal(2, cnt)
	cnt, err = q.im.EstimateCount(mockCTX, mockTable, bson.M{"dummy": "test-value-count0"})
	q.NoError(err)
	q.Equal(2, cnt)
}

func (q *querySuite) testSearch() {
	type Dummy struct {
		Dummy  string `bson:"dummy" json:"dummy"`
		Update string `bson:"updatekey" json:"updatekey"`
	}

	mockDummyValue := []Dummy{{"test-value222232", "test-value222255"}}

	// First set-insert
	err := q.im.Upsert(
		mockCTX, mockTable, bson.M{"dummy": "test-value222232"},
		bson.M{"dummy": "test-value222232", "updatekey": "test-value222255"},
	)
	q.NoError(err)

	var result []Dummy
	err = q.im.Search(mockCTX, mockTable, 0, 5, "dummy", bson.M{"dummy": "test-value222232"}, &result)
	q.Require().NoError(err)
	q.Equal(mockDummyValue, result)

	err = q.im.Search(mockCTX, mockTable, 0, 5, "", bson.M{"dummy": "test-value222232"}, &result)
	q.Require().NoError(err)
	q.Equal(mockDummyValue, result)
}
func (q *querySuite) TestSearch() {
	q.testSearch()
}

func (q *querySuite) TestSearchWithIndex() {
	type Dummy struct {
		Dummy  string `bson:"dummy" json:"dummy"`
		Update string `bson:"updatekey" json:"updatekey"`
	}

	mockDummyValue := []Dummy{{"test-value222232", "test-value222255"}}

	client := q.im.getClient(mockCTX)

	indexView := client.Database(dbName).Collection(string(mockTable)).Indexes()
	_, idxErr := indexView.CreateOne(mockCTX, mongo.IndexModel{Keys: bson.M{"dummy": 1}})
	q.Require().NoError(idxErr)

	// First set-insert
	err := q.im.Upsert(
		mockCTX, mockTable, bson.M{"dummy": "test-value222232"},
		bson.M{"dummy": "test-value222232", "updatekey": "test-value222255"},
	)
	q.NoError(err)

	q.im.checkIndex = true

	var result []Dummy
	err = q.im.Search(mockCTX, mockTable, 0, 5, "dummy", bson.M{"dummy": "test-value222232"}, &result)
	q.NoError(err)
	q.Equal(mockDummyValue, result)
}

func (q *querySuite) TestSearchWithoutIndex() {
	type Dummy struct {
		Dummy  string `bson:"dummy" json:"dummy"`
		Update string `bson:"updatekey" json:"updatekey"`
	}

	// First set-insert
	err := q.im.Upsert(
		mockCTX, mockTable, bson.M{"dummy": "test-value222232"},
		bson.M{"dummy": "test-value222232", "updatekey": "test-value222255"},
	)
	q.NoError(err)

	q.im.checkIndex = true

	var result []Dummy
	err = q.im.Search(mockCTX, mockTable, 0, 5, "dummy", bson.M{"dummy": "test-value222232"}, &result)
	q.Equal(ErrCollScan, err)
}

func (q *querySuite) TestSearchNSorts() {
	type Dummy struct {
		Dummy  string `bson:"dummy" json:"dummy"`
		Update string `bson:"updatekey" json:"updatekey"`
	}

	mockDummyValue := []Dummy{{"test-value222232", "test-value222255"}}

	// First set-insert
	err := q.im.Upsert(mockCTX, mockTable, bson.M{"dummy": "test-value222232"}, bson.M{"dummy": "test-value222232", "updatekey": "test-value222255"})
	q.NoError(err)

	var result []Dummy
	err = q.im.SearchNSorts(mockCTX, mockTable, 0, 5, []string{"dummy", "updatekey"}, bson.M{"dummy": "test-value222232"}, &result)
	q.NoError(err)
	q.Equal(mockDummyValue, result)
}

func (q *querySuite) TestRemove() {
	type Dummy struct {
		Dummy  string `json:"dummy" bson:"dummy"`
		Update string `json:"updatekey" bson:"updatekey"`
	}

	mockDummyValue := Dummy{"test-value222232", "test-value222255"}

	// First set-insert
	err := q.im.Upsert(mockCTX, mockTable, bson.M{"dummy": "test-value222232"}, bson.M{"dummy": "test-value222232", "updatekey": "test-value222255"})
	q.NoError(err)

	result := &Dummy{}
	err = q.im.FindOne(mockCTX, mockTable, bson.M{"dummy": "test-value222232"}, result)
	q.NoError(err)
	q.Equal(mockDummyValue, *result)

	err = q.im.Remove(mockCTX, mockTable, bson.M{"dummy": "test-value222232"})
	q.NoError(err)
	result = &Dummy{}
	err = q.im.FindOne(mockCTX, mockTable, bson.M{"dummy": "test-value222232"}, result)
	q.Equal(err, ErrNotFound)
}

func (q *querySuite) TestRemoveAll() {
	type Dummy struct {
		Dummy  string `json:"dummy" bson:"dummy"`
		Update string `json:"updatekey" bson:"updatekey"`
	}

	mockDummyValue := Dummy{"test-value222232", "test-value222255"}
	mockDummyValue2 := Dummy{"test-value222233", "test-value222255"}

	err := q.im.Upsert(mockCTX, mockTable, bson.M{"dummy": "test-value222232"}, bson.M{"dummy": "test-value222232", "updatekey": "test-value222255"})
	q.NoError(err)
	err = q.im.Upsert(mockCTX, mockTable, bson.M{"dummy": "test-value222233"}, bson.M{"dummy": "test-value222233", "updatekey": "test-value222255"})
	q.NoError(err)

	result := &Dummy{}
	err = q.im.FindOne(mockCTX, mockTable, bson.M{"dummy": "test-value222232"}, result)
	q.NoError(err)
	q.Equal(mockDummyValue, *result)
	result = &Dummy{}
	err = q.im.FindOne(mockCTX, mockTable, bson.M{"dummy": "test-value222233"}, result)
	q.NoError(err)
	q.Equal(mockDummyValue2, *result)

	cnt, err := q.im.RemoveAll(mockCTX, mockTable, bson.M{"updatekey": "test-value222255"})
	q.NoError(err)
	q.Equal(int64(2), cnt)
	result = &Dummy{}
	err = q.im.FindOne(mockCTX, mockTable, bson.M{"dummy": "test-value222232"}, result)
	q.Equal(err, ErrNotFound)
	err = q.im.FindOne(mockCTX, mockTable, bson.M{"dummy": "test-value222233"}, result)
	q.Equal(err, ErrNotFound)
}

func (q *querySuite) TestPatch() {
	type Dummy struct {
		Dummy  string `json:"dummy" bson:"dummy"`
		Update string `json:"updatekey" bson:"updatekey"`
	}

	mockDummyValue := Dummy{"test-value", "test-value2"}

	// First set
	err := q.im.Upsert(mockCTX, mockTable, bson.M{"dummy": "test-value"}, bson.M{"dummy": "test-value", "updatekey": "test-value2"})
	q.Require().NoError(err)
	v := &Dummy{}
	err = q.im.FindOne(mockCTX, mockTable, bson.M{"dummy": "test-value"}, v)
	q.Require().NoError(err)
	q.Require().Equal(mockDummyValue, *v)

	// Test update (Second set)
	mockDummyValue.Update = "test-value-3"
	err = q.im.Patch(mockCTX, mockTable, bson.M{"dummy": "test-value"}, mockDummyValue)
	q.Require().NoError(err)
	v = &Dummy{}
	err = q.im.FindOne(mockCTX, mockTable, bson.M{"dummy": "test-value"}, v)
	q.Require().NoError(err)
	q.Equal(mockDummyValue, *v)

	// Test update multiple value
	mockMultiDummyValue := []*Dummy{{"test-multi", "test-multi2"}, {"test-multi", "test-multi3"}}
	err = q.im.Insert(mockCTX, mockTable, bson.M{"dummy": "test-multi", "updatekey": "test-multi2"})
	q.Require().NoError(err)
	err = q.im.Insert(mockCTX, mockTable, bson.M{"dummy": "test-multi", "updatekey": "test-multi3"})
	q.Require().NoError(err)
	v2 := []*Dummy{}
	err = q.im.Search(mockCTX, mockTable, 0, 100, "dummy", bson.M{"dummy": "test-multi"}, &v2)
	q.Require().NoError(err)
	q.Equal(mockMultiDummyValue, v2)

	for idx := range mockMultiDummyValue {
		mockMultiDummyValue[idx].Update = "test-multi4"
	}
	err = q.im.Patch(mockCTX, mockTable, bson.M{"dummy": "test-multi"}, bson.M{"updatekey": "test-multi4"}, WithPatchMany(true))
	q.Require().NoError(err)

	v2 = []*Dummy{}
	err = q.im.Search(mockCTX, mockTable, 0, 100, "dummy", bson.M{"dummy": "test-multi"}, &v2)
	q.Require().NoError(err)
	q.Equal(mockMultiDummyValue, v2)

	// Patch not exist document
	err = q.im.Patch(mockCTX, mockTable, bson.M{"dummy": "test-not-exist"}, bson.M{"updatekey": "test-multi4"}, WithPatchMany(true))
	q.Require().Error(err, ErrNotFound)
}

func (q *querySuite) TestCustomPatch() {
	type Dummy struct {
		Dummy  string `json:"dummy" bson:"dummy"`
		Update string `json:"updatekey" bson:"updatekey"`
		Status int32  `json:"status" bson:"status"`
		Record int32  `json:"record" bson:"record"`
	}

	mockDummyValue := Dummy{"test-value", "test-value2", 0, 0}

	// First set
	err := q.im.Upsert(mockCTX, mockTable, bson.M{"dummy": "test-value"}, bson.M{"dummy": "test-value", "updatekey": "test-value2", "status": 0, "record": 0})
	q.Require().NoError(err)

	// Test update (Second set)
	mockDummyValue.Update = "test-value-3"
	err = q.im.CustomPatch(mockCTX, mockTable, bson.M{"dummy": "test-value"}, bson.M{"$inc": bson.M{"record": 5}, "$bit": bson.M{"status": bson.M{"or": 2}}}, false)
	q.Require().NoError(err)
	v := &Dummy{}
	err = q.im.FindOne(mockCTX, mockTable, bson.M{"dummy": "test-value"}, v)
	q.Require().NoError(err)
	q.Equal(int32(2), v.Status)
	q.Equal(int32(5), v.Record)

	err = q.im.CustomPatch(mockCTX, mockTable, bson.M{"dummy": "test-value"}, bson.M{"$inc": bson.M{"record": 2}, "$bit": bson.M{"status": bson.M{"or": 1}}}, false)
	q.Require().NoError(err)
	err = q.im.FindOne(mockCTX, mockTable, bson.M{"dummy": "test-value"}, v)
	q.Require().NoError(err)
	q.Equal(int32(3), v.Status)
	q.Equal(int32(7), v.Record)

	err = q.im.CustomPatch(mockCTX, mockTable, bson.M{"dummy": "test-value"}, bson.M{"$inc": bson.M{"record": -1}, "$bit": bson.M{"status": bson.M{"or": 4}}}, false)
	q.NoError(err)
	err = q.im.FindOne(mockCTX, mockTable, bson.M{"dummy": "test-value"}, v)
	q.Require().NoError(err)
	q.Equal(int32(7), v.Status)
	q.Equal(int32(6), v.Record)

	// Test ErrNotFound
	err = q.im.CustomPatch(mockCTX, mockTable, bson.M{"dummy": "test-value-6"}, bson.M{"$set": bson.M{"record": 6}, "$setOnInsert": bson.M{"status": 3}}, false)
	q.Require().Error(err, ErrNotFound)

	// Test upsert
	err = q.im.CustomPatch(mockCTX, mockTable, bson.M{"dummy": "test-value-6"}, bson.M{"$set": bson.M{"record": 6}, "$setOnInsert": bson.M{"status": 3}}, true)
	q.NoError(err)
	v = &Dummy{}
	err = q.im.FindOne(mockCTX, mockTable, bson.M{"dummy": "test-value-6"}, v)
	q.Require().NoError(err)
	q.Equal(int32(3), v.Status)
	q.Equal(int32(6), v.Record)
}
func (q *querySuite) TestIncrement() {
	type Dummy struct {
		Dummy  string  `json:"dummy" bson:"dummy"`
		Update float64 `json:"updatekey" bson:"updatekey"`
	}

	mockDummyValue := Dummy{"test-value222dff", 1237.24156}

	// First set-insert
	err := q.im.Upsert(mockCTX, mockTable, bson.M{"dummy": "test-value222dff"}, bson.M{"dummy": "test-value222dff", "updatekey": 1234.1})
	q.NoError(err)

	result := &Dummy{}
	err = q.im.Increment(mockCTX, mockTable, bson.M{"dummy": "test-value222dff"}, result, "updatekey", 3.14156)
	q.NoError(err)
	q.Equal(mockDummyValue, *result)
}

func (q *querySuite) TestIncrementInsert() {
	type Dummy struct {
		Dummy  string  `bson:"dummy" json:"dummy"`
		Update float64 `bson:"updatekey" json:"updatekey"`
	}

	mockDummyValue := Dummy{"test-value222dff112", 3.14156}

	result := &Dummy{}
	err := q.im.Increment(mockCTX, mockTable, bson.M{"dummy": "test-value222dff112"}, result, "updatekey", 3.14156)
	q.NoError(err)
	q.Equal(mockDummyValue, *result)
}

func (q *querySuite) TestIncrementManyIncrement() {
	type Dummy struct {
		Dummy  string  `json:"dummy" bson:"dummy"`
		Set    string  `json:"setKey" bson:"setKey"`
		Update float64 `json:"updatekey" bson:"updatekey"`
	}

	mockDummyValue := Dummy{"test-value222dffSetIncrement", "wtfset", 1237.24156}

	// First set-insert
	err := q.im.Upsert(mockCTX, mockTable, bson.M{"dummy": "test-value222dffSetIncrement"}, bson.M{"dummy": "test-value222dffSetIncrement", "updatekey": 1234.1, "setKey": "wtfset"})
	q.NoError(err)

	result := &Dummy{}
	err = q.im.IncrementMany(mockCTX, mockTable, bson.M{"dummy": "test-value222dffSetIncrement"}, bson.M{"updatekey": 3.14156}, bson.M{"setKey": "wtfsetNotInsert"}, result)
	q.NoError(err)
	q.Equal(mockDummyValue, *result)
}

func (q *querySuite) testIncrementManyInsert() {
	type Dummy struct {
		Dummy  string  `json:"dummy" bson:"dummy"`
		Set    string  `json:"setKey" bson:"setKey"`
		Update float64 `json:"updatekey" bson:"updatekey"`
	}

	mockDummyValue := Dummy{"test-value222dffSetInsert", "wtfset", 3.14156}

	result := &Dummy{}
	err := q.im.IncrementMany(mockCTX, mockTable, bson.M{"dummy": "test-value222dffSetInsert"}, bson.M{"updatekey": 3.14156}, bson.M{"setKey": "wtfset"}, result)
	q.NoError(err)
	q.Equal(mockDummyValue, *result)
}

func (q *querySuite) TestIncrementManyInsert() {
	q.testIncrementManyInsert()
}

func (q *querySuite) testPush() {
	type Inner struct {
		A string `json:"A" bson:"A"`
	}
	type Dummy struct {
		Dummy  string   `json:"dummy" bson:"dummy"`
		Update []string `json:"updatekey" bson:"updatekey"`
		A      []Inner  `json:"inner" bson:"inner"`
	}

	mockDummyValue := Dummy{"test-value222dff", []string{"test"}, nil}
	mockDummyValue2 := Dummy{"test-value222dff", []string{"test"}, []Inner{Inner{A: "test"}}}

	// First set-insert
	err := q.im.Upsert(mockCTX, mockTable, bson.M{"dummy": "test-value222dff"}, bson.M{"dummy": "test-value222dff", "updatekey": []string{}})
	q.NoError(err)

	result := &Dummy{}
	err = q.im.Push(mockCTX, mockTable, bson.M{"dummy": "test-value222dff"}, result, "updatekey", "test")
	q.Require().NoError(err)
	q.Equal(mockDummyValue, *result)

	err = q.im.Push(mockCTX, mockTable, bson.M{"dummy": "test-value222dff"}, result, "inner", &Inner{A: "test"})
	q.Require().NoError(err)
	q.Equal(mockDummyValue2, *result)
}

func (q *querySuite) TestPush() {
	q.testPush()
}

func (q *querySuite) testPull() {
	type Inner struct {
		A string `json:"A" bson:"A"`
	}
	type Dummy struct {
		Dummy  string   `json:"dummy" bson:"dummy"`
		Update []string `json:"updatekey" bson:"updatekey"`
		A      []Inner  `json:"inner" bson:"inner"`
	}

	mockDummyValue := Dummy{"test-value222dff", []string{"test", "test"}, []Inner{}}
	mockDummyValue2 := Dummy{"test-value222dff", []string{}, []Inner{}}

	// First set-insert
	err := q.im.Upsert(
		mockCTX, mockTable,
		bson.M{"dummy": "test-value222dff"},
		bson.M{
			"dummy":     "test-value222dff",
			"updatekey": []string{"test", "test"},
			"inner":     []Inner{Inner{A: "test"}},
		},
	)
	q.NoError(err)

	result := &Dummy{}
	err = q.im.Pull(
		mockCTX, mockTable,
		bson.M{"dummy": "test-value222dff"},
		result,
		"inner", &Inner{A: "test"},
	)
	q.NoError(err)
	q.Equal(mockDummyValue, *result)

	err = q.im.Pull(
		mockCTX, mockTable,
		bson.M{"dummy": "test-value222dff"},
		result,
		"updatekey", "test",
	)
	q.NoError(err)
	q.Equal(mockDummyValue2, *result)
}

func (q *querySuite) TestPull() {
	q.testPull()
}

func (q *querySuite) TestBulkUpsert() {
	type Dummy struct {
		Dummy  string `json:"dummy" bson:"dummy"`
		Update string `json:"updatekey" bson:"updatekey"`
	}

	client := q.im.getClient(mockCTX)
	collection := client.Database(dbName).Collection(string(mockTable))

	// normal case-1: 2 insert operations
	op1 := UpsertOp{
		Selector: bson.M{"dummy": "test-value111-1"},
		Updater:  bson.M{"dummy": "test-value111-1", "updatekey": "test-value222-1"},
	}
	op2 := UpsertOp{
		Selector: bson.M{"dummy": "test-value111-2"},
		Updater:  bson.M{"dummy": "test-value111-2", "updatekey": "test-value222-2"},
	}
	mockDummyAns1 := Dummy{"test-value111-1", "test-value222-1"}
	mockDummyAns2 := Dummy{"test-value111-2", "test-value222-2"}
	ans := Dummy{}
	_, _, err := q.im.BulkUpsert(mockCTX, mockTable, []UpsertOp{op1, op2})
	q.Require().NoError(err)

	err = collection.FindOne(mockCTX, op1.Selector).Decode(&ans)
	q.NoError(err)
	q.Equal(mockDummyAns1, ans)
	err = collection.FindOne(mockCTX, op2.Selector).Decode(&ans)
	q.NoError(err)
	q.Equal(mockDummyAns2, ans)

	// normal case-2: 1 insert operation and 1 update operation
	op3 := UpsertOp{
		Selector: bson.M{"dummy": "test-value111-2"},
		Updater:  bson.M{"dummy": "test-value111-2", "updatekey": "test-value333-2"},
	}
	op4 := UpsertOp{
		Selector: bson.M{"dummy": "test-value111-4"},
		Updater:  bson.M{"dummy": "test-value111-4", "updatekey": "test-value333-4"},
	}
	mockDummyAns3 := Dummy{"test-value111-2", "test-value333-2"}
	mockDummyAns4 := Dummy{"test-value111-4", "test-value333-4"}
	_, _, err = q.im.BulkUpsert(mockCTX, mockTable, []UpsertOp{op3, op4})
	q.Require().NoError(err)
	err = collection.FindOne(mockCTX, op3.Selector).Decode(&ans)
	q.NoError(err)
	q.Equal(mockDummyAns3, ans)
	err = collection.FindOne(mockCTX, op4.Selector).Decode(&ans)
	q.NoError(err)
	q.Equal(mockDummyAns4, ans)

	//failure case: invalid selector
	op0 := UpsertOp{
		Selector: "badSelector",
		Updater:  bson.M{"dummy": "test-value111-1", "updatekey": "test-value222-1"},
	}
	_, _, err = q.im.BulkUpsert(mockCTX, mockTable, []UpsertOp{op0})
	q.Error(err)

	//failure case: invalid updater
	op0 = UpsertOp{
		Selector: bson.M{"dummy": "test-value111-1"},
		Updater:  "badUpdater",
	}
	_, _, err = q.im.BulkUpsert(mockCTX, mockTable, []UpsertOp{op0})
	q.Error(err)
}

func (q *querySuite) testPipe() {
	type Dummy struct {
		PrimaryKey string `bson:"primaryKey" json:"primaryKey"`
		Key1       string `bson:"key1" json:"key1"`
	}

	err := q.im.Upsert(
		mockCTX, mockTable, bson.M{"primaryKey": "primaryValue1"},
		bson.M{"primaryKey": "primaryValue1", "key1": "value1"},
	)
	q.NoError(err)

	err = q.im.Upsert(
		mockCTX, mockTable, bson.M{"primaryKey": "primaryValue2"},
		bson.M{"primaryKey": "primaryValue2", "key1": "value2"},
	)
	q.NoError(err)

	err = q.im.Upsert(
		mockCTX, mockTable, bson.M{"primaryKey": "primaryValue3"},
		bson.M{"primaryKey": "primaryValue3", "key1": "value2"},
	)
	q.NoError(err)

	match := bson.M{
		"key1": "value2",
	}
	iter, fnClose, err := q.im.Pipe(mockCTX, mockTable, []bson.M{
		bson.M{"$match": match},
	})
	q.NoError(err)
	defer fnClose()

	var result []Dummy
	for {
		d := Dummy{}
		ok, err := iter.Next(mockCTX, &d)
		q.NoError(err)
		if !ok {
			break
		}
		result = append(result, d)
	}

	mockDummyValue := []Dummy{
		{"primaryValue2", "value2"},
		{"primaryValue3", "value2"},
	}
	q.Equal(mockDummyValue, result)

	// Test iter.All
	var allResult []Dummy
	iter2, fnClose2, err := q.im.Pipe(mockCTX, mockTable, []bson.M{
		bson.M{"$match": match},
	})
	q.NoError(err)
	defer fnClose2()
	q.Require().NoError(iter2.All(mockCTX, &allResult))
	q.Equal(mockDummyValue, allResult)
}

func (q *querySuite) TestPipe() {
	q.testPipe()
}

func (q *querySuite) TestRunWithTransaction() {
	type Dummy struct {
		Dummy string `json:"dummy" bson:"dummy"`
	}

	run := func(c ctx.Ctx) error {
		err := q.im.Insert(c, mockTable, bson.M{"dummy": "test-value-1"})
		fmt.Printf("=======%T=====\n", err)
		q.Require().NoError(q.im.Insert(c, mockTable, bson.M{"dummy": "test-value-1"}))
		q.Require().NoError(q.im.Insert(c, mockTable, bson.M{"dummy": "test-value-2"}))
		return errors.New("error")
	}

	// test fail
	err := q.im.RunWithTransaction(mockCTX, run)
	q.Require().Error(err, "error")

	result := &Dummy{}
	err = q.im.FindOne(mockCTX, mockTable, bson.M{"dummy": "test-value-1"}, result)
	q.Equal(err, ErrNotFound)

	err = q.im.FindOne(mockCTX, mockTable, bson.M{"dummy": "test-value-2"}, result)
	q.Equal(err, ErrNotFound)

	run = func(c ctx.Ctx) error {
		q.Require().NoError(q.im.Insert(c, mockTable, bson.M{"dummy": "test-value-1"}))
		q.Require().NoError(q.im.Insert(c, mockTable, bson.M{"dummy": "test-value-2"}))
		return nil
	}

	// test success
	err = q.im.RunWithTransaction(mockCTX, run)
	q.Require().NoError(err)

	err = q.im.FindOne(mockCTX, mockTable, bson.M{"dummy": "test-value-1"}, result)
	q.Require().NoError(err)
	q.Require().Equal("test-value-1", result.Dummy)

	err = q.im.FindOne(mockCTX, mockTable, bson.M{"dummy": "test-value-2"}, result)
	q.Require().NoError(err)
	q.Require().Equal("test-value-2", result.Dummy)
}

func TestQuerySuite(t *testing.T) {
	q := new(querySuite)

	suite.Run(t, q)
}
