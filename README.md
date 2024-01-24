# Example Messenger Rollup
Set up Execution API dependencies (protobufs and gRPC):

```bash
go get buf.build/gen/go/astria/astria/grpc/go@latest

go get buf.build/gen/go/astria/astria/protocolbuffers/go@latest
```

Set up Sequencer Client and Tendermint RPC types dependencies:
```bash
go get "github.com/astriaorg/go-sequencer-client"
go get "github.com/cometbft/cometbft"
```


