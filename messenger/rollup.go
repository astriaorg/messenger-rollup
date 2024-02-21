package messenger

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"time"

	astriaPb "buf.build/gen/go/astria/execution-apis/protocolbuffers/go/astria/execution/v1alpha2"
	"google.golang.org/protobuf/types/known/timestamppb"

	log "github.com/sirupsen/logrus"
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
			return nil, err
		} else {
			txBytes = append(txBytes, bytes)
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
	helloTx := Transaction{
		Sender:  "astria",
		Message: "hello, world!",
	}

	helloHash, err := HashTxs([]Transaction{helloTx})
	if err != nil {
		log.Errorf("error hashing genesis tx: %s\n", err)
		panic(err)
	}

	return Block{
		ParentHash: [32]byte{0x00000000},
		Hash:       *helloHash,
		Height:     0,
		Timestamp:  time.Now(),
		Txs: []Transaction{
			helloTx,
		},
	}
}

// Messenger is a struct that manages the blocks in the blockchain.
type Messenger struct {
	Blocks []Block
	soft   uint32
	firm   uint32
}

func NewMessenger() *Messenger {
	return &Messenger{
		Blocks: []Block{GenesisBlock()},
		soft:   0,
		firm:   0,
	}
}

// GetSingleBlock retrieves a block by its height, failing if the requested
// height is higher than the current height.
func (m *Messenger) GetSingleBlock(height uint32) (*Block, error) {
	log.Debugf("getting block at height %d\n", height)
	if height > uint32(len(m.Blocks)) {
		return nil, errors.New("block not found")
	}
	return &m.Blocks[height], nil
}

func (m *Messenger) GetSoftBlock() *Block {
	return &m.Blocks[m.soft]
}

func (m *Messenger) GetFirmBlock() *Block {
	return &m.Blocks[m.firm]
}

func (m *Messenger) GetLatestBlock() *Block {
	return &m.Blocks[len(m.Blocks)-1]
}

func (m *Messenger) Height() uint32 {
	return uint32(len(m.Blocks))
}
