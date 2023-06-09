// Code generated by mockery v2.16.0. DO NOT EDIT.

package mocks

import (
	ctx "github.com/x-xyz/goapi/base/ctx"
	apecoinstaking "github.com/x-xyz/goapi/domain/apecoinstaking"

	mock "github.com/stretchr/testify/mock"
)

// UseCase is an autogenerated mock type for the UseCase type
type UseCase struct {
	mock.Mock
}

type UseCase_Expecter struct {
	mock *mock.Mock
}

func (_m *UseCase) EXPECT() *UseCase_Expecter {
	return &UseCase_Expecter{mock: &_m.Mock}
}

// Get provides a mock function with given fields: _a0, id
func (_m *UseCase) Get(_a0 ctx.Ctx, id apecoinstaking.Id) (*apecoinstaking.ApecoinStaking, error) {
	ret := _m.Called(_a0, id)

	var r0 *apecoinstaking.ApecoinStaking
	if rf, ok := ret.Get(0).(func(ctx.Ctx, apecoinstaking.Id) *apecoinstaking.ApecoinStaking); ok {
		r0 = rf(_a0, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*apecoinstaking.ApecoinStaking)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(ctx.Ctx, apecoinstaking.Id) error); ok {
		r1 = rf(_a0, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UseCase_Get_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Get'
type UseCase_Get_Call struct {
	*mock.Call
}

// Get is a helper method to define mock.On call
//   - _a0 ctx.Ctx
//   - id apecoinstaking.Id
func (_e *UseCase_Expecter) Get(_a0 interface{}, id interface{}) *UseCase_Get_Call {
	return &UseCase_Get_Call{Call: _e.mock.On("Get", _a0, id)}
}

func (_c *UseCase_Get_Call) Run(run func(_a0 ctx.Ctx, id apecoinstaking.Id)) *UseCase_Get_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(ctx.Ctx), args[1].(apecoinstaking.Id))
	})
	return _c
}

func (_c *UseCase_Get_Call) Return(_a0 *apecoinstaking.ApecoinStaking, _a1 error) *UseCase_Get_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// Upsert provides a mock function with given fields: _a0, s
func (_m *UseCase) Upsert(_a0 ctx.Ctx, s *apecoinstaking.ApecoinStaking) error {
	ret := _m.Called(_a0, s)

	var r0 error
	if rf, ok := ret.Get(0).(func(ctx.Ctx, *apecoinstaking.ApecoinStaking) error); ok {
		r0 = rf(_a0, s)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UseCase_Upsert_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Upsert'
type UseCase_Upsert_Call struct {
	*mock.Call
}

// Upsert is a helper method to define mock.On call
//   - _a0 ctx.Ctx
//   - s *apecoinstaking.ApecoinStaking
func (_e *UseCase_Expecter) Upsert(_a0 interface{}, s interface{}) *UseCase_Upsert_Call {
	return &UseCase_Upsert_Call{Call: _e.mock.On("Upsert", _a0, s)}
}

func (_c *UseCase_Upsert_Call) Run(run func(_a0 ctx.Ctx, s *apecoinstaking.ApecoinStaking)) *UseCase_Upsert_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(ctx.Ctx), args[1].(*apecoinstaking.ApecoinStaking))
	})
	return _c
}

func (_c *UseCase_Upsert_Call) Return(_a0 error) *UseCase_Upsert_Call {
	_c.Call.Return(_a0)
	return _c
}

type mockConstructorTestingTNewUseCase interface {
	mock.TestingT
	Cleanup(func())
}

// NewUseCase creates a new instance of UseCase. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewUseCase(t mockConstructorTestingTNewUseCase) *UseCase {
	mock := &UseCase{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
