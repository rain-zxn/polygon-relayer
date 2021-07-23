module github.com/polynetwork/polygon-relayer

go 1.14

require (
	github.com/boltdb/bolt v1.3.1
	github.com/btcsuite/btcd v0.21.0-beta
	github.com/christianxiao/tendermint v0.26.0-polygon
	github.com/christianxiao/tm-db v0.2.0-polygon
	github.com/cmars/basen v0.0.0-20150613233007-fe3947df716e // indirect
	github.com/ethereum/go-ethereum v1.9.15
	github.com/maticnetwork/bor v0.0.0-20191204165821-bd9cd503a1b3
	github.com/ontio/ontology v1.11.0
	github.com/ontio/ontology-crypto v1.0.9
	github.com/pkg/errors v0.9.1
	github.com/polynetwork/eth-contracts v0.0.0-20200814062128-70f58e22b014
	github.com/polynetwork/poly v0.0.0-20200722075529-eea88acb37b2
	github.com/polynetwork/poly-go-sdk v0.0.0-20200729103825-af447ef53ef0
	github.com/stretchr/testify v1.7.0
	github.com/tendermint/go-amino v0.15.1
	github.com/urfave/cli v1.22.4
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9
	gopkg.in/yaml.v2 v2.2.8
	launchpad.net/gocheck v0.0.0-20140225173054-000000000087 // indirect
	poly_bridge_sdk v0.0.0-00010101000000-000000000000
)

replace poly_bridge_sdk => github.com/blockchain-develop/poly_bridge_sdk v0.0.0-20210327080022-0e6eb4b31700

//replace github.com/tendermint/tm-db => github.com/tendermint/tm-db v0.2.0

//replace github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.2-alpha.regen.4

// replace github.com/christianxiao/tendermint => github.com/tendermint/tendermint v0.26.0

// replace github.com/cosmos/cosmos-sdk => github.com/maticnetwork/cosmos-sdk v0.37.5-0.20200503092858-55131f25dd9d
