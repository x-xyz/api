package domain

import (
	"github.com/x-xyz/goapi/base/ctx"
)

const DefaultTag = "default"

type TrackerState struct {
	ChainId               ChainId `bson:"chainId"`
	ContractAddress       Address `bson:"contractAddress"`
	Tag                   string  `bson:"tag"`
	Version               uint64  `bson:"version"`
	LastBlockProcessed    uint64  `bson:"lastBlockProcessed"`
	LastLogIndexProcessed int64   `bson:"lastLogIndexProcessed"`
}

func (s *TrackerState) ToId() *TrackerStateId {
	return &TrackerStateId{
		ChainId:         s.ChainId,
		ContractAddress: s.ContractAddress,
		Tag:             s.Tag,
	}
}

type TrackerStateId struct {
	ChainId         ChainId `bson:"chainId"`
	ContractAddress Address `bson:"contractAddress"`
	Tag             string  `bson:"tag"`
}

type TrackerStateRepo interface {
	Get(ctx.Ctx, *TrackerStateId) (*TrackerState, error)
	Update(ctx.Ctx, *TrackerState) error
	Store(ctx.Ctx, *TrackerState) error
}

type TrackerStateUseCase interface {
	Get(ctx.Ctx, *TrackerStateId) (*TrackerState, error)
	Update(ctx.Ctx, *TrackerState) error
	Store(ctx.Ctx, *TrackerState) error
}
