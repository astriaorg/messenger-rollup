package messenger

import (
	"context"
	"encoding/json"
	"fmt"

	astriaPb "buf.build/gen/go/astria/astria/protocolbuffers/go/astria/sequencer/v1alpha1"
	client "github.com/astriaorg/go-sequencer-client/client"
	tendermintPb "github.com/cometbft/cometbft/rpc/core/types"
)

type SequencerClient struct {
	c      *client.Client
	signer *client.Signer
}

func NewSequencerClient(sequencerAddr string) *SequencerClient {
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
		c:      c,
		signer: signer,
	}
}

func (sc *SequencerClient) broadcastTxSync(tx *astriaPb.SignedTransaction) (*tendermintPb.ResultBroadcastTx, error) {
	return sc.c.BroadcastTxSync(context.Background(), tx)
}

func (sc *SequencerClient) SendMessage(tx Transaction) (*tendermintPb.ResultBroadcastTx, error) {
	data, err := json.Marshal(tx)
	if err != nil {
		return nil, err
	}

	unsigned := &astriaPb.UnsignedTransaction{
		Nonce: 1,
		Actions: []*astriaPb.Action{
			{
				Value: &astriaPb.Action_SequenceAction{
					SequenceAction: &astriaPb.SequenceAction{
						RollupId: []byte("messenger-rollup"),
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

	fmt.Printf("submitting tx to sequencer. sender: %s, message: %s\n", tx.Sender, tx.Message)

	resp, err := sc.broadcastTxSync(signed)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
