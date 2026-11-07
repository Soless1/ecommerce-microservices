package app

import (
	"context"
	"net"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/config"
	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/controller"
	lompb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/loms/api/loms/v1"
	productpb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/loms/api/product/v1"
	stockspb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/loms/api/stocks/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	stockscontroller "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/controller/stocks"
	stocksrepo "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/repository/stocks"
	stocksusecase "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/usercase/stocks"

	productcontroller "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/controller/product"
	productrepo "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/repository/product"
	productusecase "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/usercase/product"

	lomscontroller "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/controller/loms"
	orderrepo "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/repository/order"
	lomsusecase "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/usercase/loms"
)

func Run(logger *zap.Logger, cfg *config.Config) {

	stocksRepo := stocksrepo.NewInMemoryRepository()
	stocksService := stocksusecase.NewStocksService(stocksRepo)
	stocksServer := stockscontroller.NewStocksServer(stocksService)

	productRepo := productrepo.NewInMemoryRepository()
	productService := productusecase.NewProductService(productRepo)
	productServer := productcontroller.NewProductServer(productService)

	lomsRepo := orderrepo.NewInMemoryRepository()
	lomsService := lomsusecase.NewLomsService(lomsRepo, stocksRepo)
	lomsServer := lomscontroller.NewLomsServer(lomsService)

	api := controller.New(stocksServer, productServer, lomsServer)

	ctx := context.Background()
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
