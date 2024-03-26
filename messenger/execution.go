package messenger

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"

	log "github.com/sirupsen/logrus"

	astriaGrpc "buf.build/gen/go/astria/execution-apis/grpc/go/astria/execution/v1alpha2/executionv1alpha2grpc"
	astriaPb "buf.build/gen/go/astria/execution-apis/protocolbuffers/go/astria/execution/v1alpha2"
)

// ExecutionServiceServerV1Alpha2 is a server that implements the ExecutionServiceServer interface.
type ExecutionServiceServerV1Alpha2 struct {
	astriaGrpc.UnimplementedExecutionServiceServer
	rollupBlocks *RollupBlocks
	rollupID     []byte
}

// NewExecutionServiceServerV1Alpha2 creates a new ExecutionServiceServerV1Alpha2.
func NewExecutionServiceServerV1Alpha2(rollupBlocks *RollupBlocks, rollupID []byte) *ExecutionServiceServerV1Alpha2 {
	return &ExecutionServiceServerV1Alpha2{
		rollupBlocks: rollupBlocks,
		rollupID:     rollupID,
	}
}

func (s *ExecutionServiceServerV1Alpha2) GetGenesisInfo(ctx context.Context, req *astriaPb.GetGenesisInfoRequest) (*astriaPb.GenesisInfo, error) {
	log.Debug("GetGenesisInfo called")
	res := &astriaPb.GenesisInfo{
		RollupId:                    s.rollupID,
		SequencerGenesisBlockHeight: uint32(1),
		CelestiaBaseBlockHeight:     uint32(1),
		CelestiaBlockVariance:       uint32(1),
	}
	log.WithFields(log.Fields{
		"rollupId": hex.EncodeToString(res.RollupId),
	}).Debug("GetGenesisInfo completed")
	return res, nil
}

// GetBlock retrieves a block by its identifier.
func (s *ExecutionServiceServerV1Alpha2) GetBlock(ctx context.Context, req *astriaPb.GetBlockRequest) (*astriaPb.Block, error) {
	log.WithField(
		"identifier", req.Identifier,
	).Debug("GetBlock called")
	switch req.Identifier.Identifier.(type) {
	case *astriaPb.BlockIdentifier_BlockNumber:
		block, err := s.rollupBlocks.GetSingleBlock(uint32(req.Identifier.GetBlockNumber()))
		if err != nil {
			return nil, err
		}
		blockPb, err := block.ToPb()
		if err != nil {
			return nil, err
		}

		log.WithField(
			"blockHash", hex.EncodeToString(block.Hash[:]),
		).Debugf("GetBlock completed with response: %v\n", block)
		return blockPb, nil
	default:
		log.Debugf("GetBlock completed with error: invalid identifier: %v\n", req.Identifier)
		return nil, errors.New("invalid identifier")
	}
}

// BatchGetBlocks retrieves multiple blocks by their identifiers.
func (s *ExecutionServiceServerV1Alpha2) BatchGetBlocks(ctx context.Context, req *astriaPb.BatchGetBlocksRequest) (*astriaPb.BatchGetBlocksResponse, error) {
	log.WithField(
		"identifiers", req.Identifiers,
	).Debug("BatchGetBlocks called")
	res := &astriaPb.BatchGetBlocksResponse{
		Blocks: []*astriaPb.Block{},
	}
	for _, id := range req.Identifiers {
		switch id.Identifier.(type) {
		case *astriaPb.BlockIdentifier_BlockNumber:
			height := uint32(id.GetBlockNumber())
			block, err := s.rollupBlocks.GetSingleBlock(height)
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

	log.Debugf("BatchGetBlocks completed with response: %v\n", res)
	return res, nil
}

// ExecuteBlock executes a block and adds it to the blockchain.
func (s *ExecutionServiceServerV1Alpha2) ExecuteBlock(ctx context.Context, req *astriaPb.ExecuteBlockRequest) (*astriaPb.Block, error) {
	log.WithFields(
		log.Fields{
			"prevBlockHash": hex.EncodeToString(req.PrevBlockHash),
			"txCount":       len(req.Transactions),
		},
	).Debugf("ExecuteBlock called")
	block := NewBlock(req.PrevBlockHash, uint32(len(s.rollupBlocks.Blocks)), req.Transactions, req.Timestamp.AsTime())
	err := s.rollupBlocks.AddBlock(block)
	if err != nil {
		return nil, err
	}

	blockPb, err := block.ToPb()
	if err != nil {
		return nil, errors.New("failed to convert block to protobuf")
	}

	log.WithField("blockHash", hex.EncodeToString(block.Hash[:])).Debugf("ExecuteBlock completed")
	return blockPb, nil
}

// GetCommitmentState retrieves the current commitment state of the blockchain.
func (s *ExecutionServiceServerV1Alpha2) GetCommitmentState(ctx context.Context, req *astriaPb.GetCommitmentStateRequest) (*astriaPb.CommitmentState, error) {
	log.Debug("GetCommitmentState called")
	soft, err := s.rollupBlocks.GetSoftBlock().ToPb()
	if err != nil {
		return nil, err
	}
	firm, err := s.rollupBlocks.GetFirmBlock().ToPb()
	if err != nil {
		return nil, err
	}

	res := &astriaPb.CommitmentState{
		Soft: soft,
		Firm: firm,
	}

	log.WithFields(
		log.Fields{
			"soft": soft.Number,
			"firm": firm.Number,
		},
	).Debugf("GetCommitmentState completed")
	return res, nil
}

// UpdateCommitmentState updates the commitment state of the blockchain.
func (s *ExecutionServiceServerV1Alpha2) UpdateCommitmentState(ctx context.Context, req *astriaPb.UpdateCommitmentStateRequest) (*astriaPb.CommitmentState, error) {
	log.WithFields(
		log.Fields{
			"soft":     req.CommitmentState.Soft.Number,
			"softHash": hex.EncodeToString(req.CommitmentState.Soft.Hash),
			"firm":     req.CommitmentState.Firm.Number,
			"firmHash": hex.EncodeToString(req.CommitmentState.Firm.Hash),
		},
	).Debugf("UpdateCommitmentState called")
	softHeight := req.CommitmentState.Soft.Number
	firmHeight := req.CommitmentState.Firm.Number

	// get the actual soft and firm blocks
	firmBlock, err := s.rollupBlocks.GetSingleBlock(firmHeight)
	if err != nil {
		return nil, err
	}
	softBlock, err := s.rollupBlocks.GetSingleBlock(softHeight)
	if err != nil {
		return nil, err
	}

	// compare actual blocks to commit state
	if !bytes.Equal(softBlock.Hash[:], req.CommitmentState.Soft.Hash) {
		return nil, errors.New("soft block hash mismatch")
	}
	if !bytes.Equal(firmBlock.Hash[:], req.CommitmentState.Firm.Hash) {
		return nil, errors.New("firm block hash mismatch")
	}

	// update the commitment state
	s.rollupBlocks.soft = softHeight
	s.rollupBlocks.firm = firmHeight

	log.WithFields(
		log.Fields{
			"soft": s.rollupBlocks.soft,
			"firm": s.rollupBlocks.firm,
		},
	).Debugf("UpdateCommitmentState completed")
	return req.CommitmentState, nil
}
