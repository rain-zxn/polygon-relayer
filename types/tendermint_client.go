package types

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cosmos/cosmos-sdk/codec"

	abcitypes "github.com/tendermint/tendermint/abci/types"
	rpcclient "github.com/tendermint/tendermint/rpc/client"
	tdmt_types "github.com/tendermint/tendermint/types"

	hmTypes "github.com/polynetwork/polygon-relayer/heimdall/types"
	"github.com/polynetwork/polygon-relayer/log"
)

type TendermintClient struct {
	RPCHttp *rpchttp.HTTP
	Codec *codec.Codec
}

var SpanPrefixKey = []byte{0x36} // prefix key to store span

// GetSpanKey appends prefix to start block
func GetSpanKey(id uint64) []byte {
	return append(SpanPrefixKey, []byte(strconv.FormatUint(id, 10))...)
}

func NewTendermintClient(addr string) (*TendermintClient, error) {
	c, err := rpchttp.NewHTTP(addr, "/websocket")
	if err != nil {
		log.Errorf("heimdall_client.NewHeimdallClient - init failed\n")
		return nil, err
	}
	c.Start()

	return &TendermintClient{
		RPCHttp: c,
		Codec: codec.New(),
	}, nil
}

func (this *TendermintClient) GetSpan(id uint64, block int64) (*hmTypes.Span, error) {
	res, err := this.GetSpanRes(id, block)
	if err != nil {
		log.Errorf("tendermint_client.GetSpan - failed, id %s, block %s, %s\n", id, block, err.Error())
		return nil, err
	}

	var span = new(hmTypes.Span)
	err2 := this.Codec.UnmarshalBinaryBare(res.Value, span)
	if err2 != nil {
		log.Errorf("tendermint_client.GetSpan - unmarshal failed, id %s, block %s, %s\n", id, block, err.Error())
		return nil, err2
	}

	return span, nil
}

// block: 0 = latest
func (this *TendermintClient) GetSpanRes(id uint64, block int64) (*abcitypes.ResponseQuery, error) {
	res, err := this.RPCHttp.ABCIQueryWithOptions(
		"/store/bor/key",
		GetSpanKey(11),
		rpcclient.ABCIQueryOptions{Prove: true, Height: block})
	if err != nil {
		log.Errorf("tendermint_client.GetSpanRes - failed, id %s, block %s, %s\n", id, block, err.Error())
		return nil, err
	}

	return &res.Response, nil
}

func (this *TendermintClient) GetCosmosHdr(h int64) (*CosmosHeader, error) {
	rc, err := this.RPCHttp.Commit(&h)
	if err != nil {
		return nil, fmt.Errorf("tendermint_client.GetCosmosHdr - commit error, to get Commit of height %d: %v", h, err)
	}
	vSet, err := this.getValidators(h)
	if err != nil {
		return nil, fmt.Errorf("failed to get Validators of height %d: %v", h, err)
	}
	return &CosmosHeader{
		Header:  *rc.Header,
		Commit:  rc.Commit,
		Valsets: vSet,
	}, nil
}

func (this *TendermintClient) getValidators(h int64) ([]*tdmt_types.Validator, error) {
	p := 1
	vSet := make([]*tdmt_types.Validator, 0)
	for {
		res, err := this.RPCHttp.Validators(&h, p, 100)
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

