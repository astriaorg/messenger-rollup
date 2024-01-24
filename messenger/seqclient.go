package messenger

// import (
// 	"context"

// 	proto "github.com/astriaorg/go-sequencer-client/"
// 	client "github.com/astriaorg/go-sequencer-client/client"
// )

// type SequencerClient struct {
// 	c      proto.SequencerClient
// 	signer client.Signer
// }

// func NewSequencerClient(sequencerAddr string) *SequencerClient {
// 	signer, err := client.GenerateSigner()
// 	if err != nil {
// 		panic(err)
// 	}

// 	// default tendermint RPC endpoint
// 	c, err := client.NewClient(sequencerAddr)
// 	if err != nil {
// 		panic(err)
// 	}

// 	return &SequencerClient{
// 		c:      c,
// 		signer: signer,
// 	}
// }

// func (c *SequencerClient) broadcastTxSync(tx *proto.SignedTransaction) (*proto.BroadcastTxSyncResponse, error) {
// 	return c.c.BroadcastTxSync(context.Background(), tx)
// }

// // func test() {

// // 	tx := &sqproto.UnsignedTransaction{
// // 		Nonce: 1,
// // 		Actions: []*sqproto.Action{
// // 			{
// // 				Value: &sqproto.Action_SequenceAction{
// // 					SequenceAction: &sqproto.SequenceAction{
// // 						ChainId: []byte("test-chain"),
// // 						Data:    []byte("test-data"),
// // 					},
// // 				},
// // 			},
// // 		},
// // 	}

// // 	signed, err := signer.SignTransaction(tx)
// // 	if err != nil {
// // 		panic(err)
// // 	}

// // 	resp, err := c.BroadcastTxSync(context.Background(), signed)
// // 	if err != nil {
// // 		panic(err)
// // 	}

// // 	fmt.Println(resp)
// // }
