package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/zy84338719/filecodebox/internal/config"
	"github.com/zy84338719/filecodebox/internal/services"
)

// fakeUserService implements minimal ValidateToken for tests
type fakeUserService struct {
	claims *services.AuthClaims
	err    error
}

func (f *fakeUserService) ValidateToken(token string) (interface{}, error) {
	return f.claims, f.err
}

func TestCombinedAdminAuth_Success(t *testing.T) {
	g := gin.New()
	cfg := config.NewConfigManager()
	fsvc := &fakeUserService{claims: &services.AuthClaims{UserID: 1, Username: "admin", Role: "admin", SessionID: "s1"}, err: nil}
	g.Use(CombinedAdminAuth(cfg, fsvc))
	g.GET("/ok", func(c *gin.Context) { c.String(200, "ok") })

	req := httptest.NewRequest("GET", "/ok", nil)
	req.Header.Set("Authorization", "Bearer token")
	w := httptest.NewRecorder()
	g.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestCombinedAdminAuth_Fail(t *testing.T) {
	g := gin.New()
	cfg := config.NewConfigManager()
	fsvc := &fakeUserService{claims: nil, err: http.ErrNoCookie}
	g.Use(CombinedAdminAuth(cfg, fsvc))
	g.GET("/ok", func(c *gin.Context) { c.String(200, "ok") })

	req := httptest.NewRequest("GET", "/ok", nil)
	req.Header.Set("Authorization", "Bearer badtoken")
	w := httptest.NewRecorder()
	g.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}
