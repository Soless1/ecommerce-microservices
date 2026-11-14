package app

import (
	"context"
	"errors"
	"net"

	"github.com/igoroutine-courses/microservices.ecommerce.notifications/internal/config"
	"github.com/igoroutine-courses/microservices.ecommerce.notifications/internal/controller"
	"github.com/igoroutine-courses/microservices.ecommerce.notifications/internal/controller/notifications"
	notificationsusecase "github.com/igoroutine-courses/microservices.ecommerce.notifications/internal/usecase/notifications"
	notificaitonspb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/notifications/api/notifications/v1"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

func Run(logger *zap.Logger, cfg *config.Config) {

	notificationsService := notificationsusecase.NewNotificationsServer(logger, cfg.CallbackAddr)

	notificationsServer := notifications.NewNotificationsServer(notificationsService)
	api := controller.New(notificationsServer)

	group, ctx := errgroup.WithContext(context.Background())
	group.Go(func() error {
		return runGrpc(ctx, logger, cfg, api)
	})
	if err := group.Wait(); err != nil {
		logger.Error("Notifications server stopped with an error", zap.Error(err))
	}
}

func runGrpc(ctx context.Context, logger *zap.Logger, cfg *config.Config, ctrl *controller.API) error {
	lis, err := net.Listen("tcp", ":"+cfg.GRPC.Port)
	if err != nil {
		return err
	}
	grpcServer := grpc.NewServer()

	notificaitonspb.RegisterNotificationsServer(grpcServer, ctrl.Notifications)

	logger.Info("gRPC server is listening", zap.String("port", cfg.GRPC.Port))

	serveErr := make(chan error, 1)
	go func() {
		serveErr <- grpcServer.Serve(lis)
	}()

	select {
	case <-ctx.Done():
		grpcServer.GracefulStop()
		err := <-serveErr
		if err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			return err
		}
		return nil
	case err := <-serveErr:
		if errors.Is(err, grpc.ErrServerStopped) {
			return nil
		}
		return err
	}
}
