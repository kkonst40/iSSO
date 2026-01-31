package app

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/kkonst40/isso/internal/config"
	pb "github.com/kkonst40/isso/internal/gen/user"
	"github.com/kkonst40/isso/internal/handler"
	"github.com/kkonst40/isso/internal/middleware"
	"github.com/kkonst40/isso/internal/repo"
	"github.com/kkonst40/isso/internal/service"
	"github.com/kkonst40/isso/internal/utils"
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
	db, err := SetupDB(cfg.DB.User, cfg.DB.Password, cfg.DB.Host, cfg.DB.DBName)
	if err != nil {
		return nil, err
	}

	var (
		jwtProvider   = utils.NewJWTProvider(cfg)
		pwdHasher     = utils.NewPasswordHandler()
		credValidator = utils.NewValidator(cfg)
	)

	var (
		userRepo    = repo.New(db)
		userService = service.New(jwtProvider, pwdHasher, credValidator, userRepo, uuid.UUID{})
		userHandler = handler.New(userService, cfg)
	)

	mux := http.NewServeMux()

	// for test
	mux.HandleFunc("GET /r", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/register.html")
	})
	mux.HandleFunc("GET /l", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/login.html")
	})
	mux.HandleFunc("GET /checkauth", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/me.html")
	})

	mux.HandleFunc("GET /all", userHandler.All)
	mux.HandleFunc("GET /me", middleware.Auth(userHandler.Me, jwtProvider))
	mux.HandleFunc("POST /exist", userHandler.Exist)
	mux.HandleFunc("POST /login", userHandler.Login)
	mux.HandleFunc("POST /logout", middleware.Auth(userHandler.Logout, jwtProvider))
	mux.HandleFunc("POST /register", userHandler.Create)
	mux.HandleFunc("PUT /updatelogin", middleware.Auth(userHandler.UpdateLogin, jwtProvider))
	mux.HandleFunc("PUT /updatepassword", middleware.Auth(userHandler.UpdatePassword, jwtProvider))
	mux.HandleFunc("DELETE /{id}", middleware.Auth(userHandler.Delete, jwtProvider))

	httpServer := &http.Server{
		Addr:    ":" + cfg.HttpPort,
		Handler: middleware.Timeout(mux, 3*time.Second),
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

func SetupDB(user, pwd, host, dbName string) (*sql.DB, error) {
	dbUrl := fmt.Sprintf("postgres://%s:%s@%s/%s", user, pwd, host, dbName)

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
