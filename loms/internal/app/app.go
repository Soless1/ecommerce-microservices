package app

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	notificationsadapter "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/adapter/notifications/grpc"
	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/config"
	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/controller"
	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/outbox"
	outboxrepo "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/repository/outbox"
	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/repository/transactor"
	lomsmigrations "github.com/igoroutine-courses/microservices.ecommerce.loms/migrations"
	lompb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/loms/api/loms/v1"
	productpb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/loms/api/product/v1"
	stockspb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/loms/api/stocks/v1"
	notificaitonspb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/notifications/api/notifications/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	stockscontroller "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/controller/stocks"
	stocksrepo "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/repository/stocks"
	stocksusecase "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/usecase/stocks"

	productcontroller "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/controller/product"
	productrepo "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/repository/product"
	productusecase "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/usecase/product"

	lomscontroller "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/controller/loms"
	orderrepo "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/repository/order"
	lomsusecase "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/usecase/loms"
)

func Run(logger *zap.Logger, cfg *config.Config) {
	ctx := context.Background() //TODO сделать через errGroup как в cart

	notificaitonConn, err := grpc.NewClient(
		cfg.Clients.NotificationsGrpcAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		logger.Fatal("failed to connect to server", zap.Error(err))
	}

	defer func(conn *grpc.ClientConn) {
		_ = conn.Close()
	}(notificaitonConn)

	notificationsClient := notificationsadapter.NewNotificationsClient(
		notificaitonspb.NewNotificationsClient(notificaitonConn),
	)

	pool, err := newPostgresPool(ctx, cfg.Postgres.DSN)
	if err != nil {
		logger.Error("failed to initialize PostgreSQL", zap.Error(err))
		return
	}
	defer pool.Close()

	if err := lomsmigrations.Run(pool); err != nil {
		logger.Error("failed to run LOMS migrations", zap.Error(err))
		return
	}

	lomsTransactor := transactor.NewTransactor(pool)

	outboxRepo := outboxrepo.NewOutboxRepository(pool)

	stocksRepo := stocksrepo.NewPostgresRepository(pool)
	stocksService := stocksusecase.NewStocksService(stocksRepo)
	stocksServer := stockscontroller.NewStocksServer(stocksService)

	productRepo := productrepo.NewPostgresRepository(pool)
	productService := productusecase.NewProductService(productRepo)
	productServer := productcontroller.NewProductServer(productService)

	lomsRepo := orderrepo.NewPostgresRepository(pool)
	lomsService := lomsusecase.NewLomsService(lomsRepo, stocksRepo, lomsTransactor, notificationsClient, outboxRepo)
	lomsServer := lomscontroller.NewLomsServer(lomsService)

	api := controller.New(stocksServer, productServer, lomsServer)

	globalOutboxHandler := func(kind outboxrepo.Kind) (outbox.KindHandler, error) {
		switch kind {
		case outboxrepo.KindNotification:
			return lomsService.OrderStatusChangedNotificationKindHandler, nil

		default:
			return nil, fmt.Errorf("unsupported outboxCore kind")
		}
	}

	outbox := outbox.New(logger, outboxRepo, globalOutboxHandler, cfg, lomsTransactor)

	outbox.Start(ctx, cfg.Outbox.Workers, cfg.Outbox.BatchSize, cfg.Outbox.FetchPeriod, cfg.Outbox.TTL)

	go runGrpc(logger, cfg, api)
	runRest(ctx, logger, cfg)
}

func runRest(ctx context.Context, logger *zap.Logger, cfg *config.Config) {
	mux := runtime.NewServeMux()

	endpoint := "localhost:" + cfg.GRPC.Port
	options := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	if err := lompb.RegisterLomsHandlerFromEndpoint(ctx, mux, endpoint, options); err != nil {
		logger.Error("failed to register LOMS gateway", zap.Error(err))
		return
	}
	if err := productpb.RegisterProductServiceHandlerFromEndpoint(ctx, mux, endpoint, options); err != nil {
		logger.Error("failed to register Product gateway", zap.Error(err))
		return
	}
	if err := stockspb.RegisterStocksHandlerFromEndpoint(ctx, mux, endpoint, options); err != nil {
		logger.Error("failed to register Stocks gateway", zap.Error(err))
		return
	}

	gatewayPort := ":" + cfg.GRPC.GatewayPort
	handler := corsHandler(mux)

	if err := http.ListenAndServe(gatewayPort, handler); err != nil {
		logger.Error("gateway listen error", zap.Error(err))
	}
}

func runGrpc(logger *zap.Logger, cfg *config.Config, ctrl *controller.API) {
	lis, err := net.Listen("tcp", ":"+cfg.GRPC.Port)
	if err != nil {
		logger.Error("failed to listen for gRPC", zap.Error(err))
		return
	}

	grpcServer := grpc.NewServer()

	lompb.RegisterLomsServer(grpcServer, ctrl.LomsServer)
	productpb.RegisterProductServiceServer(grpcServer, ctrl.ProductServer)
	stockspb.RegisterStocksServer(grpcServer, ctrl.StockServer)

	logger.Info("gRPC server is listening", zap.String("port", cfg.GRPC.Port))

	if err := grpcServer.Serve(lis); err != nil {
		logger.Error("gRPC server failed", zap.Error(err))
	}
}

func corsHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == "" {
			origin = "http://localhost:5173"
		}
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept")
		w.Header().Set("Access-Control-Max-Age", "86400")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}
