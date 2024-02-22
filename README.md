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


## Helpful things
```bash
curl -kv localhost:8080/message
   -H "Accept: application/json" -H "Content-Type: application/json" \
   --data '{"sender":"1c0c490f1b5528d8173c5de46d131160e4b2c0c3","message":"hello my friends","priority":1}'
   
curl -kv localhost:8080/block/1
```

