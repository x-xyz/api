package metrics

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/DataDog/datadog-go/statsd"
	"github.com/spf13/viper"

	"github.com/x-xyz/goapi/base/log"
)

const (
	ddClientsSize    = 16 // needs to be 2^n
	ddClientsIdxMask = ddClientsSize - 1
)

var (
	initOnce = sync.Once{}

	// DdHost is the public form of DdHost
	// DdHost = flag.String("datadog_host", "", "datadog agent host name")
	// DdHost = os.Getenv("HOSTIP")
	DdHost = ""
	// DdPort is the public form of DdPort
	DdPort = 8125

	// ddClientsIdx is used for accessing ddClients by round robin scheduling
	ddClientsIdx = int32(0)
	ddClients    []statsCli
)

const (
	// ddRate is the rate to pass metrics to datadog agent. 1 means always
	ddRate = 1
	// buffer 10 counters before sending to statsd
	bufferMetrics = 10
)

func initDDClient() {
	DdHost = viper.GetString("datadog_host")
	ddClients = make([]statsCli, ddClientsSize)
	for i := 0; i < ddClientsSize; i++ {

		// Init datadog client once so the buffer is counted together. Also it's better to
		// maintain one connection toward statsd agent
		addr := fmt.Sprintf("%s:%d", DdHost, DdPort)
		log.Log().WithFields(log.Fields{"addr": addr, "idx": i}).Info("connecting to datadog agent")

		var err error
		ddClients[i], err = statsd.NewBuffered(addr, bufferMetrics)
		if err != nil {
			log.Log().WithFields(log.Fields{"addr": addr, "err": err}).Panic(
				"can't talk to datadog agent")
		}
	}
}

type statsCli interface {
	Gauge(name string, value float64, tags []string, rate float64) error
	Count(name string, value int64, tags []string, rate float64) error
	Histogram(name string, value float64, tags []string, rate float64) error
	TimeInMilliseconds(name string, value float64, tags []string, rate float64) error
}

// DDMetrics wraps datadog statsd metrics implement facebookgo/stats.Client interface.
// See https://godoc.org/github.com/facebookgo/stats#Client for interface details.
type DDMetrics struct {
	ddTags []string
}

// BumpAvg bumps the average for the given key.
func (dm *DDMetrics) BumpAvg(key string, val, sampleRate float64, tags ...string) {
	initOnce.Do(initDDClient)
	// datadog doesn't have a function to compute average only. Work-around by calculating
	// histogram (which is overkill however)

	i := atomic.AddInt32(&ddClientsIdx, 1) & ddClientsIdxMask
	if err := ddClients[i].Gauge(key, val, append(dm.ddTags, parseTag(tags)...), sampleRate); err != nil {
		log.Log().WithFields(log.Fields{"err": err, "key": key, "val": val, "func": "BumpAvg"}).Error("Bump fail")
	}
}

// BumpSum bumps the sum for the given key.
func (dm *DDMetrics) BumpSum(key string, val, sampleRate float64, tags ...string) {
	initOnce.Do(initDDClient)
	i := atomic.AddInt32(&ddClientsIdx, 1) & ddClientsIdxMask
	if err := ddClients[i].Count(key, int64(val), append(dm.ddTags, parseTag(tags)...), sampleRate); err != nil {
		log.Log().WithFields(log.Fields{"err": err, "key": key, "val": val, "func": "BumpSum"}).Error("Bump fail")
	}
}

// BumpHistogram bumps the histogram for the given key.
func (dm *DDMetrics) BumpHistogram(key string, val, sampleRate float64, tags ...string) {
	initOnce.Do(initDDClient)
	i := atomic.AddInt32(&ddClientsIdx, 1) & ddClientsIdxMask
	if err := ddClients[i].Histogram(key, val, append(dm.ddTags, parseTag(tags)...), sampleRate); err != nil {
		log.Log().WithFields(log.Fields{"err": err, "key": key, "val": val, "func": "BumpHistogram"}).Error("Bump fail")
	}
}

// BumpTime is a special version of BumpHistogram which is specialized for
// timers. Calling it starts the timer, and it returns a value on which End()
// can be called to indicate finishing the timer. A convenient way of
// recording the duration of a function is calling it like such at the top of
// the function:
//
//     defer s.BumpTime("my.function").End()
func (dm *DDMetrics) BumpTime(key string, sampleRate float64, tags ...string) interface {
	End()
} {
	initOnce.Do(initDDClient)
	return &ddTimeTracker{
		start:      time.Now(),
		key:        key,
		tags:       append(dm.ddTags, parseTag(tags)...),
		sampleRate: sampleRate,
	}
}

func parseTag(tags []string) []string {
	if tags == nil {
		return nil
	}
	if len(tags)%2 != 0 {
		log.Log().WithField("tags", tags).Panic("tag length needs to be multiple of 2")
	}
	arr := make([]string, len(tags)/2)
	for i := 0; i < len(tags); i += 2 {
		arr[i/2] = tags[i] + ":" + tags[i+1]
	}
	return arr
}

type ddTimeTracker struct {
	start      time.Time
	key        string
	tags       []string
	sampleRate float64
}

func (dt *ddTimeTracker) End() {
	d := time.Since(dt.start)
	msec := d / time.Millisecond
	nsec := d % time.Millisecond

	dur := float64(msec) + float64(nsec)*1e-6

	i := atomic.AddInt32(&ddClientsIdx, 1) & ddClientsIdxMask
	if err := ddClients[i].TimeInMilliseconds(dt.key, dur, dt.tags, dt.sampleRate); err != nil {
		log.Log().WithFields(log.Fields{"err": err, "key": dt.key, "val": dur, "func": "BumpTime"}).Error("Bump fail")
	}
}
