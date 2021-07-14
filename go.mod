module github.com/polynetwork/polygon-relayer

go 1.14

require (
	github.com/boltdb/bolt v1.3.1
	github.com/btcsuite/btcd v0.21.0-beta
	github.com/cmars/basen v0.0.0-20150613233007-fe3947df716e // indirect
	github.com/cosmos/cosmos-sdk v0.38.4
	github.com/ethereum/go-ethereum v1.9.15
	github.com/ontio/ontology v1.11.0
	github.com/ontio/ontology-crypto v1.0.9
	github.com/pkg/errors v0.9.1
	github.com/polynetwork/eth-contracts v0.0.0-20200814062128-70f58e22b014
	github.com/polynetwork/poly v0.0.0-20200722075529-eea88acb37b2
	github.com/polynetwork/poly-go-sdk v0.0.0-20200729103825-af447ef53ef0
	github.com/stretchr/testify v1.7.0
	//github.com/tendermint/tendermint v0.33.3
	github.com/urfave/cli v1.22.4
	gopkg.in/yaml.v2 v2.2.8
	launchpad.net/gocheck v0.0.0-20140225173054-000000000087 // indirect
)

//replace github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.2-alpha.regen.4

replace github.com/polygon/tendermint => github.com/maticnetwork/tendermint v0.26.0-dev0.0.20200429080413-edc079e7d4c9

// replace github.com/cosmos/cosmos-sdk => github.com/maticnetwork/cosmos-sdk v0.37.5-0.20200503092858-55131f25dd9d
