package tronscan

import (
	"bakso_ayam/pkg/log"
	"bakso_ayam/pkg/scanner"
	paypb "bakso_ayam/proto/gen/go/protocol/stats/v1"
	"context"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"sort"
	"strconv"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

const HostProd = "https://apilist.tronscanapi.com"

type Client struct {
	HttpCli *http.Client
	Host    string
	Logger  log.Logger
}

var CoinMap = map[string]*paypb.Coin{
	"TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t": {
		Address:  "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t",
		Chain:    paypb.Chain_ChainTron,
		CoinName: paypb.Token_TokenUSDT,
		Decimals: 6,
	},
}

func tokenAmWei(addr string, am string) (*big.Int, paypb.Token, error) {

	t, ok := CoinMap[addr]
	if !ok {
		return nil, paypb.Token_TokenUnknown, errors.New("token is not supported: " + addr)
	}

	b, ok := big.NewInt(0).SetString(am, 10)
	if !ok {
		return nil, paypb.Token_TokenUnknown, errors.New("invalid amount")
	}

	return b, t.CoinName, nil

}

func NewClient() *Client {
	return &Client{
		HttpCli: &http.Client{},
		Host:    HostProd,
	}
}

func (c *Client) Spec() []*paypb.Coin {

	out := make([]*paypb.Coin, 0)

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

func (c *Client) Txs(ctx context.Context, util time.Time, addr string) ([]*paypb.Tx, error) {

	start := 0
	limit := 50

	out := make([]*paypb.Tx, 0)

	untilUnix := util.Unix()

	pack := 1

	pn := "pack: " + strconv.Itoa(pack) + " "
	for {

		if c.Logger != nil {
			c.Logger.Debug(pn + "ask for " + strconv.Itoa(limit) + " txs from " + strconv.Itoa(start))
		}

		res, err := c.tx(ctx, &TxReq{
			Limit: limit,
			Start: start,
			Addr:  addr,
			Sort:  "-timestamp",
		})

		c.Debug(pn + "received " + strconv.Itoa(len(res.Data)) + " txs from " + strconv.Itoa(start))

		start += len(res.Data)
		if err != nil {
			return nil, err
		}

		out = append(out, c.castTx(res)...)

		c.Debug(pn + "total txs: " + strconv.Itoa(len(out)))

		sort.Slice(out, func(i, j int) bool {
			return out[i].Created.AsTime().After(out[j].Created.AsTime())
		})

		max := out[0].Created.AsTime()
		min := out[len(out)-1].Created.AsTime()

		c.Debug(fmt.Sprintf(pn+"[%s %s]", min.String(), max.String()))

		if len(out) > 0 {
			if min.Unix() < untilUnix {
				break
			}
		}

		pack++
	}

	out = filter(out, util)

	return out, nil
}

func filter(in []*paypb.Tx, until time.Time) []*paypb.Tx {

	sort.Slice(in, func(i, j int) bool {
		return in[i].Created.AsTime().After(in[j].Created.AsTime())
	})

	out := make([]*paypb.Tx, 0)
	for i := range in {
		if in[i].Created.AsTime().UnixMilli() < until.UnixMilli() {
			break
		}
		out = append(out, in[i])
	}

	return out
}

type TxReq struct {
	Limit int
	Start int
	Addr  string
	Sort  string
}

func (c *Client) castTx(r *TxRes) []*paypb.Tx {
	out := make([]*paypb.Tx, 0)

	for _, el := range r.Data {

		if el.TriggerInfo == nil {
			continue
		}

		tokenAddr, err := extract[string](el.TriggerInfo, "contract_address")
		if err != nil {
			c.Err(err, "extract")
			continue
		}

		methodName, err := extract[string](el.TriggerInfo, "methodName")
		if err != nil {
			c.Err(err, "extract")
			continue
		}

		if *methodName != "transfer" {
			continue
		}

		parameter, err := extract[map[string]any](el.TriggerInfo, "parameter")
		if err != nil {
			c.Err(err, "extract")
			continue
		}

		value, err := extract[string](*parameter, "_value")
		if err != nil {
			c.Err(err, "extract")
			continue
		}

		am, token, err := tokenAmWei(*tokenAddr, *value)
		if err != nil {
			c.Err(err, "extract")
			continue
		}

		status := paypb.TxStatus_TxStatusW8
		if el.Result == "SUCCESS" {
			status = paypb.TxStatus_TxStatusOK
		}

		out = append(out, &paypb.Tx{
			From:    el.ContractData.OwnerAddress,
			Status:  status,
			Token:   token,
			Created: timestamppb.New(time.UnixMilli(el.Timestamp)),
			To:      el.ToAddress,
			Amount:  am.String(),
		})
	}

	return out
}

type TxRes struct {
	Total      int `json:"total"`
	RangeTotal int `json:"rangeTotal"`
	Data       []struct {
		Block           int      `json:"block"`
		Hash            string   `json:"hash"`
		Timestamp       int64    `json:"timestamp"`
		OwnerAddress    string   `json:"ownerAddress"`
		OwnerAddressTag string   `json:"ownerAddressTag"`
		ToAddressList   []string `json:"toAddressList"`
		ToAddress       string   `json:"toAddress"`
		ContractType    int      `json:"contractType"`
		Confirmed       bool     `json:"confirmed"`
		Revert          bool     `json:"revert"`
		ContractData    struct {
			Data            string `json:"data,omitempty"`
			OwnerAddress    string `json:"owner_address"`
			ContractAddress string `json:"contract_address,omitempty"`
			Amount          int64  `json:"amount,omitempty"`
			ToAddress       string `json:"to_address,omitempty"`
		} `json:"contractData"`
		SmartCalls  string `json:"SmartCalls"`
		Events      string `json:"Events"`
		Id          string `json:"id"`
		Data        string `json:"data"`
		Fee         string `json:"fee"`
		ContractRet string `json:"contractRet"`
		Result      string `json:"result"`
		Amount      string `json:"amount"`
		CheatStatus bool   `json:"cheatStatus"`
		Cost        struct {
			NetFee             int `json:"net_fee"`
			EnergyPenaltyTotal int `json:"energy_penalty_total"`
			EnergyUsage        int `json:"energy_usage"`
			Fee                int `json:"fee"`
			EnergyFee          int `json:"energy_fee"`
			EnergyUsageTotal   int `json:"energy_usage_total"`
			OriginEnergyUsage  int `json:"origin_energy_usage"`
			NetUsage           int `json:"net_usage"`
		} `json:"cost"`
		TokenInfo struct {
			TokenId      string `json:"tokenId"`
			TokenAbbr    string `json:"tokenAbbr"`
			TokenName    string `json:"tokenName"`
			TokenDecimal int    `json:"tokenDecimal"`
			TokenCanShow int    `json:"tokenCanShow"`
			TokenType    string `json:"tokenType"`
			TokenLogo    string `json:"tokenLogo"`
			TokenLevel   string `json:"tokenLevel"`
			Vip          bool   `json:"vip"`
		} `json:"tokenInfo"`
		TokenType       string         `json:"tokenType"`
		TriggerInfo     map[string]any `json:"trigger_info,omitempty"`
		RiskTransaction bool           `json:"riskTransaction"`
	} `json:"data"`
	WholeChainTxCount int64           `json:"wholeChainTxCount"`
	ContractMap       map[string]bool `json:"contractMap"`
}

func (c *Client) tx(ctx context.Context, req *TxReq) (*TxRes, error) {

	url := scanner.MakeUrl(c.Host,
		"/api/transaction?sort=", req.Sort,
		"&count=true",
		"&limit=", strconv.Itoa(req.Limit),
		"&start=", strconv.Itoa(req.Start),
		"&address=", req.Addr,
	)

	return scanner.Request[TxReq, TxRes](ctx, c.HttpCli, http.MethodGet, url, req)
}

//https://apilist.tronscanapi.com/transaction?sort=-timestamp&count=true&limit=100&start=0&address=TNM7ySgqGHHu7gbymh3yinLLp4aR9c7N2W
//https://apilist.tronscanapi.com/api/transaction?sort=-timestamp&count=true&limit=20&start=0&address=TNM7ySgqGHHu7gbymh3yinLLp4aR9c7N2W

func extract[T any](m map[string]any, key string) (*T, error) {
	if m == nil {
		return nil, errors.New("m is empty")
	}

	v, ok := m[key]
	if !ok {
		return nil, errors.New("no such key")
	}

	vc, ok := v.(T)
	if !ok {
		return nil, errors.New("key is not type of T")
	}

	return &vc, nil
}
