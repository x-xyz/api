package goroutine

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRecoverableGo(t *testing.T) {
	res := []string{}

	<-RecoverableGo(
		func() {
			res = append(res, "run task")
			panic("panic")
		},
		WithBeforeStart(func() {
			res = append(res, "before start")
		}),
		WithAfterEnded(func() {
			res = append(res, "after ended")
		}),
		WithAfterRecovered(func(p interface{}, stack []byte) {
			res = append(res, "after recovered")
			res = append(res, p.(string))
		}),
	)

	assert.Equal(t, []string{
		"before start",
		"run task",
		"after ended",
		"after recovered",
		"panic",
	}, res)
}
