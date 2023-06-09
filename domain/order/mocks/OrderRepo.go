// Code generated by mockery v2.9.4. DO NOT EDIT.

package mocks

import (
	mock "github.com/stretchr/testify/mock"
	ctx "github.com/x-xyz/goapi/base/ctx"

	order "github.com/x-xyz/goapi/domain/order"
)

// OrderRepo is an autogenerated mock type for the OrderRepo type
type OrderRepo struct {
	mock.Mock
}

// Count provides a mock function with given fields: _a0, opts
func (_m *OrderRepo) Count(_a0 ctx.Ctx, opts ...order.OrderFindAllOptionsFunc) (int, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, _a0)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 int
	if rf, ok := ret.Get(0).(func(ctx.Ctx, ...order.OrderFindAllOptionsFunc) int); ok {
		r0 = rf(_a0, opts...)
	} else {
		r0 = ret.Get(0).(int)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(ctx.Ctx, ...order.OrderFindAllOptionsFunc) error); ok {
		r1 = rf(_a0, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// FindAll provides a mock function with given fields: _a0, opts
func (_m *OrderRepo) FindAll(_a0 ctx.Ctx, opts ...order.OrderFindAllOptionsFunc) ([]*order.Order, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, _a0)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 []*order.Order
	if rf, ok := ret.Get(0).(func(ctx.Ctx, ...order.OrderFindAllOptionsFunc) []*order.Order); ok {
		r0 = rf(_a0, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*order.Order)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(ctx.Ctx, ...order.OrderFindAllOptionsFunc) error); ok {
		r1 = rf(_a0, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// FindOne provides a mock function with given fields: _a0, id
func (_m *OrderRepo) FindOne(_a0 ctx.Ctx, id order.OrderId) (*order.Order, error) {
	ret := _m.Called(_a0, id)

	var r0 *order.Order
	if rf, ok := ret.Get(0).(func(ctx.Ctx, order.OrderId) *order.Order); ok {
		r0 = rf(_a0, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*order.Order)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(ctx.Ctx, order.OrderId) error); ok {
		r1 = rf(_a0, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RemoveAll provides a mock function with given fields: _a0, opts
func (_m *OrderRepo) RemoveAll(_a0 ctx.Ctx, opts ...order.OrderFindAllOptionsFunc) error {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, _a0)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 error
	if rf, ok := ret.Get(0).(func(ctx.Ctx, ...order.OrderFindAllOptionsFunc) error); ok {
		r0 = rf(_a0, opts...)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Update provides a mock function with given fields: _a0, id, patchable
func (_m *OrderRepo) Update(_a0 ctx.Ctx, id order.OrderId, patchable order.OrderPatchable) error {
	ret := _m.Called(_a0, id, patchable)

	var r0 error
	if rf, ok := ret.Get(0).(func(ctx.Ctx, order.OrderId, order.OrderPatchable) error); ok {
		r0 = rf(_a0, id, patchable)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Upsert provides a mock function with given fields: _a0, _a1
func (_m *OrderRepo) Upsert(_a0 ctx.Ctx, _a1 *order.Order) error {
	ret := _m.Called(_a0, _a1)

	var r0 error
	if rf, ok := ret.Get(0).(func(ctx.Ctx, *order.Order) error); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
