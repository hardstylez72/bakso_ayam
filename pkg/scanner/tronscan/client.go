package tronscan

import (
	"context"
	"fmt"
	"math/big"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/hardstylez72/bakso_ayam/pkg/log"
	"github.com/hardstylez72/bakso_ayam/pkg/scanner"
	pbv1 "github.com/hardstylez72/bakso_ayam/proto/gen/go/protocol/stats/v1"
	"github.com/pkg/errors"
	"go.uber.org/ratelimit"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const HostProd = "https://api.trongrid.io"

//docs - https://developers.tron.network/reference/select-network

type Client struct {
	HttpCli *http.Client
	Host    string
	Logger  log.Logger
	ApiKey  string
}

var CoinMap = map[string]*pbv1.Coin{
	"TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t": {
		Address:  "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t",
		Chain:    pbv1.Chain_ChainTron,
		CoinName: pbv1.Token_TokenUSDT,
		Decimals: 6,
	},
	"THk5qH79SoAaUnUh8JVdRarSESTZpqPjSQ": {
		Address:  "THk5qH79SoAaUnUh8JVdRarSESTZpqPjSQ",
		Chain:    pbv1.Chain_ChainTron,
		CoinName: pbv1.Token_TokenUSDT,
		Decimals: 18,
	},
	"TTmQYPPZ3N3AfSFx4NKm4or2zwDn8dvKRE": {
		Address:  "THk5qH79SoAaUnUh8JVdRarSESTZpqPjSQ",
		Chain:    pbv1.Chain_ChainTron,
		CoinName: pbv1.Token_TokenUSDT,
		Decimals: 18,
	},
}

func tokenAmWei(addr string, am string) (*big.Int, pbv1.Token, error) {

	t, ok := CoinMap[addr]
	if !ok {
		return nil, pbv1.Token_TokenUnknown, errors.New("token is not supported: " + addr)
	}

	b, ok := big.NewInt(0).SetString(am, 10)
	if !ok {
		return nil, pbv1.Token_TokenUnknown, errors.New("invalid amount")
	}

	return b, t.CoinName, nil

}

var rl = ratelimit.New(1, ratelimit.Per(time.Second))

func NewClient() *Client {

	return &Client{
		HttpCli: &http.Client{},
		Host:    HostProd,
		Logger:  nil,
	}
}

func (c *Client) Spec() []*pbv1.Coin {

	out := make([]*pbv1.Coin, 0)

	for _, coin := range CoinMap {
		out = append(out, coin)
	}
	return out
}

func (c *Client) Debug(s string) {
	if c.Logger != nil {
		c.Logger.Debug(s)
	}
}

func (c *Client) Err(err error, s string) {
	if c.Logger != nil {
		c.Logger.Err(err, s)
	}
}

func (c *Client) Txs(ctx context.Context, util time.Time, addr string, direction pbv1.TxDirection) ([]*pbv1.Tx, error) {

	limit := 200 // максимум

	out := make([]*pbv1.Tx, 0)

	pack := 1

	fromBase := time.Now()

	c.Debug("request for address " + addr + " direction: " + direction.String())
	c.Debug(" from " + fromBase.Format(time.DateTime) + " to " + util.Format(time.DateTime))

	from := fromBase

	next := c.base(&TxReq{
		Limit:     limit,
		From:      from,
		To:        util,
		Addr:      addr,
		Sort:      "-timestamp",
		Direction: direction,
	})

	for {

		c.Debug("всего: " + strconv.Itoa(len(out)))

		if next == "" {
			return out, nil
		}

		pn := "[pack: " + strconv.Itoa(pack) + "] "

		c.Logger.Debug(pn + "ask for " + strconv.Itoa(limit) + " txs from ")

		res, err := c.nextTx(ctx, next)
		if err != nil {
			return nil, errors.Wrap(err, "tronscan.tx")
		}

		next = res.Meta.Links.Next

		c.Debug(pn + "received " + strconv.Itoa(len(res.Data)))

		if len(res.Data) == 0 {
			break
		}

		tmp := c.castTx(res)
		out = append(out, tmp...)

		if len(tmp) == 0 {
			continue
		}

		sort.Slice(tmp, func(i, j int) bool {
			return tmp[i].Created.AsTime().After(tmp[j].Created.AsTime())
		})

		max := tmp[0].Created.AsTime()
		min := tmp[len(tmp)-1].Created.AsTime()

		c.Debug(fmt.Sprintf(pn+"[%s %s]", min.Format(time.DateTime), max.Format(time.DateTime)))

		pack++
	}

	out = filter(out, util)

	return out, nil
}

func filter(in []*pbv1.Tx, until time.Time) []*pbv1.Tx {

	uniq := make(map[string]*pbv1.Tx)

	for _, el := range in {
		uniq[el.From+el.To+el.Amount+el.Created.String()] = el
	}

	tmp := make([]*pbv1.Tx, 0)
	for _, tx := range uniq {
		tmp = append(tmp, tx)
	}

	sort.Slice(tmp, func(i, j int) bool {
		return tmp[i].Created.AsTime().After(tmp[j].Created.AsTime())
	})

	out := make([]*pbv1.Tx, 0)
	for i := range tmp {
		if tmp[i].Created.AsTime().UnixMilli() < until.UnixMilli() {
			break
		}
		out = append(out, tmp[i])
	}

	return out
}

type TxReq struct {
	Limit     int
	From      time.Time
	To        time.Time
	Addr      string
	Sort      string
	Direction pbv1.TxDirection
}

func (c *Client) castTx(r *TxRes) []*pbv1.Tx {
	out := make([]*pbv1.Tx, 0)

	for _, el := range r.Data {

		am, token, err := tokenAmWei(el.TokenInfo.Address, el.Value)
		if err != nil {
			c.Err(err, "extract")
			continue
		}

		status := pbv1.TxStatus_TxStatusOK

		out = append(out, &pbv1.Tx{
			From:    el.From,
			Status:  status,
			Token:   token,
			Created: timestamppb.New(time.UnixMilli(el.BlockTimestamp)),
			To:      el.To,
			Amount:  am.String(),
		})
	}

	return out
}

type TxRes struct {
	Data []struct {
		TransactionId string `json:"transaction_id"`
		TokenInfo     struct {
			Symbol   string `json:"symbol"`
			Address  string `json:"address"`
			Decimals int    `json:"decimals"`
			Name     string `json:"name"`
		} `json:"token_info"`
		BlockTimestamp int64  `json:"block_timestamp"`
		From           string `json:"from"`
		To             string `json:"to"`
		Type           string `json:"type"`
		Value          string `json:"value"`
	} `json:"data"`
	Success bool `json:"success"`
	Meta    struct {
		At          int64  `json:"at"`
		Fingerprint string `json:"fingerprint"`
		Links       struct {
			Next string `json:"next"`
		} `json:"links"`
		PageSize int `json:"page_size"`
	} `json:"meta"`
}

func (c *Client) nextTx(ctx context.Context, u string) (*TxRes, error) {
	rl.Take()

	m := map[string]string{"TRON-PRO-API-KEY": c.ApiKey}

	return scanner.Request[TxReq, TxRes](ctx, c.HttpCli, http.MethodGet, u, nil, m)
}
func (c *Client) base(req *TxReq) string {

	url := scanner.MakeUrl(c.Host,
		"/v1/accounts/", req.Addr,
		"/transactions/trc20?limit=", strconv.Itoa(req.Limit),
		"&only_confirmed=true",
		"&min_timestamp=", req.From.Format("2006-01-02T15:04:05"),
		//"&max_timestamp=", req.To.Format("2006-01-02T15:04:05"),
		"&order_by=block_timestamp,desc",
		//"&address=", req.Addr,
	)

	switch req.Direction {
	case pbv1.TxDirection_DirectionIn:
		url += "&only_to=true"
	case pbv1.TxDirection_DirectionOut:
		url += "&only_from=true"
	}

	return url
}

//https://api.trongrid.io/v1/accounts/THeRkUEytKBT3kRYmGqcu5YsWxuFwKAwNs/transactions/trc20?limit=3&only_confirmed=true&min_timestamp=2023-10-08T08:53:48&max_timestamp=2022-10-13T00:53:48&order_by=block_timestamp,asc&only_from=true
//https://api.trongrid.io/v1/accounts/THeRkUEytKBT3kRYmGqcu5YsWxuFwKAwNs/transactions/trc20?limit=3&only_confirmed=true&min_timestamp=2023-10-08T08:53:48&max_timestamp=2022-10-13T00:53:48&order_by=block_timestamp,asc&only_from=true
//https://apilist.tronscanapi.com/transaction?sort=-timestamp&count=true&limit=100&start=0&address=TNM7ySgqGHHu7gbymh3yinLLp4aR9c7N2W
//https://apilist.tronscanapi.com/api/transaction?sort=-timestamp&count=true&limit=20&start=0&address=TNM7ySgqGHHu7gbymh3yinLLp4aR9c7N2W
