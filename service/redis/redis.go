package redis

/*
	Persion in charge:
		abner

	Description:
		Package redis adds an abstraction layer for the 3rd party redis package, redigo.

	Algorithm and internal structure:
		1. Most methods are just simple wrappers for the underlying redis calls.

		2. MGet/MGetZip/Del
			The keys in MGet will be splitted into batches whose size is 100 keys
			to prevent performance issues in redis which is suggested by our redis
			vendor.

		3. SetStruct/GetStruct
			SetStruct Stored struct variable as Hashmap in redis
			{
				Category: "zoo_topic_guess",
				Streamer: {ID: 1},
				StreamerID: 1,
			}
			will be stored as
			1) "Category"
			2) "\"zoo_topic_guess\""
 			3) "Streamer"
 			4) "{\"ID\":1}"
 			5) "StreamerID"
			6) "1"

			GetStruct get datas and unmarshal back into struct variable


	NOTE: Be aware the usages for HGetAll and SMembers which may have potential
	performance issues if the number of members is too large.

*/

import (
	"errors"
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"

	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/metrics"
	"github.com/x-xyz/goapi/domain/keys"
)

// ScriptHdl is handler for redis script
type ScriptHdl struct {
	keyCount int
	*redis.Script
}

func (hdl *ScriptHdl) prefix(keysAndArgs ...interface{}) string {
	if hdl.keyCount > 0 && len(keysAndArgs) > 0 {
		// if keyCount > 1, expecting the keys are generated by keys.RedisLuaMultiKey and have same prefix.
		if key, ok := keysAndArgs[0].(string); ok {
			return keys.GetPrefix(key)
		}
	}
	return metrics.TagValueNA
}

// MVal is structure for redis multiple return values
type MVal struct {
	Valid bool
	Value []byte
}

// ZVal is for ZRevrange/ZRange with scores return values
type ZVal struct {
	Value string
	Score int
}

// ZFloatVal is for ZRevrange/ZRange with floating scores return values
type ZFloatVal struct {
	Value string
	Score float64
}

const (
	// Forever means no ttl and the key will last forever. Caller should handle key eviction by himself.
	Forever = time.Duration(-1)

	// PersistName is used by paging v1
	PersistName = "persistent"
)

var (
	// ErrNotFound is the error when redis can't get value from key
	ErrNotFound = redis.ErrNil

	// ErrNoTTL is the error while existed key has no associated expire
	ErrNoTTL = fmt.Errorf("no ttl")

	// ErrGapTime is the error returned while in the transition of redis endpoints
	ErrGapTime = fmt.Errorf("gap time")

	// ErrNoSuchKey is the error when renaming a non-exist key
	ErrNoSuchKey = redis.Error("ERR no such key")

	// ErrExpireNotExistOrTimeout indicates returned will be 0 if key does not exist or the timeout could not be set.
	ErrExpireNotExistOrTimeout = errors.New("reply from EXPIRE is not 1")
)

// Service implements a redis interface
type Service interface {
	// GetSet Atomically sets key to value and returns the old value stored at key
	GetSet(context ctx.Ctx, key string, val []byte, ttl time.Duration) ([]byte, error)

	// Set sets a key with a value
	Set(context ctx.Ctx, key string, val []byte, ttl time.Duration) error

	// ScanMatch runs the SCAN command with MATCH and COUNT argument.
	// It returns new cursor, an array of keys, and error.
	ScanMatch(context ctx.Ctx, cursor int64, match string, count int) (int64, []string, error)

	// MSet sets multiple key-value pairs at the same time.
	MSet(context ctx.Ctx, keyVals map[string][]byte, ttl time.Duration) error

	// SetZip performs gzip before Set operation, you must use GetZip for the
	// corresponding values.
	SetZip(context ctx.Ctx, key string, val []byte, ttl time.Duration) error

	// SetXX sets a key with a value only the key already exist, expire mean expire time for this key.
	SetXX(context ctx.Ctx, key string, val []byte, ttl time.Duration) error

	// SetNX sets a key if not exist with a value. It returns ErrNotFound if key
	// already exists.
	SetNX(context ctx.Ctx, key string, val []byte, expire time.Duration) error

	// Expire set a expire time to a key.
	Expire(context ctx.Ctx, key string, ttl time.Duration) error

	// Get gets the value of a key
	Get(context ctx.Ctx, key string) (val []byte, err error)

	// GetZip gets the gzipped values stored by SetZip and ungzips it before
	// return to the caller.
	GetZip(context ctx.Ctx, key string) (val []byte, err error)

	// MGet gets values of a set of keys
	// If key does not exist, you will not get ErrNotFound
	// You will get false value in `Valid` field in return MVal
	MGet(context ctx.Ctx, keys []string) ([]MVal, error)

	// MGetZip performs MGET on gzipped values from SetZip, it will decompress
	// it before return.
	// If key does not exist, you will not get ErrNotFound
	// You will get false value in `Valid` field in return MVal
	MGetZip(context ctx.Ctx, keys []string) ([]MVal, error)

	// Del Removes the specified keys and return the number of keys that were removed.
	// A key is ignored if it does not exist.
	Del(context ctx.Ctx, keys ...string) (int, error)

	// Unlink Removes the specified keys and return the number of keys that were removed.
	// It performs the actual memory reclaiming in a different thread, so it is not blocking, while DEL is.
	// A key is ignored if it does not exist.
	// Note: Available since Redis 4.0.0.
	Unlink(context ctx.Ctx, key ...string) (int, error)

	// HSet sets field in the hash stored at key to value , expire mean expire time for this key.
	HSet(context ctx.Ctx, key, field string, val []byte, ttl time.Duration) error

	// HSetNX sets field in the hash stored at key to value, only if field does not yet exist. expire mean expire time for this key.
	// if field already set, you will get false value in return bool and no operation was performed
	HSetNX(context ctx.Ctx, key, field string, val []byte, ttl time.Duration) (bool, error)

	// SetStruct sets struct into hash, using fields of first layer struct as fields of redis hash
	// json tag does not work for the first layer
	// second layer of nested struct will be stored as json
	SetStruct(context ctx.Ctx, key string, val interface{}, ttl time.Duration) error

	// GetStruct get data from redis hash and set into val, val should be pointer to struct variable
	GetStruct(context ctx.Ctx, key string, val interface{}) error

	// HMSet sets field in the hash stored at key to value , expire mean expire time for this key.
	HMSet(context ctx.Ctx, key string, fieldVal map[string][]byte, ttl time.Duration) error

	// HGet gets the value of a hash field
	HGet(context ctx.Ctx, key, field string) (val []byte, err error)

	// HMGet gets values of hash fields
	// If key does not exist, you will not get ErrNotFound
	// You will get false value in `Valid` field in return MVal
	HMGet(context ctx.Ctx, key string, fields ...string) ([]MVal, error)

	// HGetAll gets all the fields and values in a hash
	HGetAll(context ctx.Ctx, key string) (map[string][]byte, error)

	// HDel Removes the specified fields from the hash stored at key.
	HDel(context ctx.Ctx, key, field string) (int, error)

	// HLen Returns the number of fields contained in the hash stored at key.
	HLen(context ctx.Ctx, key string) (length int, err error)

	// Incr Increments the number stored at key by one. If the key does not exist, it is set to 0 before performing the operation.
	Incr(context ctx.Ctx, key string) (int64, error)

	// // Incrby Increments the number stored at key by increment.
	Incrby(context ctx.Ctx, key string, val int) (int64, error)

	// HIncrby Increments the number stored at Hash field by increment.
	HIncrby(context ctx.Ctx, key, field string, val int) (int64, error)

	// // HScan return count entrys
	HScan(context ctx.Ctx, key string, cursor, count int) (map[string][]byte, int, error)

	// ZAdd Adds all the specified members with the specified scores to the sorted set stored at key.
	ZAdd(context ctx.Ctx, key string, memSco map[string]int) error

	// ZAddFloat Adds all the specified members with the specified floating scores to the sorted set stored at key.
	ZAddFloat(context ctx.Ctx, key string, memSco map[string]float64) error

	// ZAddXX invokes ZAdd and only update elements that already exist. Never add elements.
	ZAddXX(context ctx.Ctx, key string, memSco map[string]int) error

	// ZAddNXFloat invokes ZAdd but doesn't update already existing elements. Always add new elements.
	ZAddNXFloat(context ctx.Ctx, key string, memSco map[string]float64) error

	// ZScore Returns the score of member in the sorted set at key.
	ZScore(context ctx.Ctx, key string, member string) (val int, err error)

	// ZScoreFloat Returns the floating score of member in the sorted set at key.
	ZScoreFloat(context ctx.Ctx, key string, member string) (val float64, err error)

	// Zincrby Increments the score of member in the sorted set stored at key by increment.
	ZIncrby(context ctx.Ctx, key string, member string, val int) (result int, err error)

	// ZincrbyFloat Increments the floating score of member in the sorted set stored at key by increment.
	ZIncrbyFloat(context ctx.Ctx, key string, member string, val float64) (result float64, err error)

	// Zscan the limit numbers of entrys
	ZScan(context ctx.Ctx, key string, cursor, limit int) (map[string]int, int, error)

	// ZCard Returns the sorted set cardinality (number of elements) of the sorted set stored at key.
	ZCard(context ctx.Ctx, key string) (count int, err error)

	// ZCount returns number of element in zset with score between minScore and maxScore
	ZCount(context ctx.Ctx, key, minScore, maxScore string) (int, error)

	// ZRevrange Returns the specified range of elements in the sorted set stored at key.(From biggest)
	ZRevrange(context ctx.Ctx, key string, offset, count int) ([]string, error)

	// ZRange Returns the specified range of elements in the sorted set stored at key.(From smallest)
	// If key not found, will return empty slice with no error
	ZRange(context ctx.Ctx, key string, offset, count int) ([]string, error)

	// ZRangeByScoreWithScore returns specified range of elements in sorted list within min and max scores
	ZRangeByScoreWithScore(context ctx.Ctx, key, minScore, maxScore string) ([]ZVal, error)

	// ZRevrangeScore Returns the specified range of elements in the sorted set stored at key with score
	// redis> ZREVRANGE salary 0 -1 WITHSCORES
	// 1) "jack"
	// 2) "5000"
	// 3) "tom"
	// 4) "4000"
	// 5) "peter"
	// 6) "3500"
	ZRevrangeScore(context ctx.Ctx, key string, offset, count int) ([]ZVal, error)

	// ZRevrangeFloatScore Returns the specified range of elements in the sorted set stored at key with floating score
	ZRevrangeFloatScore(context ctx.Ctx, key string, offset, count int) ([]ZFloatVal, error)

	// ZRevrangeByScoreWithScore return specified range of elements in sorted list within min and max scores
	ZRevrangeByScoreWithScore(context ctx.Ctx, key, minScore, maxScore string) ([]ZVal, error)

	// ZRevrangeByScoreWithFloatScore return specified range of elements in sorted list within min and max scores in floating
	ZRevrangeByScoreWithFloatScore(context ctx.Ctx, key, minScore, maxScore string) ([]ZFloatVal, error)

	// ZRem removes the specified members from the sorted set stored at key
	ZRem(context ctx.Ctx, key string, members ...string) error

	// ZRemRangeByScore removes all elements in the sorted set stored at key with a score between min and max (inclusive)
	ZRemRangeByScore(context ctx.Ctx, key string, minScore, maxScore int) (int, error)

	// ZRemRangeByRank Removes all elements in the sorted set stored at key with rank between start and stop
	ZRemRangeByRank(context ctx.Ctx, key string, start, stop int) (int, error)

	// ZRevRank returns member's rank with the scores ordered from high to low
	// The member with the highest score has rank 0
	// If member does not exist in the sorted set or key does not exist, ErrNotFound will be returned
	ZRevRank(context ctx.Ctx, key, member string) (int, error)

	// ZPopMin removes and returns up to count members with the lowest scores in the sorted set stored at key.
	// Note: Available since Redis 5.0.0.
	ZPopMin(context ctx.Ctx, key string, count int) ([]ZFloatVal, error)

	// ZUnionStore Returns the union of sorted sets
	// given by the specified paris (key and weight), and stores the result in destination.
	// len(paris) should be longer than or equal to 2
	//
	// Benchmark with increasing scores for 5000 members
	// go test -bench=BenchmarkZUnionStore -benchtime=1s
	// result: 	300	   	   3989611 ns/op
	// go test -bench=BenchmarkZIncrby -benchtime=1s
	// result:    1		2759058805 ns/op
	ZUnionStore(context ctx.Ctx, paris []Pair, dest string) error

	// SAdd add a member to the set
	SAdd(context ctx.Ctx, key string, member ...string) error

	// SAddFullInfo add members to set, and return # of members added into set
	SAddFullInfo(context ctx.Ctx, key string, member ...string) (int64, error)

	// SRem remove a member from the set
	SRem(context ctx.Ctx, key string, member ...string) error

	// SPop pop a member from the set
	SPop(context ctx.Ctx, key string) (string, error)

	// SMPop pops members from the set up to count
	SMPop(context ctx.Ctx, key string, count int) ([]string, error)

	// // LTrim Trim an existing list so that it will contain only the specified range of elements specified.
	LTrim(context ctx.Ctx, key string, start, end int) error

	// LPush Insert all the specified values at the head of the list stored at key
	LPush(context ctx.Ctx, key string, val []byte) error

	// RPush Insert all the specified values at the tail of the list stored at key
	// Returns the size of the pushed list
	RPush(context ctx.Ctx, key string, val []byte) (int, error)

	// LPop Pop a item at head of the list stored at key
	LPop(context ctx.Ctx, key string) ([]byte, error)

	// RPop Pop a item at tail of the list stored at key
	RPop(context ctx.Ctx, key string) ([]byte, error)

	// LSet Set a item at specified index of element of the list stored at key
	LSet(context ctx.Ctx, key string, index int, val []byte) error

	// LRange Returns the specified elements of the list stored at key
	LRange(context ctx.Ctx, key string, offset, count int) (val [][]byte, err error)

	// LINDEX Returns the specified index of element of the list stored at key
	LIndex(context ctx.Ctx, key string, index int64) (val []byte, err error)

	// LLEN Returns the number of item in the list stored at key
	LLen(context ctx.Ctx, key string) (length int, err error)

	// LINSER Insert all the specified values before the specified values
	LInsert(context ctx.Ctx, key string, before, val []byte) error

	// SMembers returns all members of the set
	SMembers(context ctx.Ctx, key string) ([]string, error)

	// SIsMember Returns if member is a member of the set stored at key.
	SIsMember(context ctx.Ctx, key, member string) (bool, error)

	// SCard Returns the set cardinality (number of elements) of the set stored at key.
	SCard(context ctx.Ctx, key string) (int, error)

	// SScan returns the members of key
	SScan(context ctx.Ctx, key string, cursor, count int) (members []string, nextCursor int, err error)

	// PFAdd adds all the element arguments to the HyperLogLog data structure stored at the variable name specified as first argument.
	// If the approximated cardinality estimated by the HyperLogLog changed after executing the command, PFADD returns 1, otherwise 0 is returned.
	// The command automatically creates an empty HyperLogLog structure if the specified key does not exist.
	PFAdd(context ctx.Ctx, key string, members ...string) (int, error)

	// Rename renames 'oldKey' by 'newkey' and will overwrite allready exist 'newKey''s value.
	// Return ErrNoSuchKey if 'oldKey' does not exist.
	Rename(context ctx.Ctx, oldKey, newKey string) error

	// Exists Returns if the key exists.
	Exists(context ctx.Ctx, key string) (bool, error)

	// // TTL returns key's ttl in terms of second
	TTL(context ctx.Ctx, key string) (int, error)

	// ScriptDo evaluates the script
	ScriptDo(context ctx.Ctx, hdl *ScriptHdl, keysAndArgs ...interface{}) (interface{}, error)

	//GetConn return a redis conn
	GetConn() (redis.Conn, error)

	// Name return redis name
	Name() string

	// RandomKey return a random key from the currently selected database.
	RandomKey(context ctx.Ctx) ([]byte, error)

	// Type return key's type
	// The different types that can be returned are: string, list, set, zset, hash and stream
	Type(context ctx.Ctx, key string) ([]byte, error)

	// Strlen returns the length of the string value stored at key.
	// An error is returned when key holds a non-string value.
	Strlen(context ctx.Ctx, key string) (int, error)
}

// NewScript creates redis script
func NewScript(keyCount int, src string) *ScriptHdl {
	return &ScriptHdl{
		keyCount: keyCount,
		Script:   redis.NewScript(keyCount, src),
	}
}

// Int is a helper function to convert the redis output into int
func Int(i interface{}, err error) (int, error) {
	return redis.Int(i, err)
}

// Ints is a helper function to convert the redis output into []int
func Ints(i interface{}, err error) ([]int, error) {
	return redis.Ints(i, err)
}

// Bool is a helper function to convert the redis output into bool
func Bool(reply interface{}, err error) (bool, error) {
	return redis.Bool(reply, err)
}

// Int64 is a helper function to convert the redis output into int64
func Int64(reply interface{}, err error) (int64, error) {
	return redis.Int64(reply, err)
}

// Strings is a helper function to convert the redis output into []string
func Strings(reply interface{}, err error) ([]string, error) {
	return redis.Strings(reply, err)
}

// String is a helper function to convert the redis output into string
func String(reply interface{}, err error) (string, error) {
	return redis.String(reply, err)
}

// Values is a helper that converts an array command reply to a []interface{}.
// If err is not equal to nil, then Values returns nil, err. Otherwise, Values
// converts the reply as follows:
//
//  Reply type      Result
//  array           reply, nil
//  nil             nil, ErrNil
//  other           nil, error
func Values(reply interface{}, err error) ([]interface{}, error) {
	return redis.Values(reply, err)
}

// StringMap is a helper that converts an array of strings (alternating key, value)
// into a map[string]string. The HGETALL and CONFIG GET commands return replies in this format.
// Requires an even number of values in result.
func StringMap(reply interface{}, err error) (map[string]string, error) {
	return redis.StringMap(reply, err)
}

// Bytes is a helper function to convert the redis output into []bytes
func Bytes(reply interface{}, err error) ([]byte, error) {
	return redis.Bytes(reply, err)
}
