package tracker

import (
	"context"
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	bCtx "github.com/x-xyz/goapi/base/ctx"
	"github.com/x-xyz/goapi/domain"
	"github.com/x-xyz/goapi/domain/mocks"
	"github.com/x-xyz/goapi/service/query"
)

func Test_getDeployedBlock(t *testing.T) {
	req := require.New(t)
	tests := []struct {
		currentBlock  uint64
		deployedBlock uint64
	}{
		{currentBlock: 12345, deployedBlock: 3},
		{currentBlock: 12345, deployedBlock: 4},
		{currentBlock: 12345, deployedBlock: 6000},
		{currentBlock: 12345, deployedBlock: 6001},
		{currentBlock: 12345, deployedBlock: 12000},
		{currentBlock: 12345, deployedBlock: 12001},
		{currentBlock: 12346, deployedBlock: 3},
		{currentBlock: 12346, deployedBlock: 4},
		{currentBlock: 12346, deployedBlock: 6000},
		{currentBlock: 12346, deployedBlock: 6001},
		{currentBlock: 12346, deployedBlock: 12000},
		{currentBlock: 12346, deployedBlock: 12001},
	}
	ctx := bCtx.Background()
	for _, tt := range tests {
		name := fmt.Sprintf("%d/%d", tt.deployedBlock, tt.currentBlock)
		t.Run(name, func(t *testing.T) {
			client := new(mocks.EthClientRepo)
			addr := common.Address{}
			client.On("BlockNumber", mock.Anything).Return(tt.currentBlock, nil)
			client.On("CodeAt",
				mock.Anything,
				mock.AnythingOfType("common.Address"),
				mock.AnythingOfType("*big.Int"),
			).Return(
				codeAtFunc(tt.deployedBlock),
				nil,
			)
			blk, err := getDeployedBlock(ctx, client, addr)
			req.NoError(err)
			req.Equal(tt.deployedBlock, blk)
		})
	}
}

func TestEventTracker_setupTrackerState(t *testing.T) {
	chainId := int64(1)
	contractAddr := common.BigToAddress(big.NewInt(1))
	contractAddrStr := ToLowerHexStr(contractAddr)

	t.Run("exists in repo", func(t *testing.T) {
		req := require.New(t)
		ctx := bCtx.Background()
		trackerStateUseCase := new(mocks.TrackerStateUseCase)
		f := &EventTracker{
			chainId:             chainId,
			trackerStateUseCase: trackerStateUseCase,
			contractAddress:     contractAddr,
		}

		status := &domain.TrackerState{ChainId: domain.ChainId(chainId), ContractAddress: domain.Address(contractAddrStr), LastBlockProcessed: 20}
		trackerStateUseCase.On("Get", mock.Anything, status.ToId()).Return(status, nil)

		got, err := f.setupTrackerState(ctx)
		req.NoError(err)
		req.Equal(status, got)
	})

	t.Run("get from deployed block", func(t *testing.T) {
		req := require.New(t)
		ctx := bCtx.Background()
		trackerStateUseCase := new(mocks.TrackerStateUseCase)
		ethClient := new(mocks.EthClientRepo)
		f := &EventTracker{
			chainId:             chainId,
			currentBlockGetter:  ethClient,
			wsClient:            ethClient,
			clientWithArchive:   ethClient,
			trackerStateUseCase: trackerStateUseCase,
			contractAddress:     contractAddr,
		}

		deployedBlk := uint64(365)
		currentBlk := uint64(1024)
		status := &domain.TrackerState{ChainId: domain.ChainId(chainId), ContractAddress: domain.Address(contractAddrStr), LastBlockProcessed: deployedBlk - 1}
		trackerStateUseCase.On("Get", mock.Anything, status.ToId()).Return(nil, query.ErrNotFound)
		trackerStateUseCase.On("Store", mock.Anything, status).Return(nil)
		ethClient.On("BlockNumber", mock.Anything).Return(currentBlk, nil)
		ethClient.On("CodeAt",
			mock.Anything,
			mock.AnythingOfType("common.Address"),
			mock.AnythingOfType("*big.Int"),
		).Return(
			codeAtFunc(deployedBlk),
			nil,
		)

		got, err := f.setupTrackerState(ctx)
		req.NoError(err)
		req.Equal(status, got)
	})
}

func codeAtFunc(deployedBlock uint64) func(context.Context, common.Address, *big.Int) []byte {
	return func(_ context.Context, _ common.Address, blk *big.Int) []byte {
		if blk.Uint64() >= deployedBlock {
			return []byte("1")
		}
		return []byte{}
	}
}
