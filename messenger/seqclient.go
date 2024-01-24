package messenger

import (
	"context"

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

func (sc *SequencerClient) SendMessage(sender string, message string, priority uint32) (*tendermintPb.ResultBroadcastTx, error) {

	tx := &astriaPb.UnsignedTransaction{
		Nonce: 1,
		Actions: []*astriaPb.Action{
			{
				Value: &astriaPb.Action_SequenceAction{
					SequenceAction: &astriaPb.SequenceAction{
						RollupId: []byte("test-chain"),
						Data:     []byte("test-data"),
					},
				},
			},
		},
	}

	signed, err := sc.signer.SignTransaction(tx)
	if err != nil {
		panic(err)
	}

	resp, err := sc.broadcastTxSync(signed)
	if err != nil {
		panic(err)
	}

	return resp, nil
}
