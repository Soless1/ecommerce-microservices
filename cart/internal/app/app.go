package app

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os/signal"
	"syscall"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	adapterloms "github.com/igoroutine-courses/microservices.ecommerce.cart/internal/adapter/loms/grpc"
	adapterproduct "github.com/igoroutine-courses/microservices.ecommerce.cart/internal/adapter/product/grpc"
	adapterstocks "github.com/igoroutine-courses/microservices.ecommerce.cart/internal/adapter/stocks/grpc"
	"github.com/igoroutine-courses/microservices.ecommerce.cart/internal/config"
	"github.com/igoroutine-courses/microservices.ecommerce.cart/internal/controller"
	cartcontroller "github.com/igoroutine-courses/microservices.ecommerce.cart/internal/controller/cart"
	cartrepo "github.com/igoroutine-courses/microservices.ecommerce.cart/internal/repository/cart"
	"github.com/igoroutine-courses/microservices.ecommerce.cart/internal/repository/transactor"
	cartusecase "github.com/igoroutine-courses/microservices.ecommerce.cart/internal/usecase/cart"
	itemusecase "github.com/igoroutine-courses/microservices.ecommerce.cart/internal/usecase/item"
	cartmigrations "github.com/igoroutine-courses/microservices.ecommerce.cart/migrations"
	cartpb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/cart/api/cart/v1"
	lomspb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/loms/api/loms/v1"
	productpb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/loms/api/product/v1"
	stockspb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/loms/api/stocks/v1"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func Run(logger *zap.Logger, cfg *config.Config) {
	signalCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	lomsConn, err := grpc.NewClient(
		cfg.Clients.LOMSGrpcAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		logger.Fatal("failed to connect to server", zap.Error(err))
	}

	defer func(conn *grpc.ClientConn) {
		_ = conn.Close()
	}(lomsConn)

	productClient := adapterproduct.NewProductClient(
		productpb.NewProductServiceClient(lomsConn),
	)
	lomsClient := adapterloms.NewLOMSClient(
		lomspb.NewLomsClient(lomsConn),
	)
	stocksClient := adapterstocks.NewStocksClient(
		stockspb.NewStocksClient(lomsConn),
	)

	group, ctx := errgroup.WithContext(signalCtx)

	pool, err := newPostgresPool(ctx, cfg)
	if err != nil {
		logger.Error("failed to initialize PostgreSQL", zap.Error(err))
		return
	}
	defer pool.Close()

	if err := cartmigrations.Run(pool); err != nil {
		logger.Error("failed to run Cart migrations", zap.Error(err))
		return
	}
	cartTransactor := transactor.NewTransactor(pool)

	cartRepo := cartrepo.NewPostgresRepository(pool)
	cartService := cartusecase.NewCartService(cartRepo, productClient, lomsClient, cartTransactor)
	itemService := itemusecase.NewItemService(cartRepo, productClient, stocksClient)

	cartServer := cartcontroller.NewCartServer(itemService, cartService)
	api := controller.New(cartServer)

	group.Go(func() error {
		return runGrpc(ctx, logger, cfg, api)
	})
	group.Go(func() error {
		return runRest(ctx, cfg)
	})
	if err := group.Wait(); err != nil {
		logger.Error("Cart server stopped with an error", zap.Error(err))
	}
}

func runRest(ctx context.Context, cfg *config.Config) error {
	mux := runtime.NewServeMux()

	endpoint := cfg.GRPC.Endpoint
	options := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	if err := cartpb.RegisterCartHandlerFromEndpoint(ctx, mux, endpoint, options); err != nil {
		return err
	}

	gatewayPort := ":" + cfg.GRPC.GatewayPort
	server := &http.Server{
		Addr:    gatewayPort,
		Handler: corsHandler(mux),
	}
	shutdownErr := make(chan error, 1)
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.GRPC.GatewayShutdownTimeout)
		defer cancel()
		shutdownErr <- server.Shutdown(shutdownCtx)
	}()

	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	if ctx.Err() != nil {
		if err := <-shutdownErr; err != nil {
			return err
		}
	}
	return nil
}

func runGrpc(ctx context.Context, logger *zap.Logger, cfg *config.Config, ctrl *controller.API) error {
	lis, err := net.Listen("tcp", ":"+cfg.GRPC.Port)
	if err != nil {
		return err
	}
	grpcServer := grpc.NewServer()

	cartpb.RegisterCartServer(grpcServer, ctrl.Cart)

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
