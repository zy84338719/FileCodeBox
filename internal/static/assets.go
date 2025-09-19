package static

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/zy84338719/filecodebox/internal/config"
)

// RegisterStaticRoutes registers public-facing static routes (assets, css, js, components)
func RegisterStaticRoutes(router *gin.Engine, cfg *config.ConfigManager) {
	themeDir := fmt.Sprintf("./%s", cfg.ThemesSelect)

	router.Static("/assets", fmt.Sprintf("%s/assets", themeDir))
	router.Static("/css", fmt.Sprintf("%s/css", themeDir))
	router.Static("/js", fmt.Sprintf("%s/js", themeDir))
	router.Static("/components", fmt.Sprintf("%s/components", themeDir))
}

// RegisterAdminStaticRoutes registers admin panel static routes (admin css/js/templates)
func RegisterAdminStaticRoutes(adminGroup *gin.RouterGroup, cfg *config.ConfigManager) {
	// Deprecated: do NOT register admin static routes publicly here.
	// Admin static assets are security-sensitive and must be served
	// through protected handlers (see internal/routes/admin.go) which
	// apply the required authentication middleware. Keeping this
	// function as a no-op avoids accidental public registration while
	// preserving the API for older callers.
	_ = adminGroup
	_ = cfg
	return
}

// ServeIndex serves the main index page with basic template replacements.
func ServeIndex(c *gin.Context, cfg *config.ConfigManager) {
	indexPath := filepath.Join(".", cfg.ThemesSelect, "index.html")

	content, err := os.ReadFile(indexPath)
	if err != nil {
		c.String(http.StatusNotFound, "Index file not found")
		return
	}

	html := string(content)
	// template replacements
	html = strings.ReplaceAll(html, "{{title}}", cfg.Base.Name)
	html = strings.ReplaceAll(html, "{{description}}", cfg.Base.Description)
	html = strings.ReplaceAll(html, "{{keywords}}", cfg.Base.Keywords)
	html = strings.ReplaceAll(html, "{{page_explain}}", cfg.PageExplain)
	html = strings.ReplaceAll(html, "{{opacity}}", fmt.Sprintf("%.1f", cfg.Opacity))
	html = strings.ReplaceAll(html, "src=\"js/", "src=\"/js/")
	html = strings.ReplaceAll(html, "href=\"css/", "href=\"/css/")
	html = strings.ReplaceAll(html, "src=\"assets/", "src=\"/assets/")
	html = strings.ReplaceAll(html, "href=\"assets/", "href=\"/assets/")
	html = strings.ReplaceAll(html, "src=\"components/", "src=\"/components/")
	html = strings.ReplaceAll(html, "{{background}}", cfg.Background)

	c.Header("Cache-Control", "no-cache")
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, html)
}

// ServeSetup serves the setup page with template replacements.
func ServeSetup(c *gin.Context, cfg *config.ConfigManager) {
	setupPath := filepath.Join(".", cfg.ThemesSelect, "setup.html")

	content, err := os.ReadFile(setupPath)
	if err != nil {
		c.String(http.StatusNotFound, "Setup page not found")
		return
	}

	html := string(content)
	html = strings.ReplaceAll(html, "{{title}}", cfg.Base.Name+" - 系统初始化")
	html = strings.ReplaceAll(html, "{{description}}", cfg.Base.Description)
	html = strings.ReplaceAll(html, "{{keywords}}", cfg.Base.Keywords)
	html = strings.ReplaceAll(html, "src=\"js/", "src=\"/js/")
	html = strings.ReplaceAll(html, "href=\"css/", "href=\"/css/")
	html = strings.ReplaceAll(html, "src=\"assets/", "src=\"/assets/")

	c.Header("Cache-Control", "no-cache")
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, html)
}

// ServeAdminPage serves the admin index page
func ServeAdminPage(c *gin.Context, cfg *config.ConfigManager) {
	adminPath := filepath.Join(".", cfg.ThemesSelect, "admin", "index.html")

	content, err := os.ReadFile(adminPath)
	if err != nil {
		c.String(http.StatusNotFound, "Admin page not found")
		return
	}

	c.Header("Cache-Control", "no-cache")
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, string(content))
}

// ServeUserPage serves user-facing static pages (login/register/dashboard/etc.)
func ServeUserPage(c *gin.Context, cfg *config.ConfigManager, pageName string) {
	userPagePath := filepath.Join(".", cfg.ThemesSelect, pageName)

	content, err := os.ReadFile(userPagePath)
	if err != nil {
		c.String(http.StatusNotFound, "User page not found: "+pageName)
		return
	}

	html := string(content)
	// normalize relative static paths to absolute paths so pages under /user/* load correctly
	html = strings.ReplaceAll(html, "src=\"js/", "src=\"/js/")
	html = strings.ReplaceAll(html, "href=\"css/", "href=\"/css/")
	html = strings.ReplaceAll(html, "src=\"assets/", "src=\"/assets/")
	html = strings.ReplaceAll(html, "href=\"assets/", "href=\"/assets/")

	c.Header("Cache-Control", "no-cache")
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, html)
}
