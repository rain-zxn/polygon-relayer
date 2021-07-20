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

package context

import (
	"sync"

	"github.com/polynetwork/poly/core/types"
	poly_go_sdkp "github.com/polynetwork/polygon-relayer/poly_go_sdk"
	poly_go_sdk "github.com/polynetwork/poly-go-sdk"

	tcrypto "github.com/christianxiao/tendermint/crypto"
	rpcclient "github.com/christianxiao/tendermint/rpc/client"
	rpctypes "github.com/christianxiao/tendermint/rpc/core/types"

	"github.com/polynetwork/polygon-relayer/config"
	"github.com/polynetwork/polygon-relayer/cosmos-sdk/codec"
	ctypes "github.com/polynetwork/polygon-relayer/cosmos-sdk/types"
	"github.com/polynetwork/polygon-relayer/db"
	cosmos "github.com/polynetwork/polygon-relayer/types"

	cryptoamino "github.com/christianxiao/tendermint/crypto/encoding/amino"
)

type InfoType int

const (
	TyTx InfoType = iota
	TyHeader
	TyUpdateHeight
)

var (
	RCtx = &Ctx{}
)

func InitCtx(conf *config.TendermintConfig, db *db.BoltDB, poly *poly_go_sdkp.PolySdk) error {
	var (
		err error
	)

	RCtx.Conf = conf

	// channels
	RCtx.ToPoly = make(chan *CosmosInfo, ChanBufSize)

	// prepare COSMOS staff
	RCtx.CMRpcCli = rpcclient.NewHTTP(conf.CosmosRpcAddr, "/websocket")

	cdc := codec.New()
	cryptoamino.RegisterAmino(cdc)
	RCtx.CMCdc = cdc

	RCtx.Poly = poly
	if RCtx.PolyAcc, err = GetAccountByPassword(RCtx.Poly, conf.PolyWallet, []byte(conf.PolyWalletPwd)); err != nil {
		return err
	}

	RCtx.Db = db

	return nil
}

type Ctx struct {
	// configuration
	Conf *config.TendermintConfig

	// To transfer cross chain tx from listening to relaying
	ToCosmos chan *PolyInfo
	ToPoly   chan *CosmosInfo

	// Cosmos
	CMRpcCli *rpcclient.HTTP
	CMPrivk  tcrypto.PrivKey
	CMAcc    ctypes.AccAddress
	CMSeq    *CosmosSeq
	CMAccNum uint64

	CMGas uint64
	CMCdc *codec.Codec

	// Poly chain
	Poly    *poly_go_sdkp.PolySdk
	PolyAcc *poly_go_sdk.Account

	// DB
	Db *db.BoltDB
}

type PolyInfo struct {
	// type 0 means only tx; type 2 means header and tx; type 1 means only header;
	Type InfoType

	// to update height of Poly on COSMOS
	Height uint32

	// tx part
	Tx *PolyTx

	// header part
	Hdr *types.Header

	// proof of header which is not during current epoch
	HeaderProof string

	// any header in current epoch can be trust anchor
	EpochAnchor string
}

type PolyTx struct {
	Height      uint32
	Proof       string
	TxHash      string
	IsEpoch     bool
	CCID        []byte
	FromChainId uint64
}

type CosmosInfo struct {
	// type 1 means header and tx; type 2 means only header;
	Type InfoType

	// to update height of chain
	Height int64

	// tx part
	Tx *CosmosTx

	// header part
	Hdrs []*cosmos.CosmosHeader
}

type CosmosTx struct {
	Tx          *rpctypes.ResultTx
	ProofHeight int64
	Proof       []byte
	PVal        []byte
}

type CosmosSeq struct {
	lock sync.Mutex
	val  uint64
}

func (seq *CosmosSeq) GetAndAdd() uint64 {
	seq.lock.Lock()
	defer func() {
		seq.val += 1
		seq.lock.Unlock()
	}()
	return seq.val
}
