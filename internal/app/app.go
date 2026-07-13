package app

import (
	"grocerics-backend/internal/auth"
	"grocerics-backend/internal/config"
	"grocerics-backend/internal/dto"
	"grocerics-backend/internal/integration/quickcommerce"
	"grocerics-backend/internal/logging"
	"grocerics-backend/internal/middleware"
	"grocerics-backend/internal/migrate"
	"grocerics-backend/internal/repository"
	v1 "grocerics-backend/internal/route/v1"
	"grocerics-backend/internal/service"

	docs "grocerics-backend/docs"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	swaggerFile "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type App struct {
	Cfg        *config.Config
	DB         *gorm.DB
	JWTService *auth.JWTService
	Router     *gin.Engine
	UserRepo   *repository.UserRepository
	QC         quickcommerce.Client

	AuthService *service.AuthService
}

func New(cfg *config.Config) (*App, error) {
	db, err := config.ConnectDB(cfg.DB)
	if err != nil {
		return nil, err
	}
	zap.S().Info("database connected")

	if err := migrate.Up(db); err != nil {
		return nil, err
	}
	zap.S().Info("migrations applied")

	jwt := auth.NewJWTService(cfg.JWT.SecretKey)
	userRepo := repository.NewUserRepository(db)
	qc := quickcommerce.New(quickcommerce.Config{APIKey: cfg.QC.APIKey, BaseURL: cfg.QC.BaseURL})

	authService := service.NewAuthService(
		userRepo,
		repository.NewRefreshTokenRepository(db),
		repository.NewPasswordResetRepository(db),
		jwt,
		"",
	)

	config.SeedAdmin(db, cfg.Seed, cfg.Env)
	config.SeedDemo(db, cfg.Env)

	a := &App{
		Cfg:         cfg,
		DB:          db,
		JWTService:  jwt,
		UserRepo:    userRepo,
		QC:          qc,
		AuthService: authService,
	}
	a.Router = a.buildRouter()
	return a, nil
}

func (a *App) buildRouter() *gin.Engine {
	r := gin.New()
	r.HandleMethodNotAllowed = true
	r.SetTrustedProxies([]string{"127.0.0.1"})

	// Init installs the global zap logger, it never returns nil, so there's no erro to check here
	logging.Init(a.Cfg.Env)
	r.Use(middleware.RequestID())
	r.Use(middleware.CORS(a.Cfg.FrontendURL))
	r.Use(middleware.SecurityHeaders())
	r.Use(middleware.ErrorHandler())
	r.Use(middleware.GinZapLogger(logging.Log))
	r.Use(middleware.GinZapRecovery(logging.Log))

	r.NoRoute(func(c *gin.Context) {
		c.JSON(404, dto.Response{Status: "failed", Code: "ROUTE_NOT_FOUND", Message: "Route not found"})
	})
	r.NoMethod(func(c *gin.Context) {
		c.JSON(405, dto.Response{Status: "failed", Code: "METHOD_NOT_ALLOWED", Message: "Method not allowed"})
	})

	v1.RegisterAuthRoutes(r, a.AuthService, a.JWTService, a.UserRepo)
	v1.RegisterUserRoutes(r, a.JWTService, a.UserRepo)
	v1.RegisterBannerRoutes(a.JWTService, a.UserRepo, r)

	rg := r.Group("/")
	v1.RegisterDashboardRoutes(a.JWTService, a.UserRepo, rg)
	v1.RegisterInventoryManagementRoutes(a.JWTService, a.UserRepo, rg)
	v1.RegisterBrandsRoutes(a.JWTService, a.UserRepo, rg)
	v1.RegisterCategoryRoutes(a.JWTService, a.UserRepo, rg)
	v1.RegisterSubcategoryRoutes(a.JWTService, a.UserRepo, rg)

	v1.RegisterConsumerRoutes(r, v1.ConsumerDeps{
		JWT:     a.JWTService,
		Users:   a.UserRepo,
		Cities:  repository.NewCityRepository(a.DB),
		Catalog: service.NewCatalogService(a.DB),
		Cart:    service.NewCartService(a.DB),
		Loc:     service.NewLocationResolver(a.DB),
	})

	v1.RegisterProfileRoutes(r, v1.ProfileDeps{
		JWT:     a.JWTService,
		Users:   a.UserRepo,
		Profile: service.NewProfileService(a.DB),
	})

	docs.SwaggerInfo.BasePath = "/"
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFile.Handler))
	return r
}

func (a *App) Run() error {
	addr := ":" + a.Cfg.Port
	zap.S().Infow("starting server", "addr", addr, "env", a.Cfg.Env)
	return a.Router.Run(addr)
}
