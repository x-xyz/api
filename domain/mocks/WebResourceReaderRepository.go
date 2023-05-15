// Code generated by mockery v2.9.4. DO NOT EDIT.

package mocks

import (
	ctx "github.com/x-xyz/goapi/base/ctx"

	mock "github.com/stretchr/testify/mock"
)

// WebResourceReaderRepository is an autogenerated mock type for the WebResourceReaderRepository type
type WebResourceReaderRepository struct {
	mock.Mock
}

// Get provides a mock function with given fields: _a0, _a1
func (_m *WebResourceReaderRepository) Get(_a0 ctx.Ctx, _a1 string) ([]byte, error) {
	ret := _m.Called(_a0, _a1)

	var r0 []byte
	if rf, ok := ret.Get(0).(func(ctx.Ctx, string) []byte); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(ctx.Ctx, string) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
