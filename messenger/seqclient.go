package messenger

import (
	"context"
	"crypto/ed25519"
	"fmt"

	"buf.build/gen/go/astria/composer-apis/grpc/go/astria/composer/v1/composerv1grpc"
	primitivev1 "buf.build/gen/go/astria/primitives/protocolbuffers/go/astria/primitive/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	astriaComposerPb "buf.build/gen/go/astria/composer-apis/protocolbuffers/go/astria/composer/v1"
	astriaPb "buf.build/gen/go/astria/protocol-apis/protocolbuffers/go/astria/protocol/transaction/v1"
	bech32m "github.com/astriaorg/astria-cli-go/modules/bech32m"
	client "github.com/astriaorg/astria-cli-go/modules/go-sequencer-client/client"
	tendermintPb "github.com/cometbft/cometbft/rpc/core/types"

	log "github.com/sirupsen/logrus"
)

// SequencerClient is a client for interacting with the sequencer.
type SequencerClient struct {
	c              *client.Client
	composerClient *grpc.ClientConn
	signer         *client.Signer
	nonce          uint32
	rollupId       primitivev1.RollupId
}

// NewSequencerClient creates a new SequencerClient.
func NewSequencerClient(sequencerAddr string, composerAddr string, rollupId primitivev1.RollupId, private ed25519.PrivateKey) *SequencerClient {
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
func (sc *SequencerClient) broadcastTxSync(tx *astriaPb.Transaction) (*tendermintPb.ResultBroadcastTx, error) {
	log.Debug("broadcasting tx")
	return sc.c.BroadcastTxSync(context.Background(), tx)
}

func (sc *SequencerClient) SendMessageViaComposer(tx []byte) error {
	log.Debug("broadcasting tx through composer!")

	grpcCollectorServiceClient := composerv1grpc.NewGrpcCollectorServiceClient(sc.composerClient)
	// if the request succeeds, then an empty response will be returned which can be ignored for now
	_, err := grpcCollectorServiceClient.SubmitRollupTransaction(context.Background(), &astriaComposerPb.SubmitRollupTransactionRequest{
		RollupId: &sc.rollupId,
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

	unsigned := &astriaPb.TransactionBody{
		Params: &astriaPb.TransactionParams{
			Nonce:   sc.nonce,
			ChainId: "astria",
		},
		Actions: []*astriaPb.Action{{Value: &astriaPb.Action_RollupDataSubmission{
			RollupDataSubmission: &astriaPb.RollupDataSubmission{
				RollupId: &sc.rollupId,
				Data:     tx,
				FeeAsset: "nria",
			},
		},
		}}}

	signed, err := sc.signer.SignTransaction(unsigned)
	if err != nil {
		panic(err)
	}

	log.Debugf("submitting tx to sequencer: %s.", tx)
	address, err := bech32m.EncodeFromBytes("astria", sc.signer.Address())
	if err != nil {
		return nil, err
	}
	resp, err := sc.broadcastTxSync(signed)
	if err != nil {
		return nil, err
	}
	if resp.Code == 4 {
		// fetch new nonce
		newNonce, err := sc.c.GetNonce(context.Background(), address.String())
		if err != nil {
			return nil, err
		}
		sc.nonce = newNonce

		// create new tx
		unsigned = &astriaPb.TransactionBody{
			Params: &astriaPb.TransactionParams{
				Nonce:   sc.nonce,
				ChainId: "astria",
			},
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
