package redis

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"reflect"
	"strconv"
	"time"

	"github.com/gomodule/redigo/redis"

	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/log"
	"github.com/x-xyz/goapi/base/metrics"
	"github.com/x-xyz/goapi/domain/keys"
)

const (
	redisStatsInterval = 30 * time.Second

	// retTTLNoKey is the return value of TTL when the key does not exist
	retTTLNoKey = -2

	// retTTLNoExpire is the return value of TTL when the key exists but has
	// no associated expire
	retTTLNoExpire = -1

	keyAttribute   = "key"
	matchAttribute = "match"

	switchTargetSrc = "src"
	switchTargetDst = "dst"
	switchTargetBak = "bak"

	healthCheckPeriod = time.Second
)

var (
	mgetBatchSize = 100 // redis lab recommended
	delBatchSize  = 100
	timeNow       = time.Now
)

// Pair is used in ZUnionStore which contains Key and Weight pair
// The default value of Weight will be 1
type Pair struct {
	Key    string
	Weight float64
}

type redImpl struct {
	name  string
	met   metrics.Service
	pools *Pools
}

// Pools represents different pool types
type Pools struct {
	Src *redis.Pool
}

// New redis pool
func New(name string, metrics metrics.Service, pools *Pools) Service {
	im := &redImpl{
		name:  name,
		met:   metrics,
		pools: pools,
	}

	return im
}

func (r *redImpl) getConn(command string) (redis.Conn, error) {
	defer r.met.BumpTime("getconn.time", "cluster", r.name).End()
	var conn redis.Conn

	pool := r.getPool(command)
	if pool == nil {
		return nil, ErrGapTime
	}

	conn = pool.Get()
	if err := conn.Err(); err != nil {
		r.met.BumpSum("getConn.err", 1, "cluster", r.name, "reason", err.Error())
		return nil, err
	}

	return conn, nil
}

func (r *redImpl) getPool(command string) *redis.Pool {
	return r.pools.Src
}

func (r *redImpl) connDo(context ctx.Ctx, commandName string, args ...interface{}) (interface{}, error) {
	conn, err := r.getConn(commandName)
	if err != nil {
		return nil, err
	}

	// FIXME: add trace
	// context, span := ctx.StartSpan(context, fmt.Sprintf("redis.%s", commandName))
	// if commandName == "SCAN" {
	// 	span.AddAttributes(
	// 		trace.StringAttribute(matchAttribute, args[1].(string)),
	// 	)
	// } else if len(args) != 0 && reflect.TypeOf(args[0]).Kind() == reflect.String {
	// 	span.AddAttributes(
	// 		trace.StringAttribute(keyAttribute, args[0].(string)),
	// 	)
	// }

	reply, err := conn.Do(commandName, args...)
	// span.End()

	// Closing conn explicitly asap improves redigo's performance,
	// bacause longer an connection is hold and not closed, the
	// pool need to handle more connections at the same time and
	// getConn time might burst.
	if err := conn.Close(); err != nil {
		r.met.BumpSum("conn.Close.err", 1, "cluster", r.name)
	}
	return reply, err
}

func (r *redImpl) get(context ctx.Ctx, key string, zip bool) (val []byte, err error) {
	funcName := "get"
	if zip == true {
		funcName = "getzip"
	}

	tags := []string{
		"func", funcName,
		"cluster", r.name,
		"prefix", keys.GetPrefix(key),
	}
	defer r.met.BumpTime("time", tags...).End()

	val, err = redis.Bytes(r.connDo(context, "GET", key))
	r.met.BumpHistogram("bytes", float64(len(val)), tags...)
	if err != nil {
		return nil, err
	}

	defer r.met.BumpTime("postprocess.time", tags...).End()
	if !zip {
		r.met.BumpHistogram("gzip", float64(0), tags...)
		return val, err
	}

	buf := bytes.NewBuffer(val)
	rb, err := gzip.NewReader(buf)
	if err != nil {
		context.WithField("err", err).Warn("new gzip reader failed")
		r.met.BumpHistogram("gzip", float64(0), tags...)
		return val, nil
	}
	res, err := ioutil.ReadAll(rb)
	rb.Close()
	r.met.BumpHistogram("gzip", float64(1), tags...)
	return res, err
}

func (r *redImpl) Get(context ctx.Ctx, key string) (val []byte, err error) {
	return r.get(context, key, false)
}

func (r *redImpl) GetSet(context ctx.Ctx, key string, val []byte, expire time.Duration) ([]byte, error) {
	tags := []string{"func", "getset", "cluster", r.name, "prefix", keys.GetPrefix(key)}
	defer r.met.BumpTime("time", tags...).End()
	if expire == Forever {
		r.met.BumpSum("ttl.forever", 1, tags...)
	} else {
		r.met.BumpAvg("ttl", expire.Seconds(), tags...)
	}
	r.met.BumpHistogram("bytes", float64(len(val)), tags...)
	res, err := redis.Bytes(r.connDo(context, "GETSET", key, val))
	if err != nil {
		context.WithField("err", err).Error("GetSet redis failed")
		return nil, err
	}
	if expire == Forever {
		_, err := r.connDo(context, "PERSIST", key)
		if err != nil {
			context.WithField("err", err).Warn("GetSet PERSIST redis key failed")
		}
	} else {
		_, err = r.connDo(context, "PEXPIRE", key, int(expire/time.Millisecond))
		if err != nil {
			context.WithField("err", err).Warn("GetSet PEXPIRE redis key failed")
		}
	}
	return res, nil

}

func (r *redImpl) GetZip(context ctx.Ctx, key string) (val []byte, err error) {
	return r.get(context, key, true)
}

func (r *redImpl) mget(context ctx.Ctx, ks []string, zip bool) ([]MVal, error) {
	if len(ks) == 0 {
		return []MVal{}, nil
	} else if len(ks) == 1 {
		// using GET instead of MGET with one key has lower resource consumption
		value, err := r.get(context, ks[0], zip)
		if err == ErrNotFound {
			return []MVal{
				{
					Valid: false,
					Value: []byte(""),
				},
			}, nil
		} else if err != nil {
			return nil, err
		}
		return []MVal{
			{
				Valid: true,
				Value: value,
			},
		}, nil
	}

	funcName := "mget"
	if zip == true {
		funcName = "mgetzip"
	}
	tags := []string{
		"func", funcName,
		"cluster", r.name,
		"prefix", keys.GetPrefix(ks[0]),
	}
	defer r.met.BumpTime("time", tags...).End()
	r.met.BumpHistogram("elements", float64(len(ks)), tags...)

	// Retrieve value in batch
	vals, err := r.mgetBatch(context, ks, mgetBatchSize)
	if err != nil {
		context.WithField("err", err).Error("MGET redis failed")
		r.met.BumpHistogram("gzip", float64(0), tags...)
		return nil, err
	}

	// Process result of mgetBatch
	return r.processMgetValues(context, vals, zip, []string{}), nil
}

func (r *redImpl) processMgetValues(context ctx.Ctx, values []interface{}, zip bool, tags []string) []MVal {
	defer r.met.BumpTime("postprocess.time", tags...).End()
	// add counters to record length of result form redis.MGet.
	size := 0
	mvals := []MVal{}
	for k := range values {
		if values[k] == nil {
			mvals = append(mvals, MVal{
				Valid: false,
				Value: []byte(""),
			})
			continue
		}

		mval := MVal{Valid: true}

		if zip {
			var err error
			mval.Value, err = r.unzip(values[k].([]byte))
			if err != nil {
				context.WithField("key", k).Error("not a gzip format.")
				mvals = append(mvals, MVal{
					Valid: false,
					Value: []byte(""),
				})
				continue
			}
		} else {
			mval.Value = values[k].([]byte)
		}

		size += len(mval.Value)
		mvals = append(mvals, mval)
	}

	if len(values) > 0 {
		r.met.BumpHistogram("bytes", float64(size/len(values)), tags...)
	}

	if zip {
		r.met.BumpHistogram("gzip", float64(1), tags...)
	} else {
		r.met.BumpHistogram("gzip", float64(0), tags...)
	}

	return mvals
}

func (r *redImpl) mgetBatch(context ctx.Ctx, keys []string, mgetBatchSize int) ([]interface{}, error) {
	vals := []interface{}{}
	for i := 0; i < len(keys); i += mgetBatchSize {
		bottom := i
		upper := i + mgetBatchSize
		if upper > len(keys) {
			upper = len(keys)
		}

		v, err := redis.Values(r.connDo(context, "MGET", redis.Args{}.AddFlat(keys[bottom:upper])...))
		if err != nil {
			return nil, err
		}

		vals = append(vals, v...)
	}

	return vals, nil
}

func (r *redImpl) unzip(value []byte) ([]byte, error) {
	buf := bytes.NewBuffer(value)
	rb, err := gzip.NewReader(buf)
	if err != nil {
		return nil, err
	}
	defer rb.Close()

	return ioutil.ReadAll(rb)
}

func (r *redImpl) MGet(context ctx.Ctx, keys []string) ([]MVal, error) {
	return r.mget(context, keys, false)
}

func (r *redImpl) MGetZip(context ctx.Ctx, keys []string) ([]MVal, error) {
	return r.mget(context, keys, true)
}

func (r *redImpl) Del(context ctx.Ctx, ks ...string) (int, error) {
	if len(ks) == 0 {
		return 0, fmt.Errorf("length of keys is 0")
	}

	tags := []string{"func", "del", "cluster", r.name, "prefix", keys.GetPrefix(ks[0])}
	defer r.met.BumpTime("time", tags...).End()
	r.met.BumpHistogram("elements", float64(len(ks)), tags...)

	affected := 0
	for i := 0; i < len(ks); i += delBatchSize {
		start := i
		end := i + delBatchSize
		if end > len(ks) {
			end = len(ks)
		}
		res, err := redis.Int(r.connDo(context, "DEL", redis.Args{}.AddFlat(ks[start:end])...))
		if err != nil {
			context.WithField("err", err).Error("DEL redis failed")
			return 0, err
		}
		affected += res
	}

	return affected, nil
}

func (r *redImpl) Unlink(context ctx.Ctx, ks ...string) (int, error) {
	if len(ks) == 0 {
		return 0, fmt.Errorf("length of keys is 0")
	}

	tags := []string{"func", "unlink", "cluster", r.name, "prefix", keys.GetPrefix(ks[0])}
	defer r.met.BumpTime("time", tags...).End()
	r.met.BumpHistogram("elements", float64(len(ks)), tags...)

	affected, err := redis.Int(r.connDo(context, "UNLINK", redis.Args{}.AddFlat(ks)...))
	if err != nil {
		context.WithField("err", err).Error("UNLINK redis failed")
		return 0, err
	}
	return affected, nil
}

func (r *redImpl) set(context ctx.Ctx, key string, val []byte, expire time.Duration, zip bool) error {
	funcName := "set"
	if zip == true {
		funcName = "setzip"
	}
	tags := []string{
		"func", funcName,
		"cluster", r.name,
		"prefix", keys.GetPrefix(key),
	}
	defer r.met.BumpTime("time", tags...).End()
	if expire == Forever {
		r.met.BumpSum("ttl.forever", 1, tags...)
	} else {
		r.met.BumpAvg("ttl", expire.Seconds(), tags...)
	}

	var newVal []byte
	timer := r.met.BumpTime("preprocess.time", tags...)
	if zip {
		buf := &bytes.Buffer{}
		writer := gzip.NewWriter(buf)
		writer.Write(val)
		writer.Flush()
		writer.Close()
		b := buf.Bytes()
		newVal = append(newVal, b...)
		r.met.BumpHistogram("gzip", float64(1), tags...)
	} else {
		newVal = append(newVal, val...)
		r.met.BumpHistogram("gzip", float64(0), tags...)
	}
	timer.End()

	if expire == Forever {
		_, err := r.connDo(context, "SET", key, newVal)
		if err != nil {
			context.WithField("err", err).Error("set redis failed")
		}
		return err
	}
	r.met.BumpHistogram("bytes", float64(len(newVal)), tags...)
	_, err := r.connDo(context, "SET", key, newVal, "PX", int(expire/time.Millisecond))
	if err != nil {
		context.WithField("err", err).Error("set redis failed")
	}
	return err
}

func (r *redImpl) Set(context ctx.Ctx, key string, val []byte, expire time.Duration) error {
	return r.set(context, key, val, expire, false)
}

func (r *redImpl) SetZip(context ctx.Ctx, key string, val []byte, expire time.Duration) error {
	return r.set(context, key, val, expire, true)
}

func (r *redImpl) SetXX(context ctx.Ctx, key string, val []byte, expire time.Duration) error {
	tags := []string{"func", "setxx", "cluster", r.name, "prefix", keys.GetPrefix(key)}
	defer r.met.BumpTime("time", tags...).End()
	if expire == Forever {
		r.met.BumpSum("ttl.forever", 1, tags...)
	} else {
		r.met.BumpAvg("ttl", expire.Seconds(), tags...)
	}
	r.met.BumpHistogram("bytes", float64(len(val)), tags...)
	if expire == Forever {
		_, err := redis.Bytes(r.connDo(context, "SET", key, val, "XX"))

		if err != nil {
			context.WithField("err", err).Error("setXX redis failed")
		}
		return err
	}
	_, err := redis.Bytes(r.connDo(context, "SET", key, val, "PX", int(expire/time.Millisecond), "XX"))
	if err != nil {
		context.WithField("err", err).Error("setXX redis failed")
	}
	return err
}

func (r *redImpl) Expire(context ctx.Ctx, key string, ttl time.Duration) error {
	tags := []string{"func", "expire", "cluster", r.name, "prefix", keys.GetPrefix(key)}
	defer r.met.BumpTime("time", tags...).End()
	if ttl == Forever {
		r.met.BumpSum("ttl.forever", 1, tags...)
	} else {
		r.met.BumpAvg("ttl", ttl.Seconds(), tags...)
	}

	if ttl == Forever {
		_, err := r.connDo(context, "PERSIST", key)
		if err != nil {
			context.WithField("err", err).Error("Expire PERSIST redis key failed")
		}
		return err
	}

	reply, err := r.connDo(context, "EXPIRE", key, int(ttl/time.Second))
	if err != nil {
		context.WithField("err", err).Error("Expire redis failed")
		return err
	}
	// Return value will be 0 if key does not exist or the timeout could not be set.
	if reply.(int64) != 1 {
		return ErrExpireNotExistOrTimeout
	}
	return nil
}

func (r *redImpl) SetNX(context ctx.Ctx, key string, val []byte, expire time.Duration) error {
	tags := []string{"func", "setnx", "cluster", r.name, "prefix", keys.GetPrefix(key)}
	defer r.met.BumpTime("time", tags...).End()
	if expire == Forever {
		r.met.BumpSum("ttl.forever", 1, tags...)
	} else {
		r.met.BumpAvg("ttl", expire.Seconds(), tags...)
	}
	r.met.BumpHistogram("bytes", float64(len(val)), tags...)

	var err error
	if expire == Forever {
		_, err = redis.Bytes(r.connDo(context, "SET", key, val, "nx"))
	} else {
		_, err = redis.Bytes(r.connDo(context, "SET", key, val, "nx", "px", int(expire/time.Millisecond)))
	}

	return err
}

func (r *redImpl) ScanMatch(context ctx.Ctx, cursor int64, match string, count int) (int64, []string, error) {

	tags := []string{"func", "scanmatch", "cluster", r.name, "prefix", metrics.TagValueNA}
	defer r.met.BumpTime("time", tags...).End()

	var items []string

	if count < 1 {
		return cursor, items, fmt.Errorf("count cannot be less than 1")
	}

	values, err := redis.Values(r.connDo(context, "SCAN", cursor, "MATCH", match, "COUNT", count))
	if err != nil {
		return cursor, items, err
	}
	values, err = redis.Scan(values, &cursor, &items)
	if err != nil {
		return cursor, items, err
	}
	r.met.BumpHistogram("elements", float64(len(items)), tags...)
	return cursor, items, nil
}

func (r *redImpl) HSet(context ctx.Ctx, key, field string, val []byte, expire time.Duration) error {

	tags := []string{"func", "hset", "cluster", r.name, "prefix", keys.GetPrefix(key)}
	defer r.met.BumpTime("time", tags...).End()
	if expire == Forever {
		r.met.BumpSum("ttl.forever", 1, tags...)
	} else {
		r.met.BumpAvg("ttl", expire.Seconds(), tags...)
	}
	r.met.BumpHistogram("bytes", float64(len(val)), tags...)
	_, err := r.connDo(context, "HSET", key, field, val)
	if err != nil {
		context.WithField("err", err).Error("Hset redis failed")
		return err
	}
	if expire == Forever {
		_, err := r.connDo(context, "PERSIST", key)
		if err != nil {
			context.WithField("err", err).Error("HSet PERSIST redis key failed")
		}
		return err
	}
	_, err = r.connDo(context, "PEXPIRE", key, int(expire/time.Millisecond))
	if err != nil {
		context.WithField("err", err).Error("HSet PEXPIRE redis key failed")
	}
	return err
}

// HSetnx sets field in the hash stored at key to value, only if field does not yet exist.
// Return true if field not exists, false if field already exists and no operation was performed.
func (r *redImpl) HSetNX(context ctx.Ctx, key, field string, val []byte, expire time.Duration) (bool, error) {

	tags := []string{"func", "hsetnx", "cluster", r.name, "prefix", keys.GetPrefix(key)}
	defer r.met.BumpTime("time", tags...).End()
	if expire == Forever {
		r.met.BumpSum("ttl.forever", 1, tags...)
	} else {
		r.met.BumpAvg("ttl", expire.Seconds(), tags...)
	}
	r.met.BumpHistogram("bytes", float64(len(val)), tags...)
	ok, err := redis.Bool(r.connDo(context, "HSETNX", key, field, val))
	if err != nil {
		context.WithField("err", err).Error("HSetNX redis failed")
		return false, err
	}
	if !ok {
		return ok, nil
	}
	if expire == Forever {
		_, err := r.connDo(context, "PERSIST", key)
		if err != nil {
			context.WithField("err", err).Error("HSetNX PERSIST redis key failed")
		}
		return ok, err
	}
	_, err = r.connDo(context, "PEXPIRE", key, int(expire/time.Millisecond))
	if err != nil {
		context.WithField("err", err).Error("HSetNX PEXPIRE redis key failed")
	}
	return ok, err
}

func (r *redImpl) SetStruct(context ctx.Ctx, key string, val interface{}, expire time.Duration) error {

	tags := []string{"func", "setstruct", "cluster", r.name, "prefix", keys.GetPrefix(key)}
	defer r.met.BumpTime("time", tags...).End()
	if expire == Forever {
		r.met.BumpSum("ttl.forever", 1, tags...)
	} else {
		r.met.BumpAvg("ttl", expire.Seconds(), tags...)
	}
	fieldVal := map[string][]byte{}
	relectVal := reflect.ValueOf(val)
	if relectVal.Kind() == reflect.Ptr {
		relectVal = relectVal.Elem()
	}
	if relectVal.Kind() != reflect.Struct {
		return errors.New("only accept struct")
	}
	r.met.BumpHistogram("bytes", float64(binary.Size(val)), tags...)
	for i := 0; i < relectVal.NumField(); i++ {
		b, err := json.Marshal(relectVal.Field(i).Interface())
		if err != nil {
			context.WithFields(log.Fields{"err": err, "field": relectVal.Type().Field(i).Name}).Error("json Marshal fail")
			return err
		}
		fieldVal[relectVal.Type().Field(i).Name] = b
	}
	err := r.hmset(context, key, fieldVal, expire)
	return err
}

func (r *redImpl) GetStruct(context ctx.Ctx, key string, val interface{}) error {
	//
	tags := []string{"func", "getstruct", "cluster", r.name, "prefix", keys.GetPrefix(key)}
	defer r.met.BumpTime("time", tags...).End()
	redisHash, err := ByteMap(r.connDo(context, "HGETALL", key))
	if err != nil {
		context.WithField("err", err).Error("HGetAll failed")
		return err
	}
	if len(redisHash) == 0 {
		return ErrNotFound
	}
	reflectValue := reflect.ValueOf(val).Elem()
	for fieldName, value := range redisHash {
		field := reflectValue.FieldByName(fieldName)
		//   data in redis, but field is not exists in struct
		//   field will be zero value or invalid value
		if !field.IsValid() {
			continue
		}
		intf := reflect.New(field.Type()).Interface()
		if err := json.Unmarshal(value, &intf); err != nil {
			context.WithFields(log.Fields{"err": err, "fieldName": fieldName}).Error("json unmarshal failed")
			return err
		}
		if reflect.ValueOf(intf).IsValid() {
			field.Set(reflect.ValueOf(intf).Elem())
		}
	}
	r.met.BumpHistogram("bytes", float64(binary.Size(val)), tags...)
	return nil
}

// MSet sets  keys with  values, expire mean expire time for this key.
func (r *redImpl) MSet(context ctx.Ctx, keyVals map[string][]byte, expire time.Duration) error {

	tags := []string{"func", "mset", "cluster", r.name, "prefix", metrics.TagValueNA}
	defer r.met.BumpTime("time", tags...).End()
	if expire == Forever {
		r.met.BumpSum("ttl.forever", 1, tags...)
	} else {
		r.met.BumpAvg("ttl", expire.Seconds(), tags...)
	}
	r.met.BumpHistogram("elements", float64(len(keyVals)), tags...)
	size := 0
	for _, v := range keyVals {
		size += len(v)
	}

	r.met.BumpHistogram("bytes", float64(size), tags...)
	args := []interface{}{}
	for k, v := range keyVals {
		args = append(args, k, v)
	}

	conn, err := r.getConn("")
	if err != nil {
		return err
	}
	defer func() {
		if err := conn.Close(); err != nil {
			r.met.BumpSum("conn.Close.err", 1, "cluster", r.name)
		}
	}()

	err = conn.Send("MSet", redis.Args{}.AddFlat(args)...)
	if err != nil {
		context.WithField("err", err).Error("MSet redis failed")
		return err
	}

	if expire == Forever {
		for k := range keyVals {
			err = conn.Send("PERSIST", k)
			if err != nil {
				context.WithField("err", err).Error("MSet redis failed")
				return err
			}
		}
	} else {
		for k := range keyVals {
			err = conn.Send("PEXPIRE", k, int(expire/time.Millisecond))
			if err != nil {
				context.WithField("err", err).Error("MSet redis failed")
				return err
			}
		}
	}

	err = conn.Flush()
	if err != nil {
		context.WithField("err", err).Error("MSet redis failed")
		return err
	}

	return err
}

func (r *redImpl) HMSet(context ctx.Ctx, key string, fieldVal map[string][]byte, expire time.Duration) error {

	tags := []string{"func", "hmset", "cluster", r.name, "prefix", keys.GetPrefix(key)}
	defer r.met.BumpTime("time", tags...).End()
	if expire == Forever {
		r.met.BumpSum("ttl.forever", 1, tags...)
	} else {
		r.met.BumpAvg("ttl", expire.Seconds(), tags...)
	}
	r.met.BumpHistogram("elements", float64(len(fieldVal)), tags...)
	// add counters to record length of result to redis.HMSet.
	size := 0
	for _, v := range fieldVal {
		size += len(v)
	}
	r.met.BumpHistogram("bytes", float64(size), tags...)

	err := r.hmset(context, key, fieldVal, expire)
	return err
}

func (r *redImpl) hmset(context ctx.Ctx, key string, fieldVal map[string][]byte, expire time.Duration) error {
	_, err := r.connDo(context, "HMSET", redis.Args{}.Add(key).AddFlat(fieldVal)...)
	if err != nil {
		context.WithField("err", err).Error("HMSET redis failed")
		return err
	}
	if expire == Forever {
		_, err := r.connDo(context, "PERSIST", key)
		if err != nil {
			context.WithField("err", err).Error("HMSET PERSIST redis key failed")
		}
		return err
	}
	_, err = r.connDo(context, "PEXPIRE", key, int(expire/time.Millisecond))
	if err != nil {
		context.WithField("err", err).Error("HMSET PEXPIRE redis key failed")
	}
	return err
}

func (r *redImpl) HGet(context ctx.Ctx, key, field string) (val []byte, err error) {
	tags := []string{"func", "hget", "cluster", r.name, "prefix", keys.GetPrefix(key)}
	defer r.met.BumpTime("time", tags...).End()
	b, err := redis.Bytes(r.connDo(context, "HGET", key, field))
	r.met.BumpHistogram("bytes", float64(len(b)), tags...)
	return b, err
}

func (r *redImpl) HMGet(context ctx.Ctx, key string, fields ...string) ([]MVal, error) {

	tags := []string{"func", "hmget", "cluster", r.name, "prefix", keys.GetPrefix(key)}
	defer r.met.BumpTime("time", tags...).End()
	r.met.BumpHistogram("elements", float64(len(fields)), tags...)
	kvs := []interface{}{key}
	for _, field := range fields {
		kvs = append(kvs, field)
	}
	vals, err := redis.Values(r.connDo(context, "HMGET", kvs...))
	if err != nil {
		return nil, err
	}
	mvals := []MVal{}
	for _, val := range vals {
		if val == nil {
			mvals = append(mvals, MVal{
				Valid: false,
				Value: []byte(""),
			})
		} else {
			mvals = append(mvals, MVal{
				Valid: true,
				Value: val.([]byte),
			})
		}
	}
	return mvals, nil
}

func (r *redImpl) HGetAll(context ctx.Ctx, key string) (map[string][]byte, error) {

	tags := []string{"func", "hgetall", "cluster", r.name, "prefix", keys.GetPrefix(key)}
	defer r.met.BumpTime("time", tags...).End()
	vals, err := ByteMap(r.connDo(context, "HGETALL", key))
	// add counters to record length of result to redis.HGetAll.
	size := 0
	for _, v := range vals {
		size += len(v)
	}
	r.met.BumpHistogram("bytes", float64(size), tags...)

	if err != nil {
		context.WithField("err", err).Error("HGetAll redis failed")
	}
	// If map contains nothing, return err not found
	if len(vals) == 0 {
		return vals, ErrNotFound
	}
	return vals, err
}

func (r *redImpl) HDel(context ctx.Ctx, key, field string) (int, error) {

	defer r.met.BumpTime("time", "func", "hdel", "cluster", r.name, "prefix", keys.GetPrefix(key)).End()
	res, err := redis.Int(r.connDo(context, "HDEL", key, field))
	if err != nil {
		context.WithField("err", err).Error("HDel redis failed")
	}
	return res, err
}

func (r *redImpl) LTrim(context ctx.Ctx, key string, start, end int) error {

	defer r.met.BumpTime("time", "func", "ltrim", "cluster", r.name, "prefix", keys.GetPrefix(key)).End()
	_, err := r.connDo(context, "LTRIM", key, start, end)
	if err != nil {
		context.WithField("err", err).Error("LTrim redis failed")
	}
	return err
}

func (r *redImpl) HLen(context ctx.Ctx, key string) (length int, err error) {

	defer r.met.BumpTime("time", "func", "hlen", "cluster", r.name, "prefix", keys.GetPrefix(key)).End()
	val, err := redis.Int(r.connDo(context, "HLEN", key))
	if err != nil {
		context.WithField("err", err).Error("HLEN redis failed")
	}
	return val, err
}

// Incr Increments the number stored at key by one. If the key does not exist, it is set to 0 before performing the operation.
func (r *redImpl) Incr(context ctx.Ctx, key string) (int64, error) {

	defer r.met.BumpTime("time", "func", "incr", "cluster", r.name, "prefix", keys.GetPrefix(key)).End()
	res, err := redis.Int64(r.connDo(context, "INCR", key))
	if err != nil {
		context.WithField("err", err).Error("INCR redis failed")
	}
	return res, err
}

func (r *redImpl) Incrby(context ctx.Ctx, key string, val int) (int64, error) {

	defer r.met.BumpTime("time", "func", "incrby", "cluster", r.name, "prefix", keys.GetPrefix(key)).End()
	res, err := redis.Int64(r.connDo(context, "INCRBY", key, val))
	if err != nil {
		context.WithField("err", err).Error("INCRBY redis failed")
	}
	return res, err
}

func (r *redImpl) HIncrby(context ctx.Ctx, key, field string, val int) (int64, error) {

	defer r.met.BumpTime("time", "func", "hincrby", "cluster", r.name, "prefix", keys.GetPrefix(key)).End()
	res, err := redis.Int64(r.connDo(context, "HINCRBY", key, field, val))
	if err != nil {
		context.WithField("err", err).Error("INCRBY redis failed")
	}
	return res, err
}

func (r *redImpl) HScan(context ctx.Ctx, key string, cursor, count int) (map[string][]byte, int, error) {

	tags := []string{"func", "hscan", "cluster", r.name, "prefix", keys.GetPrefix(key)}
	defer r.met.BumpTime("time", tags...).End()
	values, err := redis.Values(r.connDo(context, "HSCAN", key, cursor, "count", count))
	if err != nil {
		context.WithField("err", err).Error("HScan redis failed")
		return nil, 0, err
	}
	s := uint8arrToString(values[0].([]uint8))
	cursorInt, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		context.WithField("err", err).Error("strconv.ParseInt failed")
		return nil, 0, err
	}
	m := make(map[string][]byte)
	var a [][]byte
	va := values[1].([]interface{})
	for _, v := range va {
		a = append(a, v.([]byte))
	}
	// add counters to record length of result to redis.HScan.
	size := 0
	for i := 0; i < len(a); i += 2 {
		m[string(a[i])] = a[i+1]
		size += len(a[i+1])
	}
	r.met.BumpHistogram("bytes", float64(size), tags...)

	return m, int(cursorInt), err
}

func (r *redImpl) ZAddXX(context ctx.Ctx, key string, memscore map[string]int) error {

	defer r.met.BumpTime("time", "func", "zaddxx", "cluster", r.name, "prefix", keys.GetPrefix(key)).End()
	_, err := r.connDo(context, "ZAdd", append([]interface{}{key, "XX"}, mapToSlice(memscore)...)...)
	if err != nil && err != ErrNotFound {
		context.WithField("err", err).Error("ZAddXX redis failed")
	}
	return err
}

func (r *redImpl) ZAddNXFloat(context ctx.Ctx, key string, memscore map[string]float64) error {

	defer r.met.BumpTime("time", "func", "zaddnx", "cluster", r.name, "prefix", keys.GetPrefix(key)).End()
	_, err := r.connDo(context, "ZAdd", append([]interface{}{key, "NX"}, mapToSliceFloat(memscore)...)...)
	if err != nil && err != ErrNotFound {
		context.WithField("err", err).Error("ZAddNX redis failed")
	}
	return err
}

func (r *redImpl) ZAdd(context ctx.Ctx, key string, memscore map[string]int) error {

	defer r.met.BumpTime("time", "func", "zadd", "cluster", r.name, "prefix", keys.GetPrefix(key)).End()
	_, err := r.connDo(context, "ZAdd", append([]interface{}{key}, mapToSlice(memscore)...)...)
	if err != nil && err != ErrNotFound {
		context.WithField("err", err).Error("ZAdd redis failed")
	}
	return err
}

func mapToSlice(m map[string]int) []interface{} {
	s := []interface{}{}
	for k, v := range m {
		s = append(s, v, k)
	}
	return s
}

func (r *redImpl) ZAddFloat(context ctx.Ctx, key string, memSco map[string]float64) error {

	defer r.met.BumpTime("time", "func", "zaddfloat", "cluster", r.name, "prefix", keys.GetPrefix(key)).End()
	_, err := r.connDo(context, "ZAdd", append([]interface{}{key}, mapToSliceFloat(memSco)...)...)
	if err != nil && err != ErrNotFound {
		context.WithField("err", err).Error("ZAdd redis failed")
	}
	return err
}

func mapToSliceFloat(m map[string]float64) []interface{} {
	s := []interface{}{}
	for k, v := range m {
		s = append(s, v, k)
	}
	return s
}

func (r *redImpl) ZScore(context ctx.Ctx, key, member string) (int, error) {

	defer r.met.BumpTime("time", "func", "zscore", "cluster", r.name, "prefix", keys.GetPrefix(key)).End()
	val, err := redis.Int(r.connDo(context, "ZScore", key, member))
	if err != nil && err != ErrNotFound {
		context.WithField("err", err).Error("ZScore redis failed")
		return 0, err
	}
	return val, err
}

func (r *redImpl) ZScoreFloat(context ctx.Ctx, key, member string) (float64, error) {

	defer r.met.BumpTime("time", "func", "zscorefloat", "cluster", r.name, "prefix", keys.GetPrefix(key)).End()
	val, err := redis.Float64(r.connDo(context, "ZScore", key, member))
	if err != nil && err != ErrNotFound {
		context.WithField("err", err).Error("ZScoreFloat redis failed")
		return 0, err
	}
	return val, err
}

func (r *redImpl) ZIncrby(context ctx.Ctx, key string, member string, val int) (int, error) {

	defer r.met.BumpTime("time", "func", "zincrby", "cluster", r.name, "prefix", keys.GetPrefix(key)).End()
	v, err := redis.Int(r.connDo(context, "ZIncrby", key, val, member))
	if err != nil && err != ErrNotFound {
		context.WithField("err", err).Error("ZIncrby redis failed")
	}
	return v, err
}

func (r *redImpl) ZIncrbyFloat(context ctx.Ctx, key string, member string, val float64) (float64, error) {

	defer r.met.BumpTime("time", "func", "zincrbyfloat", "cluster", r.name, "prefix", keys.GetPrefix(key)).End()
	v, err := redis.Float64(r.connDo(context, "ZIncrby", key, val, member))
	if err != nil && err != ErrNotFound {
		context.WithField("err", err).Error("ZIncrbyFloat redis failed")
	}
	return v, err
}

func (r *redImpl) SScan(context ctx.Ctx, key string, cursor, count int) ([]string, int, error) {

	defer r.met.BumpTime("time", "func", "sscan", "cluster", r.name, "prefix", keys.GetPrefix(key)).End()
	val, err := redis.Values(r.connDo(context, "SScan", key, cursor, "count", count))
	if err != nil {
		context.WithField("err", err).Error("SScan redis failed")
		return nil, 0, err
	}
	if len(val) < 2 {
		context.WithFields(log.Fields{
			"val":    val,
			"length": len(val),
		}).Error("SScan return length less than 2")
		return nil, 0, fmt.Errorf("SScan return length less than 2")
	}

	s := uint8arrToString(val[0].([]uint8))
	cursorInt, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		context.WithField("err", err).Error("strconv.ParseInt failed")
		return nil, 0, err
	}

	res := []string{}
	va := val[1].([]interface{})
	for _, v := range va {
		res = append(res, string(v.([]byte)))
	}
	return res, int(cursorInt), err
}

func (r *redImpl) ZScan(context ctx.Ctx, key string, cursor, limit int) (map[string]int, int, error) {

	defer r.met.BumpTime("time", "func", "zscan", "cluster", r.name, "prefix", keys.GetPrefix(key)).End()
	val, err := redis.Values(r.connDo(context, "ZScan", key, cursor, "count", limit))
	if err != nil && err != ErrNotFound {
		context.WithField("err", err).Error("ZScan redis failed")
		return nil, 0, err
	}
	if len(val) < 2 {
		return nil, 0, fmt.Errorf("ZScan return length less than 2")
	}

	s := uint8arrToString(val[0].([]uint8))
	cursorInt, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		context.WithField("err", err).Error("strconv.ParseInt failed")
		return nil, 0, err
	}

	m := make(map[string]int)
	var a []string
	va := val[1].([]interface{})
	for _, v := range va {
		a = append(a, string(v.([]byte)))
	}
	for i := 0; i < len(a); i += 2 {
		aint, err := strconv.ParseInt(a[i+1], 10, 64)
		if err != nil {
			return nil, 0, err
		}
		m[a[i]] = int(aint)
	}
	return m, int(cursorInt), err
}

func (r *redImpl) ZCard(context ctx.Ctx, key string) (count int, err error) {

	defer r.met.BumpTime("time", "func", "zcard", "cluster", r.name, "prefix", keys.GetPrefix(key)).End()
	val, err := redis.Int(r.connDo(context, "ZCard", key))
	if err != nil && err != ErrNotFound {
		context.WithField("err", err).Error("ZCard redis failed")
	}
	return val, err

}

func (r *redImpl) ZRevrange(context ctx.Ctx, key string, offset, count int) ([]string, error) {

	tags := []string{"func", "zrevrange", "cluster", r.name, "prefix", keys.GetPrefix(key)}
	defer r.met.BumpTime("time", tags...).End()
	val, err := redis.Strings(r.connDo(context, "ZREVRANGE", key, offset, count-1+offset))
	if err != nil && err != ErrNotFound {
		context.WithField("err", err).Error("ZRevrange redis failed")
	}
	r.met.BumpHistogram("elements", float64(len(val)), tags...)
	return val, err
}

func (r *redImpl) ZRange(context ctx.Ctx, key string, offset, count int) ([]string, error) {

	tags := []string{"func", "zrange", "cluster", r.name, "prefix", keys.GetPrefix(key)}
	defer r.met.BumpTime("time", tags...).End()
	val, err := redis.Strings(r.connDo(context, "ZRANGE", key, offset, count-1+offset))
	if err != nil && err != ErrNotFound {
		context.WithField("err", err).Error("Zrange redis failed")
	}
	r.met.BumpHistogram("elements", float64(len(val)), tags...)
	return val, err
}

func (r *redImpl) ZCount(context ctx.Ctx, key, minScore, maxScore string) (int, error) {

	tags := []string{"func", "zcount", "cluster", r.name, "prefix", keys.GetPrefix(key)}
	defer r.met.BumpTime("time", tags...).End()
	val, err := redis.Int(r.connDo(context, "ZCOUNT", key, minScore, maxScore))
	if err != nil && err != ErrNotFound {
		context.WithField("err", err).Error("ZCount failed")
	}
	return val, err
}

func (r *redImpl) ZRangeByScoreWithScore(context ctx.Ctx, key, minScore, maxScore string) ([]ZVal, error) {

	tags := []string{"func", "zrangebyscorewithscore", "cluster", r.name, "prefix", keys.GetPrefix(key)}
	defer r.met.BumpTime("time", tags...).End()
	results, err := redis.Values(r.connDo(context, "ZRANGEBYSCORE", key, minScore, maxScore, "withscores"))
	if err != nil && err != ErrNotFound {
		context.WithField("err", err).Error("ZRangeByScoreWithScore failed")
	}
	if len(results)%2 != 0 {
		err = errors.New("ZRANGEBYSCORE: number of results not a multiple of 2")
		context.WithFields(log.Fields{
			"err":      err,
			"key":      key,
			"minScore": minScore,
			"maxScore": maxScore,
		}).Error("ZRANGEBYSCORE size of values error")
		return nil, err
	}
	vals := make([]ZVal, len(results)/2)
	err = redis.ScanSlice(results, &vals)
	if err != nil {
		context.WithField("err", err).Error("ZRANGEBYSCORE pack values failed")
		return nil, err
	}
	r.met.BumpHistogram("elements", float64(len(vals)), tags...)
	return vals, err
}

func (r *redImpl) ZRevrangeScore(context ctx.Ctx, key string, offset, count int) ([]ZVal, error) {

	tags := []string{"func", "zrevrangescore", "cluster", r.name, "prefix", keys.GetPrefix(key)}
	defer r.met.BumpTime("time", tags...).End()
	results, err := redis.Values(r.connDo(context, "ZREVRANGE", key, offset, count-1+offset, "withscores"))
	if err != nil && err != ErrNotFound {
		context.WithField("err", err).Error("ZRevrangeScore redis failed")
		return nil, err
	}
	if len(results)%2 != 0 {
		err = errors.New("ZRevrangeScore number of results not a multiple of 2")
		context.WithFields(log.Fields{
			"err":    err,
			"key":    key,
			"offset": offset,
			"count":  count,
		}).Error("ZRevrangeScore size of values error")
		return nil, err
	}
	vals := make([]ZVal, len(results)/2)

	// ScanSlice(src, dst, fieldNames...) scans src array and packs them into new structs in dst array.
	// For example, packing [v11, v12, v21, v22, v31, v32] into struct{k1, k2}
	// returns [
	// 	 struct{k1: v11, k2: v12},
	//	 struct{k1: v21, k2: v22},
	//	 struct{k1: v31, k2: v32},
	// ]
	// By default all fields in given struct will be processed sequentially.
	// You can specify field names if src array contains only subset of fields
	// Ref: https://godoc.org/github.com/gomodule/redigo/redis#example-ScanSlice
	err = redis.ScanSlice(results, &vals)
	if err != nil {
		context.WithField("err", err).Error("ZRevrangeScore pack values failed")
		return nil, err
	}
	r.met.BumpHistogram("elements", float64(len(vals)), tags...)
	return vals, err
}

func (r *redImpl) ZRevrangeFloatScore(context ctx.Ctx, key string, offset, count int) ([]ZFloatVal, error) {

	tags := []string{"func", "zrevrangefloatscore", "cluster", r.name, "prefix", keys.GetPrefix(key)}
	defer r.met.BumpTime("time", tags...).End()
	results, err := redis.Values(r.connDo(context, "ZREVRANGE", key, offset, count-1+offset, "withscores"))
	if err != nil && err != ErrNotFound {
		context.WithField("err", err).Error("ZRevrangeFloatScore redis failed")
		return nil, err
	}
	if len(results)%2 != 0 {
		err = errors.New("ZRevrangeFloatScore number of results not a multiple of 2")
		context.WithFields(log.Fields{
			"err":    err,
			"key":    key,
			"offset": offset,
			"count":  count,
		}).Error("ZRevrangeFloatScore size of values error")
		return nil, err
	}
	vals := make([]ZFloatVal, len(results)/2)

	// ScanSlice(src, dst, fieldNames...) scans src array and packs them into new structs in dst array.
	// For example, packing [v11, v12, v21, v22, v31, v32] into struct{k1, k2}
	// returns [
	// 	 struct{k1: v11, k2: v12},
	//	 struct{k1: v21, k2: v22},
	//	 struct{k1: v31, k2: v32},
	// ]
	// By default all fields in given struct will be processed sequentially.
	// You can specify field names if src array contains only subset of fields
	// Ref: https://godoc.org/github.com/gomodule/redigo/redis#example-ScanSlice
	err = redis.ScanSlice(results, &vals)
	if err != nil {
		context.WithField("err", err).Error("ZRevrangeFloatScore pack values failed")
		return nil, err
	}
	r.met.BumpHistogram("elements", float64(len(vals)), tags...)
	return vals, err
}

func (r *redImpl) ZRevrangeByScoreWithScore(context ctx.Ctx, key, minScore, maxScore string) ([]ZVal, error) {

	tags := []string{"func", "zrevrangebyscorewithscore", "cluster", r.name, "prefix", keys.GetPrefix(key)}
	defer r.met.BumpTime("time", tags...).End()
	results, err := redis.Values(r.connDo(context, "ZREVRANGEBYSCORE", key, maxScore, minScore, "withscores"))
	if err != nil && err != ErrNotFound {
		context.WithField("err", err).Error("ZrevrangebyScorewithscore redis failed")
	}
	if len(results)%2 != 0 {
		err = errors.New("ZREVRANGEBYSCORE: number of results not a multiple of 2")
		context.WithFields(log.Fields{
			"err":      err,
			"key":      key,
			"minScore": minScore,
			"maxScore": maxScore,
		}).Error("ZREVRANGEBYSCORE size of values error")
		return nil, err
	}
	vals := make([]ZVal, len(results)/2)
	err = redis.ScanSlice(results, &vals)
	if err != nil {
		context.WithField("err", err).Error("ZREVRANGEBYSCORE pack values failed")
		return nil, err
	}
	r.met.BumpHistogram("elements", float64(len(vals)), tags...)
	return vals, err
}

func (r *redImpl) ZRevrangeByScoreWithFloatScore(context ctx.Ctx, key, minScore, maxScore string) ([]ZFloatVal, error) {

	tags := []string{"func", "zrevrangebyscorewithfloatscore", "cluster", r.name, "prefix", keys.GetPrefix(key)}
	defer r.met.BumpTime("time", tags...).End()
	results, err := redis.Values(r.connDo(context, "ZREVRANGEBYSCORE", key, maxScore, minScore, "withscores"))
	if err != nil && err != ErrNotFound {
		context.WithField("err", err).Error("ZRevrangeByScoreWithFloatScore redis failed")
	}
	if len(results)%2 != 0 {
		err = errors.New("ZRevrangeByScoreWithFloatScore: number of results not a multiple of 2")
		context.WithFields(log.Fields{
			"err":      err,
			"key":      key,
			"minScore": minScore,
			"maxScore": maxScore,
		}).Error("ZRevrangeByScoreWithFloatScore size of values error")
		return nil, err
	}
	vals := make([]ZFloatVal, len(results)/2)
	err = redis.ScanSlice(results, &vals)
	if err != nil {
		context.WithField("err", err).Error("ZREVRANGEBYSCORE pack values failed")
		return nil, err
	}
	r.met.BumpHistogram("elements", float64(len(vals)), tags...)
	return vals, err
}

func (r *redImpl) ZRem(context ctx.Ctx, key string, members ...string) error {

	defer r.met.BumpTime("time", "func", "zrem", "cluster", r.name, "prefix", keys.GetPrefix(key)).End()

	vars := []interface{}{key}
	for _, j := range members {
		vars = append(vars, interface{}(j))
	}

	_, err := r.connDo(context, "ZRem", vars...)
	if err != nil && err != ErrNotFound {
		context.WithField("err", err).Error("ZRem redis failed")
		return err
	}
	return err
}

func (r *redImpl) ZRemRangeByScore(context ctx.Ctx, key string, minScore, maxScore int) (int, error) {

	defer r.met.BumpTime("time", "func", "zremrangebyscore", "cluster", r.name, "prefix", keys.GetPrefix(key)).End()

	count, err := redis.Int(r.connDo(context, "ZREMRANGEBYSCORE", key, minScore, maxScore))
	if err != nil && err != ErrNotFound {
		context.WithField("err", err).Error("ZRemRangeByScore redis failed")
		return 0, err
	}

	return count, err
}

func (r *redImpl) ZRemRangeByRank(context ctx.Ctx, key string, start, stop int) (int, error) {

	defer r.met.BumpTime("time", "func", "zremrangebyrank", "cluster", r.name, "prefix", keys.GetPrefix(key)).End()

	count, err := redis.Int(r.connDo(context, "ZREMRANGEBYRANK", key, start, stop))
	if err != nil && err != ErrNotFound {
		context.WithField("err", err).Error("ZRemRangeByRank redis failed")
		return 0, err
	}

	return count, err
}

func (r *redImpl) ZRevRank(context ctx.Ctx, key, member string) (int, error) {

	defer r.met.BumpTime("time", "func", "zrevrank", "cluster", r.name, "prefix", keys.GetPrefix(key)).End()

	val, err := redis.Int(r.connDo(context, "ZREVRANK", key, member))
	if err != nil && err != ErrNotFound {
		context.WithField("err", err).Error("ZREVRANK redis failed")
		return 0, err
	}
	return val, err
}

func (r *redImpl) ZPopMin(context ctx.Ctx, key string, count int) ([]ZFloatVal, error) {

	tags := []string{"func", "zpopmin", "cluster", r.name, "prefix", keys.GetPrefix(key)}
	defer r.met.BumpTime("time", tags...).End()
	results, err := redis.Values(r.connDo(context, "ZPOPMIN", key, count))
	if err != nil && err != ErrNotFound {
		context.WithField("err", err).Error("ZPopMin failed")
		return nil, err
	}
	if len(results)%2 != 0 {
		err = errors.New("ZPOPMIN: number of results not a multiple of 2")
		context.WithFields(log.Fields{
			"err":   err,
			"key":   key,
			"count": count,
		}).Error("ZPOPMIN size of values error")
		return nil, err
	}
	vals := make([]ZFloatVal, len(results)/2)
	if err := redis.ScanSlice(results, &vals); err != nil {
		context.WithField("err", err).Error("ZPOPMIN pack values failed")
		return nil, err
	}
	r.met.BumpHistogram("elements", float64(len(vals)), tags...)
	return vals, nil
}

func (r *redImpl) ZUnionStore(context ctx.Ctx, paris []Pair, dest string) error {

	if len(paris) <= 1 {
		return fmt.Errorf("Invalid paris input")
	}

	defer r.met.BumpTime("time", "func", "zunionstore", "cluster", r.name, "prefix", keys.GetPrefix(dest)).End()

	vars := []interface{}{dest}
	vars = append(vars, interface{}(len(paris)))
	for _, p := range paris {
		vars = append(vars, interface{}(p.Key))
	}
	vars = append(vars, interface{}("WEIGHTS"))
	for _, p := range paris {
		var w float64 = 1
		if p.Weight != 0 {
			w = p.Weight
		}
		vars = append(vars, interface{}(w))
	}

	_, err := r.connDo(context, "ZUNIONSTORE", vars...)
	if err != nil && err != ErrNotFound {
		context.WithField("err", err).Error("ZUnionStore redis failed")
		return err
	}
	return err
}

func (r *redImpl) LPush(context ctx.Ctx, key string, val []byte) error {

	defer r.met.BumpTime("time", "func", "LPush", "cluster", r.name, "prefix", metrics.TagValueNA).End()
	r.met.BumpHistogram("bytes", float64(len(val)), "func", "LPush", "cluster", r.name, "prefix", keys.GetPrefix(key))

	if _, err := r.connDo(context, "LPUSH", key, val); err != nil {
		context.WithField("err", err).Error("LPush redis failed")
		return err
	}
	return nil
}

func (r *redImpl) LRange(context ctx.Ctx, key string, offset, count int) (val [][]byte, err error) {

	tags := []string{"func", "LRANGE", "cluster", r.name, "prefix", keys.GetPrefix(key)}
	defer r.met.BumpTime("time", tags...).End()
	val, err = redis.ByteSlices(r.connDo(context, "LRANGE", key, offset, count-1+offset))
	if err != nil {
		context.WithField("err", err).Error("LRANGE redis failed")
		return nil, err
	}
	r.met.BumpHistogram("elements", float64(len(val)), tags...)
	return val, nil
}

func (r *redImpl) RPush(context ctx.Ctx, key string, val []byte) (int, error) {

	defer r.met.BumpTime("time", "func", "RPush", "cluster", r.name, "prefix", metrics.TagValueNA).End()
	r.met.BumpHistogram("bytes", float64(len(val)), "func", "RPush", "cluster", r.name, "prefix", keys.GetPrefix(key))

	listSize, err := redis.Int(r.connDo(context, "RPUSH", key, val))
	if err != nil {
		context.WithField("err", err).Error("RPush redis failed")
		return 0, err
	}
	return listSize, nil
}

func (r *redImpl) LPop(context ctx.Ctx, key string) ([]byte, error) {

	defer r.met.BumpTime("time", "func", "LPop", "cluster", r.name, "prefix", metrics.TagValueNA).End()

	val, err := redis.Bytes(r.connDo(context, "LPOP", key))
	if err == redis.ErrNil {
		return nil, ErrNotFound
	}
	r.met.BumpHistogram("bytes", float64(len(val)), "func", "LPop", "cluster", r.name, "prefix", keys.GetPrefix(key))
	return val, err
}

func (r *redImpl) RPop(context ctx.Ctx, key string) ([]byte, error) {

	defer r.met.BumpTime("time", "func", "RPop", "cluster", r.name, "prefix", metrics.TagValueNA).End()

	val, err := redis.Bytes(r.connDo(context, "RPOP", key))
	r.met.BumpHistogram("bytes", float64(len(val)), "func", "RPop", "cluster", r.name, "prefix", keys.GetPrefix(key))
	if err == redis.ErrNil {
		return nil, ErrNotFound
	}
	return val, err
}

func (r *redImpl) RandomKey(context ctx.Ctx) ([]byte, error) {

	defer r.met.BumpTime("time", "func", "RandomKey", "cluster", r.name).End()

	val, err := redis.Bytes(r.connDo(context, "RANDOMKEY"))
	if err != nil {
		context.WithField("err", err).Error("RandomKey redis failed")
		return nil, err
	}
	return val, nil
}

func (r *redImpl) Type(context ctx.Ctx, key string) ([]byte, error) {

	defer r.met.BumpTime("time", "func", "Type", "cluster", r.name).End()

	val, err := redis.Bytes(r.connDo(context, "Type", key))
	if err != nil {
		context.WithField("err", err).Error("Type redis failed")
		return nil, err
	}
	return val, nil
}

func (r *redImpl) Strlen(context ctx.Ctx, key string) (int, error) {

	defer r.met.BumpTime("time", "func", "Strlen", "cluster", r.name).End()

	val, err := redis.Int(r.connDo(context, "STRLEN", key))
	if err != nil {
		context.WithField("err", err).Error("Strlen redis failed")
		return 0, err
	}
	return val, nil
}

func (r *redImpl) LSet(context ctx.Ctx, key string, index int, val []byte) error {

	defer r.met.BumpTime("time", "func", "LSet", "cluster", r.name, "prefix", metrics.TagValueNA).End()
	r.met.BumpHistogram("bytes", float64(len(val)), "func", "LSet", "cluster", r.name, "prefix", keys.GetPrefix(key))

	if _, err := r.connDo(context, "LSET", key, index, val); err != nil {
		context.WithField("err", err).Error("LSet redis failed")
		return err
	}
	return nil
}

func (r *redImpl) LIndex(context ctx.Ctx, key string, index int64) ([]byte, error) {

	defer r.met.BumpTime("time", "func", "LIndex", "cluster", r.name, "prefix", metrics.TagValueNA).End()

	val, err := redis.Bytes(r.connDo(context, "LIndex", key, index))
	r.met.BumpHistogram("bytes", float64(len(val)), "func", "LIndex", "cluster", r.name, "prefix", keys.GetPrefix(key))

	if err != nil {
		//context.WithField("err", err).Error("LIndex redis failed")
		return nil, err
	}
	return val, nil
}

func (r *redImpl) LLen(context ctx.Ctx, key string) (length int, err error) {

	defer r.met.BumpTime("time", "func", "LLen", "cluster", r.name, "prefix", metrics.TagValueNA).End()

	val, err := redis.Int(r.connDo(context, "LLEN", key))
	if err != nil {
		context.WithField("err", err).Error("LLen redis failed")
	}
	return val, err
}

func (r *redImpl) LInsert(context ctx.Ctx, key string, before, val []byte) error {

	defer r.met.BumpTime("time", "func", "LINSERT", "cluster", r.name, "prefix", metrics.TagValueNA).End()
	r.met.BumpHistogram("bytes", float64(len(val)), "func", "LINSERT", "cluster", r.name, "prefix", keys.GetPrefix(key))

	if _, err := r.connDo(context, "LINSERT", key, "BEFORE", before, val); err != nil {
		context.WithField("err", err).Error("LINSERT redis failed")
		return err
	}
	return nil
}

func (r *redImpl) SAdd(context ctx.Ctx, key string, member ...string) error {

	_, err := r.SAddFullInfo(context, key, member...)
	return err
}

func (r *redImpl) SAddFullInfo(context ctx.Ctx, key string, member ...string) (int64, error) {

	if len(member) == 0 {
		return 0, fmt.Errorf("length of member is 0")
	}

	tags := []string{"func", "sadd", "cluster", r.name, "prefix", keys.GetPrefix(key)}
	defer r.met.BumpTime("time", tags...).End()

	vars := []interface{}{key}
	size := 0
	for _, j := range member {
		vars = append(vars, interface{}(j))
		size += len([]byte(j))
	}

	r.met.BumpHistogram("bytes", float64(size), tags...)

	val, err := r.connDo(context, "SADD", vars...)
	if err != nil {
		context.WithField("err", err).Error("SAdd redis failed")
		return 0, err
	}

	return val.(int64), nil
}

func (r *redImpl) SRem(context ctx.Ctx, key string, member ...string) error {

	if len(member) == 0 {
		return fmt.Errorf("length of member is 0")
	}

	tags := []string{"func", "srem", "cluster", r.name, "prefix", keys.GetPrefix(key)}
	defer r.met.BumpTime("time", tags...).End()

	vars := []interface{}{key}
	size := 0
	for _, j := range member {
		vars = append(vars, interface{}(j))
		size += len([]byte(j))
	}

	r.met.BumpHistogram("bytes", float64(size), tags...)

	if _, err := r.connDo(context, "SREM", vars...); err != nil {
		context.WithField("err", err).Error("SRem redis failed")
		return err
	}
	return nil
}

func (r *redImpl) SMembers(context ctx.Ctx, key string) ([]string, error) {

	tags := []string{"func", "smembers", "cluster", r.name, "prefix", keys.GetPrefix(key)}
	defer r.met.BumpTime("time", tags...).End()
	// add counters to record length of result form redis.SMembers.
	dest, err := redis.Strings(r.connDo(context, "SMEMBERS", key))
	size := 0
	for _, s := range dest {
		size += len([]byte(s))
	}
	r.met.BumpHistogram("bytes", float64(size), tags...)

	r.met.BumpHistogram("smember.return", float64(len(dest)), tags...)
	return dest, err
}

// SIsMember Returns if member is a member of the set stored at key.
func (r *redImpl) SIsMember(context ctx.Ctx, key, member string) (bool, error) {

	defer r.met.BumpTime("time", "func", "sismember", "cluster", r.name, "prefix", keys.GetPrefix(key)).End()
	res, err := redis.Bool(r.connDo(context, "SISMEMBER", key, member))
	if err != nil {
		context.WithField("err", err).Error("SIsMember redis failed")
	}
	return res, err
}

// SCard Returns the set cardinality (number of elements) of the set stored at key.
func (r *redImpl) SCard(context ctx.Ctx, key string) (int, error) {

	defer r.met.BumpTime("time", "func", "scard", "cluster", r.name, "prefix", keys.GetPrefix(key)).End()
	res, err := redis.Int(r.connDo(context, "SCARD", key))
	if err != nil {
		context.WithField("err", err).Error("SCard redis failed")
	}
	return res, err
}

func (r *redImpl) Rename(context ctx.Ctx, oldKey, newKey string) error {

	defer r.met.BumpTime("time", "func", "rename", "cluster", r.name, "prefix", keys.GetPrefix(oldKey)).End()
	_, err := r.connDo(context, "RENAME", oldKey, newKey)
	if err != nil {
		context.WithFields(log.Fields{
			"err":    err,
			"oldKey": oldKey,
			"newKey": newKey,
		}).Error("Rename failed")
	}
	return err
}

// Exists Returns if the key exists.
func (r *redImpl) Exists(context ctx.Ctx, key string) (bool, error) {

	defer r.met.BumpTime("time", "func", "exists", "cluster", r.name, "prefix", keys.GetPrefix(key)).End()
	res, err := redis.Bool(r.connDo(context, "Exists", key))
	if err != nil {
		context.WithField("err", err).Error("Exists redis failed")
	}
	return res, err
}

func (r *redImpl) TTL(context ctx.Ctx, key string) (int, error) {

	defer r.met.BumpTime("time", "func", "TTL", "cluster", r.name, "prefix", keys.GetPrefix(key)).End()
	res, err := redis.Int(r.connDo(context, "TTL", key))
	if err != nil {
		context.WithField("err", err).Error("TTL redis failed")
		return 0, err
	}

	if res == retTTLNoKey {
		return res, ErrNotFound
	} else if res == retTTLNoExpire {
		return res, ErrNoTTL
	}
	return res, nil
}

func (r *redImpl) SPop(context ctx.Ctx, key string) (string, error) {
	defer r.met.BumpTime("time", "func", "spop", "cluster", r.name, "prefix", keys.GetPrefix(key)).End()
	dest, err := redis.String(r.connDo(context, "SPOP", key))

	return dest, err
}

func (r *redImpl) SMPop(context ctx.Ctx, key string, count int) ([]string, error) {
	defer r.met.BumpTime("time", "func", "smpop", "cluster", r.name, "prefix", keys.GetPrefix(key)).End()
	return redis.Strings(r.connDo(context, "SPOP", key, count))
}

func (r *redImpl) PFAdd(context ctx.Ctx, key string, members ...string) (int, error) {

	if len(members) == 0 {
		return 0, fmt.Errorf("length of member is 0")
	}

	tags := []string{"func", "pfadd", "cluster", r.name, "prefix", keys.GetPrefix(key)}
	defer r.met.BumpTime("time", tags...).End()

	vars := []interface{}{key}
	size := 0
	for _, j := range members {
		vars = append(vars, interface{}(j))
		size += len([]byte(j))
	}

	r.met.BumpHistogram("bytes", float64(size), tags...)

	val, err := redis.Int(r.connDo(context, "PFADD", vars...))
	if err != nil {
		context.WithField("err", err).Error("PFAdd redis failed")
		return 0, err
	}
	return val, nil
}

func (r *redImpl) ScriptDo(context ctx.Ctx, hdl *ScriptHdl, keysAndArgs ...interface{}) (interface{}, error) {

	defer r.met.BumpTime("time", "func", "scriptdo", "cluster", r.name, "prefix", hdl.prefix(keysAndArgs...)).End()

	conn, err := r.getConn("")
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := conn.Close(); err != nil {
			r.met.BumpSum("conn.Close.err", 1, "cluster", r.name)
		}
	}()

	value, err := hdl.Do(conn, keysAndArgs...)
	if err != nil && err != ErrNotFound {
		context.WithField("err", err).Error("ScriptDo redis failed")
	}

	return value, err
}

func (r *redImpl) GetConn() (redis.Conn, error) {
	return r.getConn("")
}

func (r *redImpl) Name() string {
	return r.name
}

func uint8arrToString(bs []uint8) string {
	b := make([]byte, len(bs))
	for i, v := range bs {
		b[i] = byte(v)
	}
	return string(b)
}

// ByteMap is a helper that converts an array of []byte (alternating key, value)
// into a map[string][]byte. The HGETALL and CONFIG GET commands return replies in this format.
// Requires an even number of values in result.
func ByteMap(result interface{}, err error) (map[string][]byte, error) {
	values, err := redis.Values(result, err)
	if err != nil {
		return nil, err
	}
	if len(values)%2 != 0 {
		return nil, errors.New("redigo: ByteMap expects even number of values result")
	}
	m := make(map[string][]byte, len(values)/2)
	for i := 0; i < len(values); i += 2 {
		key, okKey := values[i].([]byte)
		value, okValue := values[i+1].([]byte)
		if !okKey || !okValue {
			return nil, errors.New("redigo: ByteMap key not a bulk string value")
		}
		m[string(key)] = value
	}
	return m, nil
}
