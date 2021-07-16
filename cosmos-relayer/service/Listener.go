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

package service

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/polynetwork/poly/common"
	mhcomm "github.com/polynetwork/poly/native/service/header_sync/common"
	cosmos "github.com/polynetwork/polygon-relayer/types"
	polycosmos "github.com/polynetwork/polygon-relayer/poly/native/header_sync/cosmos"
	"github.com/polynetwork/poly/native/service/utils"

	tdmt_types "github.com/christianxiao/tendermint/types"

	"github.com/polynetwork/polygon-relayer/cosmos-relayer/context"
	"github.com/polynetwork/polygon-relayer/log"
)

var (
	ctx = context.RCtx
)

// Start listen cosmos and Poly
func StartListen() {
	// go PolyListen()
	go CosmosListen()
}


// Cosmos listen service implementation. Check the blocks of COSMOS from height
// `left` to height `right`, commit the cross-chain txs and headers to prove txs
// to chain Poly. It execute once per `ctx.Conf.CosmosListenInterval` sec. And update
// height `left` `right` after execution for next round. This function will run
// as a go-routine.
func CosmosListen() {
	log.LogTender.Infof("Listerner.CosmosListen - start")

	left, tick, err := beforeCosmosListen()
	if err != nil {
		log.LogTender.Fatalf("[ListenCosmos] failed to get start height of Cosmos: %v", err)
		panic(err)
	}
	log.LogTender.Infof("[ListenCosmos] start listen Cosmos (start_height: %d, listen_interval: %d)", left+1,
		ctx.Conf.CosmosListenInterval)

	lastRight := left
	for {
		select {
		case <-tick.C:
			status, err := ctx.CMRpcCli.Status()
			switch {
			case err != nil:
				log.LogTender.Errorf("[ListenCosmos] failed to get height of COSMOS, retry after %d sec: %v",
					ctx.Conf.CosmosListenInterval, err)
				continue
			case status.SyncInfo.LatestBlockHeight-1 <= lastRight:
				continue
			}
			right := status.SyncInfo.LatestBlockHeight - 1
			hdr, err := getCosmosHdr(right)
			if err != nil {
				log.LogTender.Errorf("[ListenCosmos] failed to get %d header to get proof, retry after %d sec: %v",
					right, ctx.Conf.CosmosListenInterval, err)
				continue
			}
			if !bytes.Equal(hdr.Header.ValidatorsHash, hdr.Header.NextValidatorsHash) {
				log.LogTender.Debugf("[ListenCosmos] header at %d is epoch switching point, so continue loop", hdr.Header.Height)
				lastRight = right
				continue
			}

			// let first element of infoArr be the info for epoch switching headers.
			infoArr := make([]*context.CosmosInfo, 1)
			infoArr[0] = &context.CosmosInfo{
				Type: context.TyHeader,
				Hdrs: make([]*cosmos.CosmosHeader, 0),
			}

			// TODO: this for loop will cost much time if two much headers between left and right
			for h := left + 1; h <= right; h++ {
				infoArrTemp, err := checkCosmosHeight(h, hdr, infoArr, &right)
				if err != nil {
					// If error happen, we should check this height again.
					h--
					if strings.Contains(err.Error(), context.RightHeightUpdate) {
						// Can't get proof from height `right-1`, update right to the latest.
						log.LogTender.Errorf("[ListenCosmos] checkCosmosHeight height: %s right: %s error: %s", h, right, err.Error())
						continue
					}
					// some error happen, could be some network error or COSMOS full node error.
					log.LogTender.Errorf("[ListenCosmos] failed to fetch info from COSMOS, retry after 10 sec: %v", err)
					context.SleepSecs(10)
					continue
				}
				infoArr = infoArrTemp
			}

			for i, v := range infoArr {
				if i == 0 && len(v.Hdrs) == 0 {
					continue
				}
				ctx.ToPoly <- v
			}
			
			ctx.ToPoly <- &context.CosmosInfo{
				Type:   context.TyUpdateHeight,
				Height: right,
			}
			lastRight = right
			left = right
		}
	}
}

// Prepare start height and ticker when init service
func beforeCosmosListen() (int64, *time.Ticker, error) {
	val, err := ctx.Poly.GetStorage(utils.HeaderSyncContractAddress.ToHexString(),
		append([]byte(mhcomm.EPOCH_SWITCH), utils.GetUint64Bytes(ctx.Conf.SideChainId)...))
	if err != nil {
		return 0, nil, err
	}
	info := &polycosmos.CosmosEpochSwitchInfo{}
	if err = info.Deserialization(common.NewZeroCopySource(val)); err != nil {
		return 0, nil, err
	}
	currHeight := info.Height
	if currHeight > 1 {
		currHeight--
	}
	log.LogTender.Infof("beforeCosmosListen, ( cosmos height on Poly: %d )", currHeight)

	if dbh := ctx.Db.GetCosmosHeight(); dbh > currHeight {
		log.LogTender.Infof("beforeCosmosListen, ( cosmos height in DB: %d )", dbh)
		currHeight = dbh
	}
	if ctx.Conf.CosmosStartHeight != 0 {
		currHeight = ctx.Conf.CosmosStartHeight
	}
	return currHeight, time.NewTicker(time.Duration(ctx.Conf.CosmosListenInterval) * time.Second), nil
}

// Fetch header at h and check tx at h-1.
//
// Put header to `hdrArr` and txs to `txArr`. Get proof from height `heightToGetProof`.
// `headersToRelay` record all hdrs need to relay. When need to update new height to
// get proof, relayer update `rightPtr` and return.
func checkCosmosHeight(h int64, hdrToVerifyProof *cosmos.CosmosHeader, infoArr []*context.CosmosInfo, rightPtr *int64) ([]*context.CosmosInfo, error) {
	rc, err := ctx.CMRpcCli.Commit(&h)
	if err != nil {
		return infoArr, err
	}
	if !bytes.Equal(rc.Header.ValidatorsHash, rc.Header.NextValidatorsHash) {
		vSet, err := getValidators(h)
		if err != nil {
			return infoArr, err
		}
		hdr := &cosmos.CosmosHeader{
			Header:  *rc.Header,
			Commit:  rc.Commit,
			Valsets: vSet,
		}
		val, _ := ctx.Poly.GetStorage(utils.CrossChainManagerContractAddress.ToHexString(),
			append(append([]byte(mhcomm.EPOCH_SWITCH), utils.GetUint64Bytes(ctx.Conf.SideChainId)...),
				utils.GetUint64Bytes(uint64(h))...))
		// check if this header is not committed on Poly
		if val == nil || len(val) == 0 {
			infoArr[0].Hdrs = append(infoArr[0].Hdrs, hdr)
		}
	}

	return infoArr, nil
}

func getValidators(h int64) ([]*tdmt_types.Validator, error) {
	p := 1
	vSet := make([]*tdmt_types.Validator, 0)
	for {
		res, err := ctx.CMRpcCli.Validators(&h)
		if err != nil {
			if strings.Contains(err.Error(), "page should be within") {
				return vSet, nil
			}
			return nil, err
		}
		// In case tendermint don't give relayer the right error
		if len(res.Validators) == 0 {
			return vSet, nil
		}
		vSet = append(vSet, res.Validators...)
		p++
	}
}

func getCosmosHdr(h int64) (*cosmos.CosmosHeader, error) {
	rc, err := ctx.CMRpcCli.Commit(&h)
	if err != nil {
		return nil, fmt.Errorf("failed to get Commit of height %d: %v", h, err)
	}
	vSet, err := getValidators(h)
	if err != nil {
		return nil, fmt.Errorf("failed to get Validators of height %d: %v", h, err)
	}
	return &cosmos.CosmosHeader{
		Header:  *rc.Header,
		Commit:  rc.Commit,
		Valsets: vSet,
	}, nil
}
