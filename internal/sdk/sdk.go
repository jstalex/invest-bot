package sdk

import (
	"context"
	"crypto/tls"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/oauth"
	pb "invest-bot/api/proto"
	"invest-bot/internal/config"
	"log"
)

type SDK struct {
	Ctx         context.Context
	Conn        *grpc.ClientConn
	RobotConfig *config.RobotConfig

	Sandbox     pb.SandboxServiceClient
	Instruments pb.InstrumentsServiceClient
	Marketdata  pb.MarketDataServiceClient
	Operations  pb.OperationsServiceClient
	Orders      pb.OrdersServiceClient
	Stoporder   pb.StopOrdersServiceClient
	Users       pb.UsersServiceClient
}

func NewSDK(cnf *config.RobotConfig) *SDK {
	ctx := context.WithValue(context.Background(), "authorization", "Bearer "+cnf.Token)

	conn, err := grpc.Dial(cnf.EndPoint,
		grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})),
		grpc.WithPerRPCCredentials(oauth.NewOauthAccess(&oauth2.Token{AccessToken: cnf.Token})))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	//defer conn.Close()
	return &SDK{
		Ctx:         ctx,
		Conn:        conn,
		RobotConfig: cnf,
		Sandbox:     pb.NewSandboxServiceClient(conn),
		Instruments: pb.NewInstrumentsServiceClient(conn),
		Marketdata:  pb.NewMarketDataServiceClient(conn),
		Operations:  pb.NewOperationsServiceClient(conn),
		Orders:      pb.NewOrdersServiceClient(conn),
		Stoporder:   pb.NewStopOrdersServiceClient(conn),
		Users:       pb.NewUsersServiceClient(conn),
	}
}
