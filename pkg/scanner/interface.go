package scanner

import (
	paypb "bakso_ayam/proto/gen/go/protocol/stats/v1"
	"context"
	"time"
)

type Scanner interface {
	Txs(ctx context.Context, until time.Time, addr string) ([]*paypb.Tx, error)
	Spec() []*paypb.Coin
}
