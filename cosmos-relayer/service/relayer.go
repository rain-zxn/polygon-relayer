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
	"fmt"
	"strings"
	"time"

	"github.com/polynetwork/polygon-relayer/cosmos-relayer/context"
	"github.com/polynetwork/polygon-relayer/log"

	mcli "github.com/polynetwork/poly-go-sdk/client"

	mcosmos "github.com/polynetwork/polygon-relayer/types"
)

func StartRelay() {
	go ToPolyRoutine()
	// go ToCosmosRoutine()
}

// Process with message from channel `ToPoly`.
// When type is `TyHeader`, we must finish the procession for this message
// before processing the transaction-messages.
// When type is `TyTx`, we relay this transaction info and its proof to Poly.
// When type is `TyUpdateHeight`, we update the cosmos height that already
// checked in our db.
// This run as a go-routine
func ToPolyRoutine() {
	log.LogTender.Infof("relayer.ToPolyRoutine - start")

	for val := range ctx.ToPoly {
		switch val.Type {
		case context.TyHeader:
			log.LogTender.Infof("relayer.ToPolyRoutine - handleCosmosHdrs, lenth: %d", len(val.Hdrs))
			if err := handleCosmosHdrs(val.Hdrs); err != nil {
				log.LogTender.Errorf("relayer.ToPolyRoutine - handleCosmosHdrs, lenth: %d, err: %w", len(val.Hdrs), err)
				// panic(err)
				continue
			}
		case context.TyTx:
			// go handleCosmosTx(val.Tx, val.Hdrs[0])
			continue
		case context.TyUpdateHeight:
			log.LogTender.Infof("relayer.ToPolyRoutine - TyUpdateHeight, height: %d", val.Height)
			go func() {
				if err := ctx.Db.SetCosmosHeight(val.Height); err != nil {
					log.LogTender.Errorf("failed to update cosmos height: %v", err)
				}
			}()
		}
	}
}

// Process cosmos-headers msg. This function would not return before our
// Ploygon tx committing headers confirmed. This guarantee that the next
// cross-chain txs next to relay can be proved on Poly.
func handleCosmosHdrs(headers []*mcosmos.CosmosHeader) error {
	for i := 0; i < len(headers); i += ctx.Conf.HeadersPerBatch {
		var hdrs []*mcosmos.CosmosHeader
		if i+ctx.Conf.HeadersPerBatch > len(headers) {
			hdrs = headers[i:]
		} else {
			hdrs = headers[i : i+ctx.Conf.HeadersPerBatch]
		}
		info := make([]string, len(hdrs))
		raw := make([][]byte, len(hdrs))
		for i, h := range hdrs {
			r, err := ctx.CMCdc.MarshalBinaryBare(*h)
			if err != nil {
				log.LogTender.Fatalf("[handleCosmosHdr] failed to marshal CosmosHeader: %v", err)
				return err
			}
			raw[i] = r
			info[i] = fmt.Sprintf("(hash: %s, height: %d)", h.Header.Hash().String(), h.Header.Height)
			log.LogTender.Info("[handleCosmosHdr] 2" + info[i])
		}

	SYNC_RETRY:
		txhash, err := ctx.Poly.Native.Hs.SyncBlockHeader(ctx.Conf.SideChainId, ctx.PolyAcc.Address,
			raw, ctx.PolyAcc)
		if err != nil {
			if _, ok := err.(mcli.PostErr); ok {
				log.LogTender.Errorf("[handleCosmosHdr] post error, retry after 10 sec wait: %v", err)
				context.SleepSecs(10)
				goto SYNC_RETRY
			}
			if strings.Contains(err.Error(), context.NoUsefulHeaders) {
				log.LogTender.Errorf("[handleCosmosHdr] your headers could be wrong or already committed: headers: [ %s ], error: %w",
					strings.Join(info, ", "), err)
				return fmt.Errorf("[handleCosmosHdr] your headers could be wrong or already committed: headers: [ %s ], error: %w",
				strings.Join(info, ", "), err)
			}
			log.LogTender.Errorf("[handleCosmosHdr] failed to relay cosmos header to Poly: %v", err)
			return err
		}
		tick := time.NewTicker(100 * time.Millisecond)
		var h uint32
		startTime := time.Now()
		for range tick.C {
			h, _ = ctx.Poly.GetBlockHeightByTxHash(txhash.ToHexString())
			curr, _ := ctx.Poly.GetCurrentBlockHeight()
			if h > 0 && curr > h {
				break
			}

			if startTime.Add(100 * time.Millisecond); startTime.Second() > ctx.Conf.ConfirmTimeout {
				str := ""
				for i, hdr := range hdrs {
					str += fmt.Sprintf("( no.%d: hdr-height: %d, hdr-hash: %s )\n", i+1,
						hdr.Header.Height, hdr.Header.Hash().String())
				}
				panic(fmt.Errorf("tx( %s ) is not confirm for a long time ( over %d sec ): {\n%s}",
					txhash.ToHexString(), ctx.Conf.ConfirmTimeout, str))
			}
		}
		log.LogTender.Infof("[handleCosmosHdr] successful to relay header and confirmed on Poly: { headers: [ %s ], poly: "+
			"(poly_tx: %s, poly_tx_height: %d) }", strings.Join(info, ", "), txhash.ToHexString(), h)
	}
	return nil
}
