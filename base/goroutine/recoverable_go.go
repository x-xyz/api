package goroutine

import (
	"github.com/x-xyz/goapi/base/log"
	"github.com/x-xyz/goapi/base/utils"
)

var (
	logger = log.Log()
)

type PanicEvent struct {
	Panic interface{}
	Stack []byte
}

type RecoverableGoOptions struct {
	beforeStart    *func()
	afterEnded     *func()
	afterRecovered *func(panic interface{}, stake []byte)
}

type RecoverableGoOptionsFunc = func(*RecoverableGoOptions) error

func getRecoverableGoOptions(fns ...RecoverableGoOptionsFunc) RecoverableGoOptions {
	opts := RecoverableGoOptions{}
	for _, fn := range fns {
		fn(&opts)
	}
	return opts
}

func WithBeforeStart(f func()) RecoverableGoOptionsFunc {
	return func(options *RecoverableGoOptions) error {
		options.beforeStart = &f
		return nil
	}
}

func WithAfterEnded(f func()) RecoverableGoOptionsFunc {
	return func(options *RecoverableGoOptions) error {
		options.afterEnded = &f
		return nil
	}
}

func WithAfterRecovered(f func(panic interface{}, stake []byte)) RecoverableGoOptionsFunc {
	return func(options *RecoverableGoOptions) error {
		options.afterRecovered = &f
		return nil
	}
}

func RecoverableGo(f func(), fns ...RecoverableGoOptionsFunc) chan *PanicEvent {
	opts := getRecoverableGoOptions(fns...)

	panicChan := make(chan *PanicEvent, 1)

	go func() {
		defer func() {
			if opts.afterEnded != nil {
				(*opts.afterEnded)()
			}

			if p := recover(); p != nil {
				stack := utils.Stack(3)

				logger.WithFields(log.Fields{
					"err":   p,
					"stack": string(stack),
				}).Error("panic")

				if opts.afterRecovered != nil {
					(*opts.afterRecovered)(p, stack)
				}

				panicChan <- &PanicEvent{p, stack}
			} else {
				close(panicChan)
			}
		}()

		if opts.beforeStart != nil {
			(*opts.beforeStart)()
		}

		f()
	}()

	return panicChan
}
