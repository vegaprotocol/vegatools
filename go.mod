module code.vegaprotocol.io/vegatools

go 1.19

require (
	code.vegaprotocol.io/shared v0.0.0-20221010085458-55c50711135f
	code.vegaprotocol.io/vega v0.67.3-0.20230214165643-b434c494e7f3
	github.com/cosmos/iavl v0.19.4
	github.com/ethereum/go-ethereum v1.10.21
	github.com/gdamore/tcell/v2 v2.5.2
	github.com/gogo/protobuf v1.3.2
	github.com/prometheus-community/pro-bing v0.1.0
	github.com/shopspring/decimal v1.3.1
	github.com/spf13/cobra v1.6.0
	github.com/stretchr/testify v1.8.0
	github.com/syndtr/goleveldb v1.0.1-0.20220614013038-64ee5596c38a
	github.com/tendermint/tm-db v0.6.7
	go.nanomsg.org/mangos/v3 v3.4.2
	golang.org/x/crypto v0.4.0
	google.golang.org/grpc v1.50.1
	google.golang.org/protobuf v1.28.2-0.20220831092852-f930b1dc76e8
)

require (
	github.com/DataDog/zstd v1.4.1 // indirect
	github.com/StackExchange/wmi v1.2.1 // indirect
	github.com/btcsuite/btcd/btcec/v2 v2.2.1 // indirect
	github.com/cespare/xxhash v1.1.0 // indirect
	github.com/confio/ics23/go v0.7.0 // indirect
	github.com/cosmos/gorocksdb v1.2.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/deckarep/golang-set v1.8.0 // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.0.1 // indirect
	github.com/dgraph-io/badger/v2 v2.2007.2 // indirect
	github.com/dgraph-io/ristretto v0.0.3-0.20200630154024-f66de99634de // indirect
	github.com/dgryski/go-farm v0.0.0-20190423205320-6a90982ecee2 // indirect
	github.com/didip/tollbooth/v7 v7.0.1 // indirect
	github.com/dustin/go-humanize v1.0.0 // indirect
	github.com/gdamore/encoding v1.0.0 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/go-pkgz/expirable-cache v0.1.0 // indirect
	github.com/go-stack/stack v1.8.0 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/btree v1.0.0 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.9.0 // indirect
	github.com/holiman/uint256 v1.2.0 // indirect
	github.com/inconshreveable/mousetrap v1.0.1 // indirect
	github.com/jmhodges/levigo v1.0.0 // indirect
	github.com/lucasb-eyer/go-colorful v1.2.0 // indirect
	github.com/mattn/go-runewidth v0.0.13 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/rjeczalik/notify v0.9.1 // indirect
	github.com/satori/go.uuid v1.2.0 // indirect
	github.com/shirou/gopsutil v3.21.4-0.20210419000835-c7a38de76ee5+incompatible // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/tendermint/tendermint v0.34.24 // indirect
	github.com/tklauser/go-sysconf v0.3.10 // indirect
	github.com/tklauser/numcpus v0.4.0 // indirect
	go.etcd.io/bbolt v1.3.6 // indirect
	go.uber.org/atomic v1.10.0 // indirect
	go.uber.org/multierr v1.8.0 // indirect
	go.uber.org/zap v1.23.0 // indirect
	golang.org/x/net v0.8.0 // indirect
	golang.org/x/sync v0.1.0 // indirect
	golang.org/x/sys v0.6.0 // indirect
	golang.org/x/term v0.6.0 // indirect
	golang.org/x/text v0.8.0 // indirect
	google.golang.org/genproto v0.0.0-20221014213838-99cd37c6964a // indirect
	gopkg.in/natefinch/npipe.v2 v2.0.0-20160621034901-c1b8fa8bdcce // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace (
	github.com/jackc/pgx/v4 v4.14.1 => github.com/pscott31/pgx/v4 v4.16.2-0.20220531164027-bd666b84b61f
	github.com/shopspring/decimal => github.com/vegaprotocol/decimal v1.3.1-uint256
)
