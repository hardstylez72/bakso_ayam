package e2e

import (
	"context"
	"testing"
	"time"

	pbv1 "github.com/hardstylez72/bakso_ayam/proto/gen/go/protocol/stats/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestCli(t *testing.T) {
	conn, err := grpc.Dial("0.0.0.0:9011", grpc.WithTransportCredentials(insecure.NewCredentials()))
	assert.NoError(t, err)

	cli := pbv1.NewPayServiceClient(conn)

	res, err := cli.Coins(context.Background(), &emptypb.Empty{})
	assert.NoError(t, err)
	assert.NotNil(t, res)

	day := time.Hour * 24
	month := day * 30

	txs, err := cli.Txs(context.Background(), &pbv1.TxsReq{
		Direction: pbv1.TxDirection_DirectionOut,
		Address:   "TNM7ySgqGHHu7gbymh3yinLLp4aR9c7N2W",
		Chain:     pbv1.Chain_ChainTron,
		Until:     timestamppb.New(time.Now().Add(-month * 12)),
	})
	assert.NoError(t, err)
	assert.NotNil(t, txs)

}
