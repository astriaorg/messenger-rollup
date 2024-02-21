package messenger

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"

	astriaGrpc "buf.build/gen/go/astria/astria/grpc/go/astria/execution/v1alpha2/executionv1alpha2grpc"
	astriaPb "buf.build/gen/go/astria/astria/protocolbuffers/go/astria/execution/v1alpha2"
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
	println("getSingleBlock called", "height", height)
	if height > s.m.Height() {
		return nil, errors.New("block not found")
	}

	block := s.m.Blocks[height]
	timestamp := timestamppb.New(block.timestamp)

	return &astriaPb.Block{
		Number:          height,
		Hash:            block.hash[:],
		ParentBlockHash: s.m.Blocks[height-1].hash[:],
		Timestamp:       timestamp,
	}, nil
}

// GetBlock retrieves a block by its identifier.
func (s *ExecutionServiceServerV1Alpha2) GetBlock(ctx context.Context, req *astriaPb.GetBlockRequest) (*astriaPb.Block, error) {
	println("GetBlock called", "request", req)
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
	if !bytes.Equal(req.PrevBlockHash, s.m.Blocks[len(s.m.Blocks)-1].hash[:]) {
		return nil, errors.New("invalid prev block hash")
	}
	txs := []Transaction{}
	for _, txBytes := range req.Transactions {
		tx := &Transaction{}
		json.Unmarshal(txBytes, *tx)
		txs = append(txs, *tx)
	}
	block := NewBlock(uint32(len(s.m.Blocks)), txs, req.Timestamp.AsTime())
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
	println("UpdateCommitmentState called", "request", req)
	println("UpdateCommitmentState completed", "request", req)
	return req.CommitmentState, nil
}

// getBlockFromIdentifier retrieves a block by its identifier.
func (s *ExecutionServiceServerV1Alpha2) getBlockFromIdentifier(identifier *astriaPb.BlockIdentifier) (*astriaPb.Block, error) {
	println("getBlockFromIdentifier called", "identifier", identifier)

	res := &astriaPb.Block{
		Number:          uint32(0),
		Hash:            []byte{0x0},
		ParentBlockHash: []byte{0x0},
	}
	println("getBlockFromIdentifier completed", "identifier", identifier, "response", res)
	return res, nil
}
