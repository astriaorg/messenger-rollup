package messenger

import (
	"errors"
	"time"

	astriaPb "buf.build/gen/go/astria/astria/protocolbuffers/go/astria/execution/v1alpha2"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Transaction struct {
	message string
}

func NewTransaction(b []byte) Transaction {
	return Transaction{
		message: string(b),
	}
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

func (b *Block) toPb() *astriaPb.Block {
	txs := [][]byte{}
	for _, tx := range b.txs {
		txs = append(txs, []byte(tx.message))
	}

	return &astriaPb.Block{
		Number:          b.height,
		Hash:            b.hash,
		ParentBlockHash: b.parent_hash,
		Timestamp:       timestamppb.New(b.timestamp),
	}
}

type Messenger struct {
	Blocks []Block
}

func NewMessenger() *Messenger {
	return &Messenger{
		Blocks: []Block{},
	}
}

func (m *Messenger) GetSingleBlock(height uint32) (*Block, error) {
	if height >= uint32(len(m.Blocks)) {
		return nil, errors.New("block not found")
	}
	return &m.Blocks[height], nil
}

func (m *Messenger) GetCurrentBlock() (*Block, error) {
	return m.GetSingleBlock(uint32(len(m.Blocks) - 1))
}
