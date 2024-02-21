# Example Messenger Rollup
Set up Execution API dependencies (protobufs and gRPC):

```bash
go get buf.build/gen/go/astria/execution-apis/grpc/go
go get buf.build/gen/go/astria/execution-apis/protocolbuffers/go
```

Set up Sequencer Client and Tendermint RPC types dependencies:
```bash
go get "github.com/astriaorg/go-sequencer-client"
go get "github.com/cometbft/cometbft"
```
