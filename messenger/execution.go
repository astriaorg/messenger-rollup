package messenger

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"

	log "github.com/sirupsen/logrus"

	astriaGrpc "buf.build/gen/go/astria/execution-apis/grpc/go/astria/execution/v1alpha2/executionv1alpha2grpc"
	astriaPb "buf.build/gen/go/astria/execution-apis/protocolbuffers/go/astria/execution/v1alpha2"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ExecutionServiceServerV1Alpha2 is a server that implements the ExecutionServiceServer interface.
type ExecutionServiceServerV1Alpha2 struct {
	astriaGrpc.UnimplementedExecutionServiceServer
	m *Messenger
}

// NewExecutionServiceServerV1Alpha2 creates a new ExecutionServiceServerV1Alpha2.
func NewExecutionServiceServerV1Alpha2(m *Messenger) *ExecutionServiceServerV1Alpha2 {
	return &ExecutionServiceServerV1Alpha2{
		m: m,
	}
}

// getSingleBlock retrieves a single block by its height.
func (s *ExecutionServiceServerV1Alpha2) getSingleBlock(height uint32) (*astriaPb.Block, error) {
	log.Println("getSingleBlock called", "height", height)
	if height > s.m.Height() {
		return nil, errors.New("block not found")
	}

	block := s.m.Blocks[height]
	timestamp := timestamppb.New(block.Timestamp)

	return &astriaPb.Block{
		Number:          height,
		Hash:            block.Hash[:],
		ParentBlockHash: s.m.Blocks[height-1].Hash[:],
		Timestamp:       timestamp,
	}, nil
}

// GetBlock retrieves a block by its identifier.
func (s *ExecutionServiceServerV1Alpha2) GetBlock(ctx context.Context, req *astriaPb.GetBlockRequest) (*astriaPb.Block, error) {
	log.Println("GetBlock called", "request", req)
	switch req.Identifier.Identifier.(type) {
	case *astriaPb.BlockIdentifier_BlockNumber:
		block, err := s.getSingleBlock(uint32(req.Identifier.GetBlockNumber()))
		if err != nil {
			return nil, err
		}
		return block, nil
	default:
		return nil, errors.New("invalid identifier")
	}
}

// BatchGetBlocks retrieves multiple blocks by their identifiers.
func (s *ExecutionServiceServerV1Alpha2) BatchGetBlocks(ctx context.Context, req *astriaPb.BatchGetBlocksRequest) (*astriaPb.BatchGetBlocksResponse, error) {
	res := &astriaPb.BatchGetBlocksResponse{
		Blocks: []*astriaPb.Block{},
	}
	for _, id := range req.Identifiers {
		switch id.Identifier.(type) {
		case *astriaPb.BlockIdentifier_BlockNumber:
			height := uint32(id.GetBlockNumber())
			block, err := s.m.GetSingleBlock(height)
			if err != nil {
				return nil, err
			}
			blockPb, err := block.ToPb()
			if err != nil {
				return nil, err
			}
			res.Blocks = append(res.Blocks, blockPb)
		}
	}
	return res, nil
}

// ExecuteBlock executes a block and adds it to the blockchain.
func (s *ExecutionServiceServerV1Alpha2) ExecuteBlock(ctx context.Context, req *astriaPb.ExecuteBlockRequest) (*astriaPb.Block, error) {
	if !bytes.Equal(req.PrevBlockHash, s.m.Blocks[len(s.m.Blocks)-1].Hash[:]) {
		return nil, errors.New("invalid prev block hash")
	}
	txs := []Transaction{}
	for _, txBytes := range req.Transactions {
		tx := &Transaction{}
		if err := json.Unmarshal(txBytes, tx); err != nil {
			return nil, errors.New("failed to unmarshal transaction")
		}
		txs = append(txs, *tx)
	}
	block := NewBlock(req.PrevBlockHash, uint32(len(s.m.Blocks)), txs, req.Timestamp.AsTime())
	s.m.Blocks = append(s.m.Blocks, block)

	blockPb, err := block.ToPb()
	if err != nil {
		return nil, errors.New("failed to convert block to protobuf")
	}
	return blockPb, nil
}

// GetCommitmentState retrieves the current commitment state of the blockchain.
func (s *ExecutionServiceServerV1Alpha2) GetCommitmentState(ctx context.Context, req *astriaPb.GetCommitmentStateRequest) (*astriaPb.CommitmentState, error) {
	soft, err := s.m.Blocks[len(s.m.Blocks)-1].ToPb()
	if err != nil {
		return nil, errors.New("failed to convert soft block to protobuf")
	}
	hard, err := s.m.Blocks[len(s.m.Blocks)-1].ToPb()
	if err != nil {
		return nil, errors.New("failed to convert hard block to protobuf")
	}
	res := &astriaPb.CommitmentState{
		Soft: soft,
		Firm: hard,
	}
	return res, nil
}

// UpdateCommitmentState updates the commitment state of the blockchain.
func (s *ExecutionServiceServerV1Alpha2) UpdateCommitmentState(ctx context.Context, req *astriaPb.UpdateCommitmentStateRequest) (*astriaPb.CommitmentState, error) {
	log.Println("UpdateCommitmentState called", "request", req)
	log.Println("UpdateCommitmentState completed", "request", req)
	return req.CommitmentState, nil
}

// getBlockFromIdentifier retrieves a block by its identifier.
func (s *ExecutionServiceServerV1Alpha2) getBlockFromIdentifier(identifier *astriaPb.BlockIdentifier) (*astriaPb.Block, error) {
	log.Println("getBlockFromIdentifier called", "identifier", identifier)

	res := &astriaPb.Block{
		Number:          uint32(0),
		Hash:            []byte{0x0},
		ParentBlockHash: []byte{0x0},
	}
	log.Println("getBlockFromIdentifier completed", "identifier", identifier, "response", res)
	return res, nil
}

func (s *ExecutionServiceServerV1Alpha2) GetGenesisInfo(ctx context.Context, req *astriaPb.GetGenesisInfoRequest) (*astriaPb.GenesisInfo, error) {
	log.Println("GetGenesisInfo called", "request", req)
	// FIXME - use envars/config
	rollupId := sha256.Sum256([]byte("messenger-rollup"))
	res := &astriaPb.GenesisInfo{
		RollupId:                    rollupId[:],
		SequencerGenesisBlockHeight: 1,
		CelestiaBaseBlockHeight:     0,
		CelestiaBlockVariance:       0,
	}
	log.Println("GetGenesisInfo completed", "request", req, "response", res)
	return res, nil
}
