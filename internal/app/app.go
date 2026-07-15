package app

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/kkonst40/sso-service/internal/api/handler"
	"github.com/kkonst40/sso-service/internal/api/middleware"
	"github.com/kkonst40/sso-service/internal/config"
	pb "github.com/kkonst40/sso-service/internal/gen/user"
	sessionrepo "github.com/kkonst40/sso-service/internal/repo/session"
	userrepo "github.com/kkonst40/sso-service/internal/repo/user"
	"github.com/kkonst40/sso-service/internal/service/auth"
	"github.com/kkonst40/sso-service/internal/service/credvalidator"
	"github.com/kkonst40/sso-service/internal/service/eventbus"
	"github.com/kkonst40/sso-service/internal/service/pwdhasher"
	userservice "github.com/kkonst40/sso-service/internal/service/user"
	"google.golang.org/grpc"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type App struct {
	httpServer *http.Server
	grpcServer *grpc.Server
	grpcPort   string
	db         *sql.DB
}

func New(cfg *config.Config) (*App, error) {
	db, err := SetupDB(cfg.DB.User, cfg.DB.Password, cfg.DB.Host, cfg.DB.Port, cfg.DB.DBName)
	if err != nil {
		return nil, err
	}

	eventProducer, err := eventbus.NewProducer(cfg)
	if err != nil {
		return nil, err
	}

	var (
		jwtProvider   = auth.NewJWTProvider(cfg)
		pwdHasher     = pwdhasher.New()
		credValidator = credvalidator.New(cfg)
	)

	var (
		userRepo    = userrepo.New(db)
		sessionRepo = sessionrepo.New(db)
		userService = userservice.New(
			jwtProvider,
			pwdHasher,
			credValidator,
			eventProducer,
			userRepo,
			sessionRepo,
		)
		userHandler = handler.New(userService, cfg)
	)

	router := http.NewServeMux()

	noAuthStack := middleware.CreateStack(
		middleware.Recovery,
		middleware.Timeout(3*time.Second),
	)

	authStack := middleware.CreateStack(
		middleware.Recovery,
		middleware.Timeout(3*time.Second),
		middleware.Auth(jwtProvider, cfg.JWT.CookieName),
	)

	router.Handle("POST /login", noAuthStack(http.HandlerFunc(userHandler.Login)))
	router.Handle("POST /register", noAuthStack(http.HandlerFunc(userHandler.Create)))

	router.Handle("GET /me", authStack(http.HandlerFunc(userHandler.Me)))
	router.Handle("POST /logout", authStack(http.HandlerFunc(userHandler.Logout)))
	router.Handle("PUT /updatelogin", authStack(http.HandlerFunc(userHandler.UpdateLogin)))
	router.Handle("PUT /updatepassword", authStack(http.HandlerFunc(userHandler.UpdatePassword)))
	router.Handle("DELETE /{id}", authStack(http.HandlerFunc(userHandler.Delete)))

	httpServer := &http.Server{
		Addr:    ":" + cfg.RestPort,
		Handler: router,
	}

	grpcServer := grpc.NewServer()
	userGRPC := handler.NewUserGRPCHandler(userService)
	pb.RegisterUserServiceServer(grpcServer, userGRPC)

	return &App{
		httpServer: httpServer,
		grpcServer: grpcServer,
		grpcPort:   cfg.GrpcPort,
		db:         db,
	}, nil
}

func (a *App) Run() error {
	errChan := make(chan error, 2)

	go func() {
		if err := a.httpServer.ListenAndServe(); err != nil {
			errChan <- fmt.Errorf("HTTP serve error: %w", err)
		}
	}()

	go func() {
		grpcListener, err := net.Listen("tcp", ":"+a.grpcPort)
		if err != nil {
			errChan <- fmt.Errorf("gRPC listener error: %w", err)
			return
		}

		if err := a.grpcServer.Serve(grpcListener); err != nil {
			errChan <- fmt.Errorf("gRPC serve error: %w", err)
		}
	}()

	return <-errChan
}

func (a *App) Shutdown(ctx context.Context) {
	a.grpcServer.GracefulStop()

	if err := a.httpServer.Shutdown(ctx); err != nil {
		log.Println("Server forced to shutdown", "error", err.Error())
	}

	if err := a.db.Close(); err != nil {
		log.Println("DB close error", "error", err.Error())
	}
}

func SetupDB(user, pwd, host, port, dbName string) (*sql.DB, error) {
	dbUrl := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", user, pwd, host, port, dbName)

	db, err := sql.Open("pgx", dbUrl)
	if err != nil {
		return nil, fmt.Errorf("Error creating db object: %v", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("Failed to connect to the database: %v", err)
	}

	return db, nil
}
