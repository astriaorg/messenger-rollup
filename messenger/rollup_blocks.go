package messenger

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"time"

	astriaPb "buf.build/gen/go/astria/execution-apis/protocolbuffers/go/astria/execution/v1alpha2"
	"google.golang.org/protobuf/types/known/timestamppb"

	log "github.com/sirupsen/logrus"
)

func HashTxs(txs [][]byte) ([32]byte, error) {
	hash := sha256.Sum256(bytes.Join(txs, []byte{}))
	return hash, nil
}

type Block struct {
	ParentHash [32]byte
	Hash       [32]byte
	Height     uint32
	Timestamp  time.Time
	Txs        [][]byte
}

func NewBlock(parentHash []byte, height uint32, txs [][]byte, timestamp time.Time) Block {
	txHash, err := HashTxs(txs)
	if err != nil {
		panic(err)
	}

	return Block{
		ParentHash: [32]byte(parentHash),
		Hash:       txHash,
		Height:     height,
		Txs:        txs,
		Timestamp:  timestamp,
	}
}

func (b *Block) ToPb() (*astriaPb.Block, error) {
	return &astriaPb.Block{
		Number:          b.Height,
		Hash:            b.Hash[:],
		ParentBlockHash: b.ParentHash[:],
		Timestamp:       timestamppb.New(b.Timestamp),
	}, nil
}

// GenesisBlock creates the genesis block.
func GenesisBlock() Block {
	genesisTx := GenesisTransaction()

	genesisHash, err := HashTxs([][]byte{genesisTx})
	if err != nil {
		log.Errorf("error hashing genesis tx: %s\n", err)
		panic(err)
	}

	return Block{
		ParentHash: [32]byte{0x00000000},
		Hash:       genesisHash,
		Height:     0,
		Timestamp:  time.Now(),
		Txs: [][]byte{
			genesisTx,
		},
	}
}

// Messenger is a struct that manages the blocks in the blockchain.
type RollupBlocks struct {
	Blocks       []Block
	soft         uint32
	firm         uint32
	NewBlockChan chan Block
}

func NewRollupBlocks(newBlockChan chan Block) *RollupBlocks {
	return &RollupBlocks{
		Blocks:       []Block{GenesisBlock()},
		soft:         0,
		firm:         0,
		NewBlockChan: newBlockChan,
	}
}

// GetSingleBlock retrieves a block by its height, failing if the requested
// height is higher than the current height.
func (rb *RollupBlocks) GetSingleBlock(height uint32) (*Block, error) {
	log.Debugf("getting block at height %d\n", height)
	if height > uint32(len(rb.Blocks)) {
		return nil, errors.New("block not found")
	}
	return &rb.Blocks[height], nil
}

func (rb *RollupBlocks) GetSoftBlock() *Block {
	return &rb.Blocks[rb.soft]
}

func (rb *RollupBlocks) GetFirmBlock() *Block {
	return &rb.Blocks[rb.firm]
}

func (rb *RollupBlocks) GetLatestBlock() *Block {
	return &rb.Blocks[len(rb.Blocks)-1]
}

func (rb *RollupBlocks) Height() uint32 {
	return uint32(len(rb.Blocks))
}

func (rb *RollupBlocks) AddBlock(block Block) error {
	if rb.GetLatestBlock().Height > 0 && !bytes.Equal(block.ParentHash[:], rb.GetLatestBlock().Hash[:]) {
		return errors.New("invalid prev block hash")
	}
	rb.Blocks = append(rb.Blocks, block)
	select {
	case rb.NewBlockChan <- block:
	default:
	}
	return nil
}
