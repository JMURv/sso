package discovery

import (
	"context"
	"fmt"
	pb "github.com/JMURv/protos/discovery"
	"github.com/JMURv/sso/internal/discovery"
	conf "github.com/JMURv/sso/pkg/config"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Discovery struct {
	cli  *grpc.ClientConn
	name string
	addr string
}

func New(url *conf.SrvDiscoveryConfig, name string, addr *conf.ServerConfig) discovery.ServiceDiscovery {
	cli, err := grpc.NewClient(
		fmt.Sprintf("%v:%v", url.Host, url.Port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		zap.L().Fatal("failed to create client", zap.Error(err))
	}

	return &Discovery{
		cli:  cli,
		name: name,
		addr: fmt.Sprintf("%v:%v", addr.Domain, addr.Port),
	}
}

func (d *Discovery) Close() error {
	return d.cli.Close()
}

func (d *Discovery) Register(ctx context.Context) error {
	_, err := pb.NewServiceDiscoveryClient(d.cli).Register(
		ctx, &pb.NameAndAddressMsg{
			Name:    d.name,
			Address: d.addr,
		},
	)
	if err != nil {
		zap.L().Debug(discovery.ErrFailedToRegister.Error(), zap.Error(err))
		return discovery.ErrFailedToRegister
	}

	return nil
}

func (d *Discovery) Deregister(ctx context.Context) error {
	_, err := pb.NewServiceDiscoveryClient(d.cli).Deregister(
		ctx, &pb.NameAndAddressMsg{
			Name:    d.name,
			Address: d.addr,
		},
	)
	if err != nil {
		zap.L().Debug(discovery.ErrFailedToDeregister.Error(), zap.Error(err))
		return discovery.ErrFailedToDeregister
	}

	return nil
}

func (d *Discovery) FindServiceByName(ctx context.Context, name string) (addr string, err error) {
	res, err := pb.NewServiceDiscoveryClient(d.cli).FindService(
		ctx, &pb.ServiceNameMsg{
			Name: name,
		},
	)
	if err != nil {
		zap.L().Debug(discovery.ErrFailedToFindService.Error(), zap.Error(err))
		return addr, discovery.ErrFailedToFindService
	}

	return res.Address, nil
}
