package messenger

import (
	"context"
	"crypto/ed25519"

	"buf.build/gen/go/astria/composer-apis/grpc/go/astria/composer/v1alpha1/composerv1alpha1grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	astriaComposerPb "buf.build/gen/go/astria/composer-apis/protocolbuffers/go/astria/composer/v1alpha1"
	client "github.com/astriaorg/astria-cli-go/modules/go-sequencer-client/client"

	log "github.com/sirupsen/logrus"
)

// SequencerClient is a client for interacting with the sequencer.
type SequencerClient struct {
	c              *client.Client
	composerClient *grpc.ClientConn
	signer         *client.Signer
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
