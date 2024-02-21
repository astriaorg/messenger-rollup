package messenger

import (
	"bytes"
	"crypto/sha256"
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

func HashTxs(txs []Transaction) (*[32]byte, error) {
	txBytes := [][]byte{}
	for _, tx := range txs {
		if bytes, err := json.Marshal(tx); err != nil {
			txBytes = append(txBytes, bytes)
		} else {
			return nil, errors.New("failed to marshal transaction into bytes")
		}
	}

	hash := sha256.Sum256(bytes.Join(txBytes, []byte{}))

	return &hash, nil
}

type Block struct {
	ParentHash [32]byte
	Hash       [32]byte
	Height     uint32
	Timestamp  time.Time
	Txs        []Transaction
}

func NewBlock(parentHash []byte, height uint32, txs []Transaction, timestamp time.Time) Block {
	txHash, err := HashTxs(txs)
	if err != nil {
		panic(err)
	}

	return Block{
		ParentHash: [32]byte(parentHash),
		Hash:       *txHash,
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
	return Block{
		ParentHash: [32]byte{0x00000000},
		Hash:       [32]byte{0x00000000},
		Height:     0,
		Timestamp:  time.Now(),
		Txs:        []Transaction{},
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

func (m *Messenger) GetLatestBlock() (*Block, error) {
	return m.GetSingleBlock(uint32(len(m.Blocks) - 1))
}

// Height returns the height of the blockchain.
func (m *Messenger) Height() uint32 {
	println("getting height of blockchain", len(m.Blocks))
	return uint32(len(m.Blocks))
}
