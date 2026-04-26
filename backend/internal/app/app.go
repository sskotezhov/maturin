package app

import (
	"context"
	"log"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/redis/go-redis/v9"
	echoSwagger "github.com/swaggo/echo-swagger"
	"gorm.io/gorm"

	"github.com/sskotezhov/maturin/config"
	"github.com/sskotezhov/maturin/internal/auth"
	"github.com/sskotezhov/maturin/internal/product"
	"github.com/sskotezhov/maturin/internal/user"
	"github.com/sskotezhov/maturin/pkg/email"
	mw "github.com/sskotezhov/maturin/pkg/middleware"
	"github.com/sskotezhov/maturin/pkg/onec"
)

type App struct {
	Echo   *echo.Echo
	DB     *gorm.DB
	Redis  *redis.Client
	Config *config.Config
}

func New(cfg *config.Config, db *gorm.DB, rdb *redis.Client) (*App, error) {
	e := echo.New()
	e.HideBanner = true

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{
			"http://localhost:3000",
			"http://93.77.160.169",
		},
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Content-Type", "Authorization"},
	}))

	if err := migrate(db); err != nil {
		return nil, err
	}

	jwtCfg := cfg.Maturin.JWT
	smtpCfg := cfg.Maturin.SMTP
	oneCCfg := cfg.Maturin.OneC

	// user
	userRepo := user.NewRepository(db)

	// email
	emailSender := email.NewSender(email.Config{
		Host:     smtpCfg.Host,
		Port:     smtpCfg.Port,
		User:     smtpCfg.User,
		Password: smtpCfg.Password,
	})

	// auth
	authSvc := auth.NewService(
		userRepo,
		rdb,
		emailSender,
		jwtCfg.Secret,
		time.Duration(jwtCfg.AccessTokenTTL)*time.Minute,
		time.Duration(jwtCfg.RefreshTokenTTL)*time.Minute,
	)

	// product
	oneCClient := onec.NewClient(oneCCfg.BaseURL, oneCCfg.User, oneCCfg.Password)
	productRepo := product.NewRepository(oneCClient, rdb, time.Duration(oneCCfg.CacheTTLMin)*time.Minute)
	productSvc := product.NewService(productRepo)
	productHandler := product.NewHandler(productSvc)

	// swagger
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	userSvc := user.NewService(userRepo)

	api := e.Group("/api/v1")
	auth.NewHandler(authSvc).Register(api.Group("/auth"))
	productHandler.Register(api.Group("/products"))
	productHandler.RegisterCategories(api.Group("/categories"))

	authed := api.Group("", mw.JWTAuth(jwtCfg.Secret))
	user.NewHandler(userSvc).Register(authed.Group("/user"))

	// прогрев кеша в фоне
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()
		if _, err := productRepo.GetAll(ctx); err != nil {
			log.Printf("catalog warmup error: %v", err)
		}
	}()

	return &App{
		Echo:   e,
		DB:     db,
		Redis:  rdb,
		Config: cfg,
	}, nil
}

func migrate(db *gorm.DB) error {
	return user.Migrate(db)
}
