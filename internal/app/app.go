package app

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/kkonst40/isso/internal/config"
	"github.com/kkonst40/isso/internal/handler"
	"github.com/kkonst40/isso/internal/middleware"
	"github.com/kkonst40/isso/internal/repo"
	"github.com/kkonst40/isso/internal/service"
	"github.com/kkonst40/isso/internal/utils"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type App struct {
	server *http.Server
	db     *sql.DB
}

func New(cfg *config.Config) (*App, error) {
	db, err := NewDB(cfg.DB.User, cfg.DB.Password, cfg.DB.Host, cfg.DB.DBName)
	if err != nil {
		return nil, err
	}

	jwtProvider := utils.NewJWTProvider(cfg)
	pwdHasher := utils.NewPasswordHandler()

	userRepo := repo.New(db)
	userService := service.New(jwtProvider, pwdHasher, userRepo, uuid.UUID{})
	userHandler := handler.New(userService, cfg)

	mux := http.NewServeMux()

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
		Addr:    ":" + cfg.Port,
		Handler: mux,
	}

	return &App{
		server: httpServer,
		db:     db,
	}, nil
}

func (a *App) Run() error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			os.Exit(1)
		}
	}()
	log.Println("Server started", "address", a.server.Addr)

	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := a.server.Shutdown(ctx); err != nil {
		log.Println("Server forced to shutdown", "error", err.Error())
	}
	if err := a.db.Close(); err != nil {
		log.Println("DB close error", "error", err.Error())
	}

	log.Println("Server exiting")
	return nil
}

func NewDB(user, pwd, host, dbName string) (*sql.DB, error) {
	dbUrl := fmt.Sprintf("postgres://%v:%v@%v/%v",
		user, pwd, host, dbName)

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
