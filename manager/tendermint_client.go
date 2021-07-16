package manager

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/polynetwork/polygon-relayer/cosmos-sdk/codec"
	"github.com/polynetwork/polygon-relayer/db"
	"github.com/polynetwork/polygon-relayer/types"

	abcitypes "github.com/christianxiao/tendermint/abci/types"
	rpcclient "github.com/christianxiao/tendermint/rpc/client"
	tdmt_types "github.com/christianxiao/tendermint/types"

	hmTypes "github.com/polynetwork/polygon-relayer/heimdall/types"
	"github.com/polynetwork/polygon-relayer/log"
)

type TendermintClient struct {
	RPCHttp *rpcclient.HTTP
	Codec   *codec.Codec

	db       *db.BoltDB
	exitChan chan int
}

var SpanPrefixKey = []byte{0x36} // prefix key to store span

// GetSpanKey appends prefix to start block
func GetSpanKey(id uint64) []byte {
	return append(SpanPrefixKey, []byte(strconv.FormatUint(id, 10))...)
}

func NewTendermintClient(addr string, db *db.BoltDB) (*TendermintClient, error) {
	c := rpcclient.NewHTTP(addr, "/websocket")

	c.Start()

	return &TendermintClient{
		RPCHttp:  c,
		Codec:    codec.New(),
		db:       db,
		exitChan: make(chan int),
	}, nil
}

type StartEnd struct {
	Start uint64
	End uint64
}

func (this *TendermintClient) GetSpanIdByBor(bor uint64) (uint64, error) {
	all, err := this.db.GetAllUint64(db.BKTSpan)
	if err != nil {
		return 0, err
	}
	for _, v := range all {
		var startEnd = new(StartEnd)
		json.Unmarshal(v.V, startEnd)
		if startEnd.Start <= bor && bor <= startEnd.End {
			return v.K, nil
		}
	}
	return 0, fmt.Errorf("GetSpanIdByBor: not found %s", bor)
}

func (this *TendermintClient) MonitorSpanLatestRoutine(seconds uint64) {
	log.LogSpanL.Infof("tendermint_client.MonitorSpanLatestRoutine - start, Duration %s", seconds)

	fetchBlockTicker := time.NewTicker(time.Duration(seconds) * time.Second)
	for {
		select {
		case <-fetchBlockTicker.C:
			h, err := this.GetLatestHeight()
			if err != nil {
				log.LogSpanL.Errorf("MonitorSpan - GetLatestHeight error, err: %s", err.Error())
				continue
			}
			span, err := this.GetLatestSpan(h)
			if err != nil {
				log.LogSpanL.Errorf("MonitorSpan - cannot get span from node height: %s err: %s", h, err.Error())
				continue
			}

			//check db
			val, err := this.db.GetUint64(db.BKTSpan, span.ID)
			if err != nil {
				log.LogSpanL.Errorf("MonitorSpan - db.GetUint64 error, spanId %s  error: %s", span.ID, err.Error())
			}
			if len(val) != 0 {
				continue
			}

			var se = &StartEnd{
				Start: span.StartBlock,
				End: span.EndBlock,
			}

			vjson, err := json.Marshal(se)
			if err != nil {
				log.LogSpanL.Errorf("MonitorSpan - Marshal, err: %s", err.Error())
				continue
			}
			err2 := this.db.PutUint64(db.BKTSpan, span.ID, vjson)
			if err2 != nil {
				log.LogSpanL.Errorf("MonitorSpan - db.PutUint64 err: %s", err2.Error())
				continue
			}

		case <-this.exitChan:
			return
		}
	}
}

func (this *TendermintClient) MonitorSpanHisRoutine(start uint64) {
	log.LogSpanH.Infof("tendermint_client.MonitorSpanHisRoutine - start, start %s", start)

	for true {
        all, err := this.db.GetAllUint64(db.BKTSpan)
		if err != nil {
			log.LogSpanH.Errorf("MonitorSpanHisRoutine - error, err: %s", err.Error())
			time.Sleep(60 * time.Second)
			continue
		}

		allmap := make(map[uint64][]byte)
		for _, v := range all {
			allmap[v.K] = v.V
		}

		max := all[0].K
		for i:=max-1; i>=start; i-- {
			_, ok := allmap[i]
			if !ok {
				_, span, err := this.GetSpanRes(i, 0)
				if err != nil {
					log.LogSpanH.Errorf("MonitorSpanHisRoutine - GetSpanRes error, id %d, err: %s", i, err.Error())
					time.Sleep(60 * time.Second)
					continue
				}

				var se = &StartEnd{
					Start: span.StartBlock,
					End: span.EndBlock,
				}
	
				vjson, err := json.Marshal(se)
				if err != nil {
					log.LogSpanH.Errorf("MonitorSpanHisRoutine - Marshal, err: %s", err.Error())
					continue
				}
				err2 := this.db.PutUint64(db.BKTSpan, span.ID, vjson)
				if err2 != nil {
					log.LogSpanH.Errorf("MonitorSpanHisRoutine - db.PutUint64 err: %s", err2.Error())
					continue
				}
			}
		}

        time.Sleep(60 * time.Second)
    }
}

func (this *TendermintClient) GetLatestHeight() (int64, error) {
	statusRes, err := this.RPCHttp.Status()
	if err != nil {
		return 0, err
	}
	return statusRes.SyncInfo.LatestBlockHeight, nil
}

func (this *TendermintClient) GetLatestSpan(block int64) (*hmTypes.Span, error) {
	res, err := this.RPCHttp.ABCIQueryWithOptions(
		"custom/bor/latest-span",
		nil,
		rpcclient.ABCIQueryOptions{Prove: true, Height: block})
	if err != nil {
		log.Errorf("tendermint_client.GetSpanRes - failed, block %s, %s\n", block, err.Error())
		return nil, err
	}

	var span = new(hmTypes.Span)
	err2 := this.Codec.UnmarshalBinaryBare(res.Response.Value[:], span)
	if err2 != nil {
		log.Errorf("tendermint_client.GetSpan - unmarshal failed, block %s, %s\n", block, err.Error())
		return nil, err2
	}

	return span, nil
}

// block: 0 = latest
func (this *TendermintClient) GetSpanRes(id uint64, block int64) (*abcitypes.ResponseQuery, *hmTypes.Span, error) {
	res, err := this.RPCHttp.ABCIQueryWithOptions(
		"/store/bor/key",
		GetSpanKey(11),
		rpcclient.ABCIQueryOptions{Prove: true, Height: block})
	if err != nil {
		log.Errorf("tendermint_client.GetSpanRes - failed, id %s, block %s, %s\n", id, block, err.Error())
		return nil, nil, err
	}

	var span = new(hmTypes.Span)
	err2 := this.Codec.UnmarshalBinaryBare(res.Response.Value[:], span)
	if err2 != nil {
		log.Errorf("tendermint_client.GetSpan - unmarshal failed, id %s, block %s, %s\n", id, block, err.Error())
		return nil, nil, err2
	}

	return &res.Response, span, nil
}

func (this *TendermintClient) GetCosmosHdr(h int64) (*types.CosmosHeader, error) {
	rc, err := this.RPCHttp.Commit(&h)
	if err != nil {
		return nil, fmt.Errorf("tendermint_client.GetCosmosHdr - commit error, to get Commit of height %d: %v", h, err)
	}
	vSet, err := this.getValidators(h)
	if err != nil {
		return nil, fmt.Errorf("failed to get Validators of height %d: %v", h, err)
	}
	return &types.CosmosHeader{
		Header:  *rc.Header,
		Commit:  rc.Commit,
		Valsets: vSet,
	}, nil
}

func (this *TendermintClient) getValidators(h int64) ([]*tdmt_types.Validator, error) {
	vSet := make([]*tdmt_types.Validator, 0)

	res, err := this.RPCHttp.Validators(&h)
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

	return vSet, nil
}
