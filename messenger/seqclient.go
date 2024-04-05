package messenger

import (
	"buf.build/gen/go/astria/composer-apis/grpc/go/astria/composer/v1alpha1/composerv1alpha1grpc"
	"context"
	"crypto/ed25519"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	astriaPb "buf.build/gen/go/astria/astria/protocolbuffers/go/astria/sequencer/v1"
	astriaComposerPb "buf.build/gen/go/astria/composer-apis/protocolbuffers/go/astria/composer/v1alpha1"
	client "github.com/astriaorg/go-sequencer-client/client"
	tendermintPb "github.com/cometbft/cometbft/rpc/core/types"

	log "github.com/sirupsen/logrus"
)

// SequencerClient is a client for interacting with the sequencer.
type SequencerClient struct {
	c              *client.Client
	composerClient *grpc.ClientConn
	signer         *client.Signer
	nonce          uint32
	rollupId       []byte
}

// NewSequencerClient creates a new SequencerClient.
func NewSequencerClient(sequencerAddr string, composerAddr string, rollupId []byte, private ed25519.PrivateKey) *SequencerClient {
	signer := client.NewSigner(private)

	// default tendermint RPC endpoint
	c, err := client.NewClient(sequencerAddr)
	if err != nil {
		panic(err)
	}

	conn, err := grpc.Dial(composerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}

	return &SequencerClient{
		c:              c,
		composerClient: conn,
		signer:         signer,
		rollupId:       rollupId,
	}
}

// broadcastTxSync broadcasts a transaction synchronously.
func (sc *SequencerClient) broadcastTxSync(tx *astriaPb.SignedTransaction) (*tendermintPb.ResultBroadcastTx, error) {
	log.Debug("broadcasting tx")
	return sc.c.BroadcastTxSync(context.Background(), tx)
}

func (sc *SequencerClient) SendMessageViaComposer(tx []byte) error {
	log.Debug("broadcasting tx through composer!")

	grpcCollectorServiceClient := composerv1alpha1grpc.NewGrpcCollectorServiceClient(sc.composerClient)
	// if the request succeeds, then an empty response will be returned which can be ignored for now
	_, err := grpcCollectorServiceClient.SubmitRollupTransaction(context.Background(), &astriaComposerPb.SubmitRollupTransactionRequest{
		RollupId: sc.rollupId,
		Data:     tx,
	})
	if err != nil {
		return err
	}

	return nil
}

// SendMessage sends a message as a transaction.
func (sc *SequencerClient) SendMessage(tx []byte) (*tendermintPb.ResultBroadcastTx, error) {
	log.Debug("sending message")

	unsigned := &astriaPb.UnsignedTransaction{
		Nonce: sc.nonce,
		Actions: []*astriaPb.Action{
			{
				Value: &astriaPb.Action_SequenceAction{
					SequenceAction: &astriaPb.SequenceAction{
						RollupId: sc.rollupId,
						Data:     tx,
					},
				},
			},
		},
	}

	signed, err := sc.signer.SignTransaction(unsigned)
	if err != nil {
		panic(err)
	}

	log.Debugf("submitting tx to sequencer: %s.", tx)

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
	sc.nonce++

	return resp, nil
}
