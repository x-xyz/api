package paging

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/viney-shih/goroutines"
	"github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/base/log"
	"github.com/x-xyz/goapi/base/math"
	"github.com/x-xyz/goapi/domain/keys"
	"github.com/x-xyz/goapi/service/redis"
)

var (
	errCursorLengthNotCorrect = errors.New("cursor length not correct")

	defaultShardSize = 100
)

type impl struct {
	keyPfx        string
	getter        Getter
	redisCache    redis.Service
	renewDuration time.Duration
	cacheDuration time.Duration

	getterTimeout time.Duration
	shardSize     int

	latestKey string
	lockKey   string

	workerPool *goroutines.Pool
}

func New(p *PagingConfig) Service {
	if p.RedisCache == nil {
		panic("RedisCache can not be nil")
	}

	if p.RenewDuration == 0 {
		panic("RenewDuration can not be 0")
	}

	if p.CacheDuration == 0 {
		panic("CacheDuration can not be 0")
	}

	if p.GetterTimeout == 0 {
		p.GetterTimeout = 10 * time.Second
	}

	if p.ShardSize == 0 {
		p.ShardSize = defaultShardSize
	}

	latestKey := keys.RedisKey(keys.PfxPagingService, "la", p.KeyPfx)
	lockKey := keys.RedisKey(keys.PfxPagingService, "lock", p.KeyPfx)

	return &impl{
		keyPfx:        p.KeyPfx,
		getter:        p.Getter,
		redisCache:    p.RedisCache,
		renewDuration: p.RenewDuration,
		cacheDuration: p.CacheDuration,
		getterTimeout: p.GetterTimeout,
		shardSize:     p.ShardSize,
		latestKey:     latestKey,
		lockKey:       lockKey,
		workerPool:    goroutines.NewPool(256),
	}
}

type cursorStruct struct {
	createTs   int64
	totalCount int
	offset     int
}

func encodeCursor(c cursorStruct) string {
	rawString := fmt.Sprintf("%v:%v:%v", c.createTs, c.totalCount, c.offset)
	return base64.StdEncoding.EncodeToString([]byte(rawString))
}

func decodeCursor(s string) (*cursorStruct, error) {
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, err
	}

	rawString := string(data)
	parts := strings.Split(rawString, ":")
	if len(parts) != 3 {
		return nil, errCursorLengthNotCorrect
	}

	createTs, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return nil, err
	}

	totalCount, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, err
	}

	offset, err := strconv.Atoi(parts[2])
	if err != nil {
		return nil, err
	}

	return &cursorStruct{
		createTs:   createTs,
		totalCount: totalCount,
		offset:     offset,
	}, nil
}

func getCoveredShard(shardSize, offset, size int) (fromShard, toShard int) {
	if size == 0 {
		return 0, 0
	}

	fromShard = offset / shardSize
	toShard = (offset+size-1)/shardSize + 1
	return fromShard, toShard
}

func (im *impl) getRawData(ctx ctx.Ctx, key string, cursor cursorStruct, size int) ([]byte, error) {
	fromShard, toShard := getCoveredShard(im.shardSize, cursor.offset, size)

	ctx.WithFields(log.Fields{
		"cursor":     cursor,
		"startShard": fromShard,
		"endShard":   toShard,
	}).Info("getCoveredShards info")

	data := []byte{}
	for i := fromShard; i < toShard; i++ {
		cacheKey := im.getCacheKey(key, cursor.createTs, i)
		shard, err := im.redisCache.Get(ctx, cacheKey)
		if err == redis.ErrNotFound {
			return nil, ErrBadCursor
		}
		if err != nil {
			ctx.WithFields(log.Fields{
				"err": err,
			}).Error("failed to Get")
			return nil, err
		}
		data = append(data, shard...)
	}

	return data, nil
}

func (im *impl) getFromCache(ctx ctx.Ctx, key string, cursor cursorStruct, size int, container interface{}) (string, int, error) {
	totalCount := cursor.totalCount
	offset := cursor.offset
	if offset+size > totalCount {
		size = totalCount - offset
	}

	rawData, err := im.getRawData(ctx, key, cursor, size)
	if err != nil {
		ctx.WithFields(log.Fields{
			"err":    err,
			"cursor": cursor,
		}).Error("failed to getCoveredShards")
		return "", 0, err
	}

	tElem := reflect.TypeOf(container).Elem().Elem() // elem
	tList := reflect.SliceOf(tElem)                  // []elem
	list := reflect.New(tList)

	err = gob.NewDecoder(bytes.NewBuffer(rawData)).Decode(list.Interface())
	if err != nil {
		ctx.WithFields(log.Fields{
			"err": err,
		}).Error("failed to decode data")
		return "", 0, err
	}

	start := offset % im.shardSize
	reflect.ValueOf(container).Elem().Set(list.Elem().Slice(start, start+size))

	nextOffset := offset + size
	nextCursorString := ""
	if nextOffset < totalCount {
		nextCursorString = encodeCursor(cursorStruct{
			offset:     nextOffset,
			totalCount: totalCount,
			createTs:   cursor.createTs,
		})
	}

	return nextCursorString, totalCount, nil
}

func (im *impl) getCursorFromLatestKey(ctx ctx.Ctx) (*cursorStruct, error) {
	val, err := im.redisCache.Get(ctx, im.latestKey)
	if err == redis.ErrNotFound {
		return nil, nil
	}
	if err != nil {
		ctx.WithFields(log.Fields{
			"err":       err,
			"latestKey": im.latestKey,
		}).Error("failed to Get")
		return nil, err
	}

	return decodeCursor(string(val))
}

func (im *impl) storeLatestKey(ctx ctx.Ctx, startCursor cursorStruct) error {
	err := im.redisCache.Set(ctx, im.latestKey, []byte(encodeCursor(startCursor)), im.cacheDuration)
	if err != nil {
		ctx.WithFields(log.Fields{
			"err": err,
		}).Error("failed to Set")
		return err
	}
	return nil
}

func (im *impl) getCacheKey(key string, createTs int64, shardNum int) string {
	return keys.RedisKey(keys.PfxPagingService, im.keyPfx, key, strconv.FormatInt(createTs, 10), strconv.Itoa(shardNum))
}

func (im *impl) storeShardToCache(ctx ctx.Ctx, key string, elems reflect.Value, errChan chan error) {
	im.workerPool.Schedule(func() {
		var val bytes.Buffer
		err := gob.NewEncoder(&val).EncodeValue(elems)
		if err != nil {
			ctx.WithFields(log.Fields{
				"err": err,
			}).Error("failed to EncodeValue")
			errChan <- err
			return
		}
		err = im.redisCache.Set(ctx, key, val.Bytes(), im.cacheDuration)
		if err != nil {
			ctx.WithFields(log.Fields{
				"err": err,
				"key": key,
			}).Error("failed to Set")
			errChan <- err
			return
		}
		errChan <- nil
	})
}

func (im *impl) storeListToCache(ctx ctx.Ctx, key string, createTs int64, wholeList interface{}) error {
	err := im.redisCache.SetNX(ctx, im.lockKey, []byte("1"), im.getterTimeout)
	if err == redis.ErrNotFound {
		return nil
	}
	if err != nil {
		ctx.WithFields(log.Fields{
			"err": err,
		}).Error("failed to redisCache.SetNX")
		return err
	}

	defer func() {
		im.redisCache.Del(ctx, im.lockKey)
	}()

	list := reflect.ValueOf(wholeList)
	totalCount := list.Len()

	shards := math.CeilInt(totalCount, im.shardSize)

	errChan := make(chan error, shards)

	// TODO: use goroutine to accelerate process
	shardNum := 0
	for shardNum*im.shardSize < totalCount {
		shardStart := shardNum * im.shardSize
		shardEnd := shardStart + im.shardSize
		if shardEnd > totalCount {
			shardEnd = totalCount
		}
		cacheKey := im.getCacheKey(key, createTs, shardNum)
		elems := list.Slice(shardStart, shardEnd)

		im.storeShardToCache(ctx, cacheKey, elems, errChan)

		shardNum += 1
	}

	for i := 0; i < shards; i++ {
		if err := <-errChan; err != nil {
			ctx.WithFields(log.Fields{
				"err": err,
			}).Error("failed to storeShardToCache")
			return err
		}
	}

	startCursor := cursorStruct{
		offset:     0,
		totalCount: totalCount,
		createTs:   createTs,
	}
	err = im.storeLatestKey(ctx, startCursor)
	if err != nil {
		ctx.WithFields(log.Fields{
			"err":    err,
			"cursor": startCursor,
		}).Error("failed to storeLatestKey")
		return err
	}

	return nil
}

// getFromGetter returns nextCursor, totalCount, error
func (im *impl) getFromGetter(ctx ctx.Ctx, key string, size int, container interface{}) (string, int, error) {
	wholeList, err := im.getter(ctx, key)
	if err != nil {
		ctx.WithFields(log.Fields{
			"err": err,
		}).Error("failed to getter")
		return "", 0, err
	}

	v := reflect.ValueOf(wholeList)
	totalCount := v.Len()

	if totalCount < size {
		size = totalCount
	}

	// assign value to container
	var buf bytes.Buffer
	err = gob.NewEncoder(&buf).EncodeValue(v.Slice(0, size))
	if err != nil {
		ctx.WithFields(log.Fields{
			"err": err,
		}).Error("failed to Encode")
		return "", 0, err
	}
	err = gob.NewDecoder(&buf).Decode(container)
	if err != nil {
		ctx.WithFields(log.Fields{
			"err": err,
		}).Error("failed to Decode")
		return "", 0, err
	}

	createTs := time.Now().UnixNano()

	err = im.storeListToCache(ctx, key, createTs, wholeList)
	if err != nil {
		ctx.WithFields(log.Fields{
			"err": err,
		}).Error("failed to storeListToCache")
		return "", 0, err
	}

	nextCursor := cursorStruct{
		offset:     size,
		totalCount: totalCount,
		createTs:   createTs,
	}

	return encodeCursor(nextCursor), totalCount, nil
}

// isSlicePtr check if an object is in the following type
// - *[]*struct
// - *[]struct
func isSlicePtr(container interface{}) bool {
	if container == nil {
		return false
	}

	v := reflect.ValueOf(container)
	// check pointer
	if v.Kind() != reflect.Ptr {
		return false
	}

	e := v.Elem()
	// check slice
	return e.Kind() == reflect.Slice
}

// Get returns nextCursor, totalCount, error
func (im *impl) Get(ctx ctx.Ctx, key string, cursorString string, size int, container interface{}) (string, int, error) {
	if !isSlicePtr(container) {
		return "", 0, ErrBadContainer
	}

	// - decode cursor if cursor exists, else generate cursor at start point
	// - get from cache
	// - if cache invalid, get from getter
	var cursor *cursorStruct
	var err error
	if cursorString != "" {
		cursor, err = decodeCursor(cursorString)
		if err != nil {
			return "", 0, ErrBadCursor
		}
		// cache invalid
		if cursor != nil && time.Now().After(time.Unix(0, cursor.createTs).Add(im.cacheDuration)) {
			cursor = nil
		}
	} else {
		cursor, err = im.getCursorFromLatestKey(ctx)
		if err != nil {
			return "", 0, ErrGetLatestKey
		}
		// cache need renew
		if cursor != nil && time.Now().After(time.Unix(0, cursor.createTs).Add(im.renewDuration)) {
			cursor = nil
		}
	}

	if cursor != nil {
		return im.getFromCache(ctx, key, *cursor, size, container)
	} else {
		return im.getFromGetter(ctx, key, size, container)
	}
}

func (im *impl) Update(ctx ctx.Ctx) error {
	return nil
}
