/*
 * Copyright (C) 2020 The poly network Authors
 * This file is part of The poly network library.
 *
 * The  poly network  is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The  poly network  is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 * You should have received a copy of the GNU Lesser General Public License
 * along with The poly network .  If not, see <http://www.gnu.org/licenses/>.
 */

package poly_go_sdk

import (
	"bytes"
	"encoding/binary"

	poly_go_sdk "github.com/polynetwork/poly-go-sdk"
	"github.com/polynetwork/polygon-relayer/log"

	"github.com/polynetwork/poly/common"

	mhcomm "github.com/polynetwork/poly/native/service/header_sync/common"
	autils "github.com/polynetwork/poly/native/service/utils"
	polycosmos "github.com/polynetwork/polygon-relayer/poly/native/header_sync/cosmos"
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

func (this *PolySdk) GetStorage(contractAddress string, key []byte) ([]byte, error) {
	if !Enable {
		return this.PolySdk.GetStorage(contractAddress, key)
	}

	log.Infof("PolySdk proxy - GetStorage, contractAddress: %s, key: %s ",
		contractAddress, key)

	// find epoch struct
	if bytes.HasPrefix(key, []byte(mhcomm.EPOCH_SWITCH)) {
		c := &polycosmos.CosmosEpochSwitchInfo{}
		sink := &common.ZeroCopySink{}
		c.Serialization(sink)
	
		return sink.Bytes(), nil
	}	

	// findLastestHeight
	if autils.HeaderSyncContractAddress.ToHexString() == contractAddress {
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, 17022221)
		return b, nil
	}

	return []byte("0"), nil
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
