package messenger

import (
	"encoding/json"
	"errors"
	"time"

	astriaPb "buf.build/gen/go/astria/execution-apis/protocolbuffers/go/astria/execution/v1alpha2"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Transaction represents a transaction in the blockchain.
type Transaction struct {
	Sender   string `json:"sender"`
	Message  string `json:"message"`
	Priority uint32 `json:"priority"`
}

// Block represents a block in the blockchain.
type Block struct {
	parent_hash [32]byte
	hash        [32]byte
	height      uint32
	timestamp   time.Time
	txs         []Transaction
}

// NewBlock creates a new block with the given height, transactions, and timestamp.
func NewBlock(height uint32, txs []Transaction, timestamp time.Time) Block {
	return Block{
		parent_hash: [32]byte{0x0},
		hash:        [32]byte{0x0},
		height:      height,
		txs:         txs,
		timestamp:   timestamp,
	}
}

// ToPb converts the block to a protobuf message.
func (b *Block) ToPb() (*astriaPb.Block, error) {
	println("converting block to protobuff")
	txs := [][]byte{}
	for _, tx := range b.txs {
		if bytes, err := json.Marshal(tx); err != nil {
			txs = append(txs, bytes)
		} else {
			return nil, errors.New("failed to marshal transaction into bytes")
		}
	}

	return &astriaPb.Block{
		Number:          b.height,
		Hash:            b.hash[:],
		ParentBlockHash: b.parent_hash[:],
		Timestamp:       timestamppb.New(b.timestamp),
	}, nil
}

// GenesisBlock creates the genesis block.
func GenesisBlock() Block {
	return Block{
		parent_hash: [32]byte{0x00000000},
		hash:        [32]byte{0x00000000},
		height:      0,
		timestamp:   time.Now(),
		txs:         []Transaction{},
	}
}

// Messenger is a struct that manages the blocks in the blockchain.
type Messenger struct {
	Blocks []Block
}

// NewMessenger creates a new Messenger with a genesis block.
func NewMessenger() *Messenger {
	return &Messenger{
		Blocks: []Block{GenesisBlock()},
	}
}

// GetSingleBlock retrieves a block by its height.
func (m *Messenger) GetSingleBlock(height uint32) (*Block, error) {
	println("getting block at height", height)
	if height > uint32(len(m.Blocks)) {
		return nil, errors.New("block not found")
	}
	return &m.Blocks[height], nil
}

// GetCurrentBlock retrieves the current block.
func (m *Messenger) GetCurrentBlock() (*Block, error) {
	println("getting current block at height", len(m.Blocks)-1)
	return m.GetSingleBlock(uint32(len(m.Blocks) - 1))
}

// Height returns the height of the blockchain.
func (m *Messenger) Height() uint32 {
	println("getting height of blockchain", len(m.Blocks))
	return uint32(len(m.Blocks))
}
