package messenger

import (
	"encoding/json"
	"errors"
	"time"

	astriaPb "buf.build/gen/go/astria/astria/protocolbuffers/go/astria/execution/v1alpha2"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Transaction struct {
	Sender   string `json:"sender"`
	Message  string `json:"message"`
	Priority uint32 `json:"priority"`
}

type Block struct {
	parent_hash []byte
	hash        []byte
	height      uint32
	timestamp   time.Time
	txs         []Transaction
}

func NewBlock(height uint32, txs []Transaction, timestamp time.Time) Block {
	return Block{
		parent_hash: []byte{0x0},
		hash:        []byte{0x0},
		height:      height,
		txs:         txs,
		timestamp:   timestamp,
	}
}

func (b *Block) ToPb() (*astriaPb.Block, error) {
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
		Hash:            b.hash,
		ParentBlockHash: b.parent_hash,
		Timestamp:       timestamppb.New(b.timestamp),
	}, nil
}

func GenesisBlock() Block {
	return Block{
		parent_hash: []byte{0x0},
		hash:        []byte{0x0},
		height:      0,
		timestamp:   time.Now(),
		txs:         []Transaction{},
	}
}

type Messenger struct {
	Blocks []Block
}

func NewMessenger() *Messenger {
	return &Messenger{
		Blocks: []Block{GenesisBlock()},
	}
}

func (m *Messenger) GetSingleBlock(height uint32) (*Block, error) {
	if height > uint32(len(m.Blocks)) {
		return nil, errors.New("block not found")
	}
	return &m.Blocks[height], nil
}

func (m *Messenger) GetCurrentBlock() (*Block, error) {
	return m.GetSingleBlock(uint32(len(m.Blocks) - 1))
}

func (m *Messenger) Height() uint32 {
	return uint32(len(m.Blocks))
}
