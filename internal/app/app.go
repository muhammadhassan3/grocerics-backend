package app

import (
	"context"

	"grocerics-backend/internal/auth"
	"grocerics-backend/internal/config"
	"grocerics-backend/internal/dto"
	"grocerics-backend/internal/integration/quickcommerce"
	"grocerics-backend/internal/logging"
	"grocerics-backend/internal/middleware"
	"grocerics-backend/internal/migrate"
	"grocerics-backend/internal/repository"
	"grocerics-backend/internal/service"

	v1 "grocerics-backend/internal/route/v1"

	docs "grocerics-backend/docs"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	swaggerFile "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	firebase "firebase.google.com/go"
	"google.golang.org/api/option"
)

type App struct {
	Cfg         *config.Config
	DB          *gorm.DB
	JWTService  *auth.JWTService
	Router      *gin.Engine
	FirebaseApp *firebase.App
	AWSClient   *config.AWSClient

	UserRepo *repository.UserRepository
	QC       quickcommerce.Client

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
	a.initializeFirebase()

	client, err := config.NewAWSClient(cfg.AWS)
	if err != nil {
		return nil, err
	}
	a.AWSClient = client

	return a, nil
}

func (a *App) initializeFirebase() error {
	opt := option.WithAuthCredentialsFile(option.ServiceAccount, "grocerics-firebase-adminsdk.json")
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		zap.S().Errorw("failed to initialize Firebase", "error", err)
		return err
	}
	zap.S().Info("Firebase initialized")
	a.FirebaseApp = app
	return nil
}

func (a *App) buildRouter() *gin.Engine {
	r := gin.New()
	r.HandleMethodNotAllowed = true
	r.SetTrustedProxies([]string{"127.0.0.1"})

	_, err := logging.Init(a.Cfg.Env)
	if err != nil {
		zap.S().Fatalw("failed to initialize logging", "error", err)
	}
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
	v1.RegisterDashboardRoutes(a.JWTService, a.UserRepo, r)
	v1.RegisterInventoryManagementRoutes(r, v1.InventoryDeps{
		JWT:           a.JWTService,
		Users:         a.UserRepo,
		Products:      repository.NewProductRepository(a.DB),
		Variants:      repository.NewProductVariantRepository(a.DB),
		Categories:    repository.NewCategoryRepository(a.DB),
		Subcategories: repository.NewSubcategoryRepository(a.DB),
		Brands:        repository.NewBrandRepository(a.DB),
	})

	v1.RegisterPlatformRoutes(r, v1.PlatformDeps{
		JWT:       a.JWTService,
		Users:     a.UserRepo,
		Platforms: repository.NewPlatformRepository(a.DB),
	})

	pricingService := service.NewPricingService(a.DB)
	v1.RegisterLinkingRoutes(r, v1.LinkingDeps{
		JWT:            a.JWTService,
		Users:          a.UserRepo,
		Platforms:      repository.NewPlatformRepository(a.DB),
		Links:          repository.NewProductPlatformLinkRepository(a.DB),
		PlatformPrices: repository.NewPlatformPriceRepository(a.DB),
		Linking:        service.NewLinkingService(a.DB, a.QC, pricingService),
		Pricing:        pricingService,
	})

	v1.RegisterCategoryRoutes(r, v1.CategoryDeps{
		JWT:        a.JWTService,
		Users:      a.UserRepo,
		Categories: repository.NewCategoryRepository(a.DB),
	})
	v1.RegisterSubcategoryRoutes(r, v1.SubcategoryDeps{
		JWT:           a.JWTService,
		Users:         a.UserRepo,
		Subcategories: repository.NewSubcategoryRepository(a.DB),
		Categories:    repository.NewCategoryRepository(a.DB),
	})
	v1.RegisterBrandsRoutes(r, v1.BrandDeps{
		JWT:    a.JWTService,
		Users:  a.UserRepo,
		Brands: repository.NewBrandRepository(a.DB),
	})
	v1.RegisterBannerRoutes(r, v1.BannerDeps{
		JWT:     a.JWTService,
		Users:   a.UserRepo,
		Banners: repository.NewBannerRepository(a.DB),
	})
	v1.RegisterCityRoutes(r, v1.CityDeps{
		JWT:    a.JWTService,
		Users:  a.UserRepo,
		Cities: repository.NewCityRepository(a.DB),
	})
	v1.RegisterPresignedURLRoutes(r, a.JWTService, a.UserRepo)
	// Mobile-contract stub routes. cart/wishlist/product-detail were dropped in
	// the master merge — consumer.go implements those for real on the same paths.
	v1.RegisterAddressRoutes(r)
	v1.RegisterTopDealsRoutes(r)
	v1.RegisterSettingsRoutes(r)

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
