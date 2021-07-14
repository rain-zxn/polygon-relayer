package types

import (
	"github.com/tendermint/tendermint/types"
)

type CosmosHeader struct {
	Header  types.Header
	Commit  *types.Commit
	Valsets []*types.Validator
}