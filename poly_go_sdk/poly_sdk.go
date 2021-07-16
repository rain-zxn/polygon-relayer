package poly_go_sdk

import (
	poly_go_sdk "github.com/polynetwork/poly-go-sdk"
	"github.com/polynetwork/polygon-relayer/log"

	"github.com/polynetwork/poly/common"
)

var Enable bool

type PolySdk struct {
	*poly_go_sdk.PolySdk
	Native NativeContract
}

type NativeContract struct {
	mcSdk *poly_go_sdk.PolySdk
	Hs    *HeaderSync
	Ccm   *poly_go_sdk.CrossChainManager
	Scm   *poly_go_sdk.SideChainManager
	Nm    *poly_go_sdk.NodeManager
	Rm    *poly_go_sdk.RelayerManager
}

func NewPolySdkp(sdk *poly_go_sdk.PolySdk, enable bool) *PolySdk {
	Enable = enable

	a := &PolySdk{PolySdk: sdk}
	b := &HeaderSync{a.PolySdk.Native.Hs}
	a.Native.Hs = b
	return a
}

type HeaderSync struct {
	*poly_go_sdk.HeaderSync
}

func (this *HeaderSync) SyncBlockHeader(chainId uint64, address common.Address, headers [][]byte, signer *poly_go_sdk.Account) (common.Uint256, error) {
	if !Enable {
		return this.HeaderSync.SyncBlockHeader(chainId, address, headers, signer)
	}

	log.Debugf("PolySdk proxy - SyncBlockHeader, chainId: %s, address: %s, headers: %s, signer: %s",
		chainId, address, headers, signer)
	return common.UINT256_EMPTY, nil
}
