module github.com/astriaorg/messenger-rollup

go 1.22.7

toolchain go1.23.2

require (
	buf.build/gen/go/astria/composer-apis/grpc/go v1.5.1-20241017141511-097580595f4b.1
	buf.build/gen/go/astria/composer-apis/protocolbuffers/go v1.35.1-20241017141511-097580595f4b.1
	buf.build/gen/go/astria/execution-apis/grpc/go v1.5.1-20241017141511-7e4bcc0ebba5.1
	buf.build/gen/go/astria/execution-apis/protocolbuffers/go v1.35.1-20241017141511-7e4bcc0ebba5.1
	github.com/cometbft/cometbft v0.38.13
	github.com/gorilla/mux v1.8.1
	github.com/gorilla/websocket v1.5.3
	github.com/rs/cors v1.11.1
	github.com/sethvargo/go-envconfig v1.0.0
	github.com/sirupsen/logrus v1.9.3
	google.golang.org/grpc v1.67.1
	google.golang.org/protobuf v1.35.1
)

require github.com/btcsuite/btcd/btcutil v1.1.6 // indirect

require (
	buf.build/gen/go/astria/primitives/protocolbuffers/go v1.35.1-20240911152449-eeebd3decdce.1
	buf.build/gen/go/astria/protocol-apis/protocolbuffers/go v1.35.1-20241025172303-24fb46273635.1
	buf.build/gen/go/astria/sequencerblock-apis/protocolbuffers/go v1.35.1-20241017141511-71aab1871615.1 // indirect
	github.com/DataDog/zstd v1.4.5 // indirect
	github.com/astriaorg/astria-cli-go/modules/bech32m v0.0.0-20241017220636-c58fd6130c6e
	github.com/astriaorg/astria-cli-go/modules/go-sequencer-client v0.0.0-20241017220636-c58fd6130c6e
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/cockroachdb/errors v1.11.3 // indirect
	github.com/cockroachdb/fifo v0.0.0-20240606204812-0bbfbd93a7ce // indirect
	github.com/cockroachdb/logtags v0.0.0-20230118201751-21c54148d20b // indirect
	github.com/cockroachdb/pebble v1.1.1 // indirect
	github.com/cockroachdb/redact v1.1.5 // indirect
	github.com/cockroachdb/tokenbucket v0.0.0-20230807174530-cc333fc44b06 // indirect
	github.com/cometbft/cometbft-db v0.14.1 // indirect
	github.com/cosmos/gogoproto v1.7.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.0.1 // indirect
	github.com/dgraph-io/badger/v4 v4.2.0 // indirect
	github.com/dgraph-io/ristretto v0.1.1 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/getsentry/sentry-go v0.27.0 // indirect
	github.com/go-kit/kit v0.13.0 // indirect
	github.com/go-kit/log v0.2.1 // indirect
	github.com/go-logfmt/logfmt v0.6.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/glog v1.2.2 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/btree v1.1.3 // indirect
	github.com/google/flatbuffers v1.12.1 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/jmhodges/levigo v1.0.0 // indirect
	github.com/klauspost/compress v1.17.9 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/linxGnu/grocksdb v1.8.14 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/oasisprotocol/curve25519-voi v0.0.0-20220708102147-0a8a51822cae // indirect
	github.com/petermattis/goid v0.0.0-20240813172612-4fcff4a6cae7 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/prometheus/client_golang v1.20.4 // indirect
	github.com/prometheus/client_model v0.6.1 // indirect
	github.com/prometheus/common v0.59.1 // indirect
	github.com/prometheus/procfs v0.15.1 // indirect
	github.com/rcrowley/go-metrics v0.0.0-20201227073835-cf1acfcdf475 // indirect
	github.com/rogpeppe/go-internal v1.11.0 // indirect
	github.com/sasha-s/go-deadlock v0.3.5 // indirect
	github.com/stretchr/testify v1.9.0 // indirect
	github.com/syndtr/goleveldb v1.0.1-0.20210819022825-2ae1ddf74ef7 // indirect
	go.etcd.io/bbolt v1.4.0-alpha.0.0.20240404170359-43604f3112c5 // indirect
	go.opencensus.io v0.24.0 // indirect
	golang.org/x/crypto v0.27.0 // indirect
	golang.org/x/exp v0.0.0-20231110203233-9a3e6036ecaa // indirect
	golang.org/x/net v0.29.0 // indirect
	golang.org/x/sys v0.25.0 // indirect
	golang.org/x/text v0.18.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20240814211410-ddb44dafa142 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240814211410-ddb44dafa142 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
