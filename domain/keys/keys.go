package keys

import (
	"crypto/md5"
	"fmt"
	"strings"
)

const (
	// PfxHealthCheck is used for prefixing health check redis key
	PfxHealthCheck = "healthcheck"
	// PfxNonce is used for prefixing nonce redis key
	PfxNonce = "nonce"
	// PfxPagingService is used for prefixing paging data
	PfxPagingService = "pagingService"
	// PfxSearchV2Paging is used for prefixing searchV2 paging
	PfxSearchV2Paging = "searchV2Paging"
)

// MD5 hashes the data with md5
func MD5(data string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(data)))
}

// CustomKey is used to join the customized key by componets with specified delimiter
func CustomKey(delimiter string, components ...string) string {
	return strings.Join(components, delimiter)
}

// RedisKey is used to join the redis key by componets
func RedisKey(components ...string) string {
	return CustomKey(":", components...)
}

// RedisLuaKey is used to join the redis key by componets for redis lua
// If a key created by RedisLuaKey prefix to a set of keys
// then the set of keys will be forced in the same shard for doing lua
func RedisLuaKey(components ...string) string {
	return "{" + CustomKey(":", components...) + "}"
}

// RedisLuaMultiKey is used to make sure all the keys passed to lua script will
// have the same hash. Redis key distribution is based on key hash and the keys
// operated in lua script must on the same server in the redis cluster.
//
// Multiple key operations can be ensured to have the same hash by using hash
// tags which are the substring in the first occurrence of `{` and `}`.
//
// Ref: https://redis.io/topics/cluster-spec#keys-hash-tags
//
// Example:
//   a := "a:b:1:2"
//   b := "a:b:3:4"
//   ret, _ := RedisLuaMultiKey(a, b)
//
//   ret: ["{a:b}:1:2", "{a:b}:3:4"]
func RedisLuaMultiKey(redisKeys ...string) ([]interface{}, error) {
	prefix := lcp(redisKeys)
	if len(prefix) == 0 {
		return nil, fmt.Errorf("No common prefix")
	}
	if strings.HasSuffix(prefix, ":") {
		prefix = string(prefix[:len(prefix)-len(":")])
	}
	newPrefix := "{" + prefix + "}"
	newKeys := []interface{}{}
	for _, k := range redisKeys {
		newKey := newPrefix + k[len(prefix):]
		newKeys = append(newKeys, newKey)
	}
	return newKeys, nil
}

// Longest common prefix
// Ref: https://rosettacode.org/wiki/Longest_common_prefix#Go
func lcp(l []string) string {
	// Special cases first
	switch len(l) {
	case 0:
		return ""
	case 1:
		return l[0]
	}
	// LCP of min and max (lexigraphically)
	// is the LCP of the whole set.
	min, max := l[0], l[0]
	for _, s := range l[1:] {
		switch {
		case s < min:
			min = s
		case s > max:
			max = s
		}
	}
	for i := 0; i < len(min) && i < len(max); i++ {
		if min[i] != max[i] {
			return min[:i]
		}
	}
	// In the case where lengths are not equal but all bytes
	// are equal, min is the answer ("foo" < "foobar").
	return min
}

// GetPrefix extracts the prefix of a key.
// will take more than one prefix. And if prefix start with capital
// letter, which means it's a table, a `Table:` prefix will be added.
func GetPrefix(key string) string {
	s := strings.Split(key, ":")
	if len(s) > 2 {
		return strings.Join([]string{s[0], s[1]}, ":")
	} else if len(s) > 3 {
		return strings.Join([]string{s[0], s[1], s[2]}, ":")
	} else if len(s) > 1 {
		return s[0]
	}
	return ""
}
