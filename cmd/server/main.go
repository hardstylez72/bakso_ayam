package main

import (
	"context"
	"errors"
	"net"
	"os"

	"github.com/hardstylez72/bakso_ayam/pkg/log"
	"github.com/hardstylez72/bakso_ayam/pkg/scanner"
	"github.com/hardstylez72/bakso_ayam/pkg/scanner/tronscan"
	"github.com/hardstylez72/bakso_ayam/proto/gen/go/protocol/stats/v1"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

var (
	port          = os.Getenv("GRPC_PORT")         // 90
	tronGridToken = os.Getenv("TRON_GRID_API_KEY") //"56d70809-dc34-44d5-a99c-e6a8d0287a34"
)

func main() {
	server := grpc.NewServer()

	s := &Server{}
	pbv1.RegisterPayServiceServer(server, s)

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		panic(err)
	}

	println("server listen: ", lis.Addr().String())

	if err := server.Serve(lis); err != nil {
		return
	}
}

type Server struct {
	pbv1.UnimplementedPayServiceServer
}

func (s *Server) Coins(ctx context.Context, req *emptypb.Empty) (*pbv1.CoinsRes, error) {
	tronsScanner := tronscan.NewClient()
	return &pbv1.CoinsRes{
		Coins: tronsScanner.Spec(),
	}, nil
}

func (s *Server) Txs(ctx context.Context, req *pbv1.TxsReq) (*pbv1.TxsRes, error) {

	var scan scanner.Scanner
	switch req.Chain {
	case pbv1.Chain_ChainTron:
		c := tronscan.NewClient()
		c.ApiKey = tronGridToken
		c.Logger = &log.Simple{}
		scan = c
	default:
		return nil, errors.New("unsupported chain: " + req.Chain.String())
	}
	txs, err := scan.Txs(ctx, req.Until.AsTime(), req.Address, req.Direction)
	if err != nil {
		return nil, err
	}

	return &pbv1.TxsRes{
		Txs: txs,
	}, nil
}
