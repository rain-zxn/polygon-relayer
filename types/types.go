package types

import (
	"github.com/christianxiao/tendermint/types"

	ethtypes "github.com/ethereum/go-ethereum/core/types"

	"github.com/christianxiao/tendermint/crypto/merkle"

	"github.com/ethereum/go-ethereum/common"

	"golang.org/x/crypto/sha3"

	"github.com/ethereum/go-ethereum/rlp"
)

// json marshal
type HeaderWithOptionalProof struct {
	Header ethtypes.Header
	Proof  []byte // json var proof CosmosProof
}

type CosmosProof struct {
	Value  CosmosProofValue
	Proof  merkle.Proof
	Header CosmosHeader
}

type CosmosHeader struct {
	Header  types.Header
	Commit  *types.Commit
	Valsets []*types.Validator
}

type CosmosProofValue struct {
	Kp    string
	Value []byte
}

func (h *HeaderWithOptionalProof) Hash() common.Hash {
	return rlpHash(h)
}

func rlpHash(x interface{}) (h common.Hash) {
	hw := sha3.NewLegacyKeccak256()
	rlp.Encode(hw, x)
	hw.Sum(h[:0])
	return h
}