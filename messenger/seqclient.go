package messenger

import (
	"context"
	"encoding/json"
	"fmt"

	astriaPb "buf.build/gen/go/astria/astria/protocolbuffers/go/astria/sequencer/v1alpha1"
	client "github.com/astriaorg/go-sequencer-client/client"
	tendermintPb "github.com/cometbft/cometbft/rpc/core/types"

	log "github.com/sirupsen/logrus"
)

// SequencerClient is a client for interacting with the sequencer.
type SequencerClient struct {
	c        *client.Client
	signer   *client.Signer
	nonce    uint32
	rollupId []byte
}

// NewSequencerClient creates a new SequencerClient.
func NewSequencerClient(sequencerAddr string, rollupId []byte) *SequencerClient {
	log.Debug("creating new sequencer client")
	signer, err := client.GenerateSigner()
	if err != nil {
		panic(err)
	}

	// default tendermint RPC endpoint
	c, err := client.NewClient(sequencerAddr)
	if err != nil {
		panic(err)
	}

	return &SequencerClient{
		c:        c,
		signer:   signer,
		rollupId: rollupId,
	}
}

// broadcastTxSync broadcasts a transaction synchronously.
func (sc *SequencerClient) broadcastTxSync(tx *astriaPb.SignedTransaction) (*tendermintPb.ResultBroadcastTx, error) {
	log.Debug("broadcasting tx")
	return sc.c.BroadcastTxSync(context.Background(), tx)
}

// SendMessage sends a message as a transaction.
func (sc *SequencerClient) SendMessage(tx Transaction) (*tendermintPb.ResultBroadcastTx, error) {
	log.Debug("sending message")
	data, err := json.Marshal(tx)
	if err != nil {
		return nil, err
	}

	unsigned := &astriaPb.UnsignedTransaction{
		Nonce: sc.nonce,
		Actions: []*astriaPb.Action{
			{
				Value: &astriaPb.Action_SequenceAction{
					SequenceAction: &astriaPb.SequenceAction{
						RollupId: sc.rollupId,
						Data:     data,
					},
				},
			},
		},
	}

	signed, err := sc.signer.SignTransaction(unsigned)
	if err != nil {
		panic(err)
	}

	log.WithFields(log.Fields{
		"sender":  tx.Sender,
		"message": tx.Message,
	}).Debug("submitting tx to sequencer.")

	resp, err := sc.broadcastTxSync(signed)
	if err != nil {
		return nil, err
	}
	if resp.Code == 4 {
		// fetch new nonce
		newNonce, err := sc.c.GetNonce(context.Background(), sc.signer.Address())
		if err != nil {
			return nil, err
		}
		sc.nonce = newNonce

		// create new tx
		unsigned = &astriaPb.UnsignedTransaction{
			Nonce:   sc.nonce,
			Actions: unsigned.Actions,
		}
		signed, err = sc.signer.SignTransaction(unsigned)
		if err != nil {
			return nil, err
		}

		// submit new tx
		resp, err = sc.broadcastTxSync(signed)
		if err != nil {
			return nil, err
		}
		if resp.Code != 0 {
			return nil, fmt.Errorf("unexpected error code: %d", resp.Code)
		}
	} else if resp.Code != 0 {
		return nil, fmt.Errorf("unexpected error code: %d", resp.Code)
	}

	return resp, nil
}
