/*Package metrics wraps datadog-go to faciliate metric recording
Following are naming convention of metric:
- Internal process time: *.time
- External latency: *.latency
- Error: *.err
- Warning: *.warn
*/
package metrics

import (
	"math/rand"
	"strings"
	"time"

	"github.com/spf13/viper"
	"github.com/x-xyz/goapi/base/env"
)

const (
	// TagValueNA is used for tags whose values are not available.
	TagValueNA = "n/a"
)

// Ender provides interface for BumpHistogram
type Ender interface {
	End()
}

// Service provides interface for metrics
type Service interface {
	BumpAvg(key string, val float64, tags ...string)
	BumpSum(key string, val float64, tags ...string)
	BumpHistogram(key string, val float64, tags ...string)

	BumpTime(key string, tags ...string) Ender
}

// Option is functional parameter for metrics option
type Option func(*opt)

// opt is struct of metrics options
type opt struct {
	// withPodName means send metrics with pod name or not
	// default: true
	withPodName bool
}

// WithoutPodName means the metrics sent by the Service will not contain pod name
// Pod name produces a lot of custom metrics. If it is unnecessary to group metrics by pod name, using it to disable pod name tag
func WithoutPodName() Option {
	return func(o *opt) {
		o.withPodName = false
	}
}

// New creates a facebook/stats compatible metric client with package name as prefix
func New(pkgName string, options ...Option) Service {
	o := opt{
		withPodName: true,
	}
	for _, option := range options {
		option(&o)
	}

	ddTags := []string{}
	if o.withPodName {
		ddTags = []string{
			// using host removes all tags associated with host
			// ref: https://docs.datadoghq.com/developers/dogstatsd/data_types/#host-tag-key
			"host:", // remove unused host tag
			"pod:" + env.PodName(),
			"env:" + viper.GetString("env_name"),
			"app:" + viper.GetString("app_name"),
		}
	} else {
		ddTags = []string{
			// using host removes all tags associated with host
			// ref: https://docs.datadoghq.com/developers/dogstatsd/data_types/#host-tag-key
			"host:", // remove unused host tag
			"env:" + viper.GetString("env_name"),
			"app:" + viper.GetString("app_name"),
		}
	}

	return &Metrics{
		pkgName: pkgName,
		level:   3,
		datadog: DDMetrics{
			ddTags: ddTags,
		},
	}
}

// Metrics wraps datadog-go to be facebookgo/stat.Client interface.
// See https://godoc.org/github.com/facebookgo/stats#Client for interface details.
type Metrics struct {
	pkgName string
	level   int
	datadog DDMetrics
}

// a BumpXXX can be dropped due to configuration / random sampling
// shouldGiveUp is a function to wrap such logic
func (mt *Metrics) shouldGiveUp(pkgName string, level int) bool {
	// inWhiteList, ok := Conf.WhiteList[pkgName]

	// return !(level <= Conf.EnabledLevel || (inWhiteList && ok))
	return false
}

// sampleRate returns the pkg's metrics firing rate defined in etcd.
// sampleRate can range from 0 to 1, and 1 means always send the metrics.
func (mt *Metrics) sampleRate(pkgName string) float64 {
	// Use pkg's sample rate if it is assigned, else use the
	// default value 1.
	// if rate, ok := Conf.SampleRate[pkgName]; ok {
	// 	return rate
	// }
	// return Conf.DefaultSampleRate
	return 1.0
}

// bumpSumPanic handles panics for all metrics vendor.
// inconsistent tagging.
func (mt *Metrics) bumpSumPanic(key, tag string) {
	mt.datadog.BumpSum(key, 1, 1, "tag", tag)
}

// bumpLatency records each bump's latency with only 0.0001 sampling rate
func (mt *Metrics) bumpLatency(typ string, start time.Time, sampleRate float64) {
	if rand.Float64() < float64(0.0001)*sampleRate {
		mt.datadog.BumpHistogram("bump.latency", float64(time.Since(start)/time.Millisecond), 1, "name", mt.pkgName, "type", typ)
	}
}

// BumpAvg bumps the average for the given key.
func (mt *Metrics) BumpAvg(key string, val float64, tags ...string) {
	if mt.shouldGiveUp(mt.pkgName, mt.level) {
		return
	}

	defer func() {
		if err := recover(); err != nil {
			mt.bumpSumPanic("bumpavg.panic", mt.pkgName+`.`+key+"#"+strings.Join(tags, "#"))
		}
	}()

	sampleRate := mt.sampleRate(mt.pkgName)
	defer mt.bumpLatency("bumpavg", time.Now(), sampleRate)

	// push data to datadog.
	mt.datadog.BumpAvg(mt.pkgName+`.`+key, val, sampleRate, tags...)
}

// BumpSum bumps the sum for the given key.
func (mt *Metrics) BumpSum(key string, val float64, tags ...string) {
	if mt.shouldGiveUp(mt.pkgName, mt.level) {
		return
	}
	defer func() {
		if err := recover(); err != nil {
			mt.bumpSumPanic("bumpsum.panic", mt.pkgName+`.`+key+"#"+strings.Join(tags, "#"))
		}
	}()

	sampleRate := mt.sampleRate(mt.pkgName)
	defer mt.bumpLatency("bumpsum", time.Now(), sampleRate)

	// push data to datadog.
	mt.datadog.BumpSum(mt.pkgName+`.`+key, val, sampleRate, tags...)
}

// BumpHistogram bumps the histogram for the given key.
func (mt *Metrics) BumpHistogram(key string, val float64, tags ...string) {
	if mt.shouldGiveUp(mt.pkgName, mt.level) {
		return
	}

	defer func() {
		if err := recover(); err != nil {
			mt.bumpSumPanic("bumphistogram.panic", mt.pkgName+`.`+key+"#"+strings.Join(tags, "#"))
		}
	}()

	sampleRate := mt.sampleRate(mt.pkgName)
	defer mt.bumpLatency("bumphistogram", time.Now(), sampleRate)

	// push data to datadog.
	mt.datadog.BumpHistogram(mt.pkgName+`.`+key, val, sampleRate, tags...)
}

// BumpTime is a special version of BumpHistogram which is specialized for
// timers. Calling it starts the timer, and it returns a value on which End()
// can be called to indicate finishing the timer. A convenient way of
// recording the duration of a function is calling it like such at the top of
// the function:
//
//     defer s.BumpTime("my.function").End()
func (mt *Metrics) BumpTime(key string, tags ...string) Ender {
	if mt.shouldGiveUp(mt.pkgName, mt.level) {
		return &timeTracker{
			ddEnd:        &fakeEnd{},
			panicHandler: func() {},
		}
	}

	// push data to datadog.
	sampleRate := mt.sampleRate(mt.pkgName)
	ddEnd := mt.datadog.BumpTime(mt.pkgName+`.`+key, sampleRate, tags...)

	return &timeTracker{
		ddEnd:       ddEnd,
		sampleRate:  sampleRate,
		bumpLatency: mt.bumpLatency,
		panicHandler: func() {
			mt.bumpSumPanic("bumptime.panic", mt.pkgName+`.`+key+"#"+strings.Join(tags, "#"))
		},
	}
}

type fakeEnd struct {
}

func (e *fakeEnd) End() {
}

type timeTracker struct {
	ddEnd interface {
		End()
	}
	sampleRate   float64
	bumpLatency  func(string, time.Time, float64)
	panicHandler func()
}

func (t *timeTracker) End() {
	defer func() {
		if err := recover(); err != nil {
			t.panicHandler()
		}
	}()

	defer t.bumpLatency("bumptime", time.Now(), t.sampleRate)

	// end datadog counter.
	t.ddEnd.End()
	return
}
