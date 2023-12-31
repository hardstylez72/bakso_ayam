package scanner

import (
	"context"
	"time"

	pbv1 "github.com/hardstylez72/bakso_ayam/proto/gen/go/protocol/stats/v1"
)

type Scanner interface {
	Txs(ctx context.Context, until time.Time, addr string, direction pbv1.TxDirection) ([]*pbv1.Tx, error)
	Spec() []*pbv1.Coin
}
