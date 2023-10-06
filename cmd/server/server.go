package main

import (
	"context"
	"net"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	paypb "github.com/hardstylez72/bakso_ayam/proto/gen/go/protocol/stats/v1"
	"google.golang.org/grpc"
)

var port = ":3311"

func main() {
	server := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer()),
	)

	paypb.RegisterPayServiceServer(server, &Server{})

	lis, err := net.Listen("tcp", port)
	if err != nil {
		panic(err)
	}

	if err := server.Serve(lis); err != nil {
		return
	}
}

type Server struct {
	paypb.UnimplementedPayServiceServer
}

func (s *Server) Coins(ctx context.Context, req *paypb.CoinsReq) (*paypb.CoinsRes, error) {

}

func (s *Server) Txs(ctx context.Context, req *paypb.TxsReq) (*paypb.TxsRes, error) {

}
