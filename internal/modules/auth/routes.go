package auth

import (
	"go-starter/internal/modules/auth/handlers"
	"go-starter/internal/modules/auth/middlewares"
	"go-starter/internal/modules/auth/repositories"
	"go-starter/internal/modules/auth/services"
	"go-starter/pkg/config"
	"go-starter/pkg/db"

	"github.com/labstack/echo/v4"
)

func RegisterRoutes(e *echo.Echo, api *echo.Group, database *db.Database, cfg *config.Config) {

	authRepo := repositories.NewRepository(database)
	jwtService := services.NewJWTService()
	authService := services.NewAuthService(authRepo, jwtService, cfg)
	authAPIHandler := handlers.NewAuthAPIHandler(authService)
	authWEBHandler := handlers.NewAuthWEBHandler(authService)

	e.GET("/login", authWEBHandler.ViewLogin)
	e.POST("/login", authWEBHandler.Login)
	e.GET("/register", authWEBHandler.ViewRegister)
	e.POST("/register", authWEBHandler.Register)
	e.POST("/logout", authWEBHandler.Logout)
	e.GET("/", authWEBHandler.HandleRoot)

	authGroup := api.Group("/auth")
	authGroup.POST("/register", authAPIHandler.Register)
	authGroup.POST("/login", authAPIHandler.Login)

	api.GET("/profile", authAPIHandler.GetProfile, middlewares.JWTMiddleware(jwtService))
}
