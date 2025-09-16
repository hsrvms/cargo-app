package handlers

import (
	"go-starter/internal/modules/auth/dto"
	"go-starter/internal/modules/auth/services"
	"go-starter/internal/modules/auth/views"
	"go-starter/internal/modules/auth/views/errors"
	"log"
	"net/http"

	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

type AuthWEBHandler struct {
	authService *services.AuthService
	validator   *validator.Validate
}

func NewAuthWEBHandler(authService *services.AuthService) *AuthWEBHandler {
	return &AuthWEBHandler{
		authService: authService,
		validator:   validator.New(),
	}
}

func (h *AuthWEBHandler) ViewLogin(c echo.Context) error {
	component := views.LoginPage()
	return component.Render(c.Request().Context(), c.Response().Writer)
}

func (h *AuthWEBHandler) ViewRegister(c echo.Context) error {
	component := views.RegisterPage()
	return component.Render(c.Request().Context(), c.Response().Writer)
}

func (h *AuthWEBHandler) Login(c echo.Context) error {

	req := dto.LoginRequest{
		Email:    strings.ToLower(strings.TrimSpace(c.FormValue("email"))),
		Password: c.FormValue("password"),
	}

	if err := h.validator.Struct(req); err != nil {
		return h.renderLoginWithError(c, http.StatusBadRequest, "Invalid email or password")
	}

	response, err := h.authService.Login(c.Request().Context(), &req)
	if err != nil {
		statusCode, message := h.categorizeAuthError(err)
		return h.renderLoginWithError(c, statusCode, message)
	}

	return h.handleAuthSuccess(c, response.Token, http.StatusOK)
}

func (h *AuthWEBHandler) Register(c echo.Context) error {
	req := dto.RegisterRequest{
		FirstName: strings.TrimSpace(c.FormValue("firstname")),
		LastName:  strings.TrimSpace(c.FormValue("lastname")),
		Email:     strings.ToLower(strings.TrimSpace(c.FormValue("email"))),
		Password:  c.FormValue("password"),
	}

	if err := h.validator.Struct(req); err != nil {
		return h.renderRegisterWithError(c, http.StatusBadRequest, "Invalid email or password")
	}

	response, err := h.authService.Register(c.Request().Context(), &req)
	if err != nil {
		statusCode, message := h.categorizeAuthError(err)
		return h.renderRegisterWithError(c, statusCode, message)
	}

	return h.handleAuthSuccess(c, response.Token, http.StatusCreated)
}

func (h *AuthWEBHandler) Logout(c echo.Context) error {
	// Clear the auth cookie
	cookie := &http.Cookie{
		Name:     "auth_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // make it true on https
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1, // Delete the cookie
	}
	c.SetCookie(cookie)

	// Redirect to login page with full page reload
	return c.Redirect(http.StatusSeeOther, "/login")
}

func (h *AuthWEBHandler) GetUserInfo(c echo.Context) (*services.Claims, error) {
	// Get the auth token from cookie
	cookie, err := c.Cookie("auth_token")
	if err != nil {
		return nil, err
	}

	// Create a temporary JWT service to validate the token
	jwtService := services.NewJWTService()
	claims, err := jwtService.ValidateToken(cookie.Value)
	if err != nil {
		return nil, err
	}

	return claims, nil
}

func (h *AuthWEBHandler) HandleRoot(c echo.Context) error {
	log.Println("Root route accessed, checking authentication...")

	// Check if user is already authenticated
	cookie, err := c.Cookie("auth_token")
	if err != nil {
		log.Println("No auth token found → redirecting to login")
		return c.Redirect(http.StatusSeeOther, "/login")
	}

	// Validate the token
	jwtService := services.NewJWTService()
	claims, err := jwtService.ValidateToken(cookie.Value)
	if err != nil {
		log.Printf("Invalid auth token → redirecting to login: %v", err)
		return c.Redirect(http.StatusSeeOther, "/login")
	}

	log.Printf("User %s authenticated → redirecting to shipments", claims.Email)
	return c.Redirect(http.StatusSeeOther, "/shipments")
}

func (h *AuthWEBHandler) setAuthCokie(c echo.Context, token string) {
	cookie := &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // make it true on https
		SameSite: http.SameSiteStrictMode,
		MaxAge:   360000,
	}
	c.SetCookie(cookie)
}

func (h *AuthWEBHandler) renderError(c echo.Context, statusCode int, message string) error {
	c.Response().Header().Set("Content-Type", "text/html")
	c.Response().WriteHeader(statusCode)
	return errors.AuthError(message).Render(c.Request().Context(), c.Response().Writer)
}

func (h *AuthWEBHandler) renderLoginWithError(c echo.Context, statusCode int, message string) error {
	c.Response().Header().Set("Content-Type", "text/html")
	c.Response().WriteHeader(statusCode)
	component := views.LoginPageWithError(message)
	return component.Render(c.Request().Context(), c.Response().Writer)
}

func (h *AuthWEBHandler) renderRegisterWithError(c echo.Context, statusCode int, message string) error {
	c.Response().Header().Set("Content-Type", "text/html")
	c.Response().WriteHeader(statusCode)
	component := views.RegisterPageWithError(message)
	return component.Render(c.Request().Context(), c.Response().Writer)
}

func (h *AuthWEBHandler) handleAuthSuccess(c echo.Context, token string, statusCode int) error {
	h.setAuthCokie(c, token)
	// Use proper HTTP redirect for full page reload to ensure CSS loads correctly
	return c.Redirect(http.StatusSeeOther, "/shipments")
}

func (h *AuthWEBHandler) categorizeAuthError(err error) (int, string) {
	errStr := err.Error()

	switch {
	case strings.Contains(errStr, "invalid credentials") || strings.Contains(errStr, "user not found"):
		return http.StatusUnauthorized, "Invalid email or password"
	case strings.Contains(errStr, "already exists"):
		return http.StatusConflict, "User with this email already exists"
	case strings.Contains(errStr, "maximum number of users reached"):
		return http.StatusForbidden, "Registration limit reached. Maximum number of users allowed has been exceeded."
	default:
		return http.StatusInternalServerError, "An error occurred. Please try again"
	}
}
