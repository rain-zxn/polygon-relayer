package manager

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/polynetwork/polygon-relayer/cosmos-sdk/codec"
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
}

var SpanPrefixKey = []byte{0x36} // prefix key to store span

// GetSpanKey appends prefix to start block
func GetSpanKey(id uint64) []byte {
	return append(SpanPrefixKey, []byte(strconv.FormatUint(id, 10))...)
}

func NewTendermintClient(addr string) (*TendermintClient, error) {
	c := rpcclient.NewHTTP(addr, "/websocket")

	c.Start()

	return &TendermintClient{
		RPCHttp: c,
		Codec:   codec.New(),
	}, nil
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
