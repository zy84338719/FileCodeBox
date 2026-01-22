package static

import (
	"embed"
	"fmt"
	"html"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/zy84338719/filecodebox/internal/config"
)

const defaultThemeDir = "themes/2025"

// embeddedFS holds the embedded themes filesystem, set via SetEmbeddedFS
var embeddedFS embed.FS
var embeddedFSSet bool

// SetEmbeddedFS sets the embedded filesystem for static assets
func SetEmbeddedFS(efs embed.FS) {
	embeddedFS = efs
	embeddedFSSet = true
}

// getEmbeddedSubFS returns a sub-filesystem for the themes directory
func getEmbeddedSubFS() (fs.FS, error) {
	if !embeddedFSSet {
		return nil, fs.ErrNotExist
	}
	return fs.Sub(embeddedFS, "themes/2025")
}

func themeCandidates(cfg *config.ConfigManager) []string {
	var candidates []string
	seen := make(map[string]struct{})
	add := func(path string) {
		path = strings.TrimSpace(path)
		if path == "" {
			return
		}
		if _, ok := seen[path]; ok {
			return
		}
		seen[path] = struct{}{}
		candidates = append(candidates, path)
	}

	if cfg != nil && cfg.UI != nil {
		add(cfg.UI.ThemesSelect)
	}
	add(defaultThemeDir)
	return candidates
}

func themeDirExists(dir string) bool {
	info, err := os.Stat(dir)
	return err == nil && info.IsDir()
}

func firstExistingThemeDir(cfg *config.ConfigManager) string {
	for _, candidate := range themeCandidates(cfg) {
		if filepath.IsAbs(candidate) {
			if themeDirExists(candidate) {
				return candidate
			}
			continue
		}
		if themeDirExists(candidate) {
			return candidate
		}
	}
	return defaultThemeDir
}

func resolveThemeFilePath(cfg *config.ConfigManager, parts ...string) (string, error) {
	var firstErr error
	for _, candidate := range themeCandidates(cfg) {
		pathParts := append([]string{candidate}, parts...)
		path := filepath.Join(pathParts...)
		info, err := os.Stat(path)
		if err == nil {
			if !info.IsDir() {
				return path, nil
			}
			continue
		}
		if firstErr == nil {
			firstErr = err
		}
	}
	if firstErr == nil {
		firstErr = os.ErrNotExist
	}
	return "", firstErr
}

func loadThemeFile(cfg *config.ConfigManager, parts ...string) ([]byte, error) {
	// First try filesystem
	path, err := resolveThemeFilePath(cfg, parts...)
	if err == nil {
		return os.ReadFile(path)
	}
	// Fallback to embedded FS
	if embeddedFSSet {
		embeddedPath := filepath.Join(parts...)
		data, embErr := fs.ReadFile(embeddedFS, "themes/2025/"+embeddedPath)
		if embErr == nil {
			return data, nil
		}
	}
	return nil, err
}

// ResolveThemeFile returns the concrete filesystem path for a theme file, applying fallbacks.
func ResolveThemeFile(cfg *config.ConfigManager, parts ...string) (string, error) {
	return resolveThemeFilePath(cfg, parts...)
}

// ServeThemeFile serves a theme file, falling back to embedded FS if filesystem path not found
func ServeThemeFile(c *gin.Context, cfg *config.ConfigManager, parts ...string) {
	// First try filesystem
	path, err := resolveThemeFilePath(cfg, parts...)
	if err == nil {
		c.File(path)
		return
	}
	// Fallback to embedded FS
	if embeddedFSSet {
		embeddedPath := "themes/2025/" + filepath.Join(parts...)
		data, embErr := fs.ReadFile(embeddedFS, embeddedPath)
		if embErr == nil {
			// Detect content type
			contentType := "application/octet-stream"
			ext := filepath.Ext(embeddedPath)
			switch ext {
			case ".html":
				contentType = "text/html; charset=utf-8"
			case ".css":
				contentType = "text/css; charset=utf-8"
			case ".js":
				contentType = "application/javascript; charset=utf-8"
			case ".json":
				contentType = "application/json; charset=utf-8"
			case ".svg":
				contentType = "image/svg+xml"
			case ".png":
				contentType = "image/png"
			case ".jpg", ".jpeg":
				contentType = "image/jpeg"
			case ".gif":
				contentType = "image/gif"
			case ".ico":
				contentType = "image/x-icon"
			case ".woff":
				contentType = "font/woff"
			case ".woff2":
				contentType = "font/woff2"
			case ".ttf":
				contentType = "font/ttf"
			}
			c.Data(http.StatusOK, contentType, data)
			return
		}
	}
	c.Status(http.StatusNotFound)
}

// ThemePath returns the resolved theme directory joined with optional relative parts.
func ThemePath(cfg *config.ConfigManager, parts ...string) string {
	root := firstExistingThemeDir(cfg)
	if len(parts) == 0 {
		return root
	}
	pathParts := append([]string{root}, parts...)
	return filepath.Join(pathParts...)
}

// RegisterStaticRoutes registers public-facing static routes (assets, css, js, components)
func RegisterStaticRoutes(router *gin.Engine, cfg *config.ConfigManager) {
	themeDir := firstExistingThemeDir(cfg)
	
	// Check if filesystem theme exists
	if themeDirExists(themeDir) {
		router.Static("/assets", filepath.Join(themeDir, "assets"))
		router.Static("/css", filepath.Join(themeDir, "css"))
		router.Static("/js", filepath.Join(themeDir, "js"))
		router.Static("/components", filepath.Join(themeDir, "components"))
		return
	}
	
	// Fallback to embedded FS
	if embeddedFSSet {
		registerEmbeddedStatic(router, "/assets", "themes/2025/assets")
		registerEmbeddedStatic(router, "/css", "themes/2025/css")
		registerEmbeddedStatic(router, "/js", "themes/2025/js")
		registerEmbeddedStatic(router, "/components", "themes/2025/components")
	}
}

// registerEmbeddedStatic registers a static route from embedded FS
func registerEmbeddedStatic(router *gin.Engine, urlPath, fsPath string) {
	subFS, err := fs.Sub(embeddedFS, fsPath)
	if err != nil {
		return
	}
	router.StaticFS(urlPath, http.FS(subFS))
}

// Note: admin static routes are intentionally not registered here.
// Admin-specific assets must be served through protected handlers
// in `internal/routes/admin.go` where authentication middleware is
// applied. This avoids accidentally exposing admin-only files via
// public `router.Static` registrations.

// ServeIndex serves the main index page with basic template replacements.
func ServeIndex(c *gin.Context, cfg *config.ConfigManager) {
	content, err := loadThemeFile(cfg, "index.html")
	if err != nil {
		html := fallbackIndexHTML(cfg)
		c.Header("Cache-Control", "no-cache")
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(http.StatusOK, html)
		return
	}

	html := string(content)
	// template replacements
	html = strings.ReplaceAll(html, "{{title}}", cfg.Base.Name)
	html = strings.ReplaceAll(html, "{{description}}", cfg.Base.Description)
	html = strings.ReplaceAll(html, "{{keywords}}", cfg.Base.Keywords)
	html = strings.ReplaceAll(html, "{{page_explain}}", cfg.UI.PageExplain)
	html = strings.ReplaceAll(html, "{{opacity}}", fmt.Sprintf("%.1f", cfg.UI.Opacity))
	html = strings.ReplaceAll(html, "src=\"js/", "src=\"/js/")
	html = strings.ReplaceAll(html, "href=\"css/", "href=\"/css/")
	html = strings.ReplaceAll(html, "src=\"assets/", "src=\"/assets/")
	html = strings.ReplaceAll(html, "href=\"assets/", "href=\"/assets/")
	html = strings.ReplaceAll(html, "src=\"components/", "src=\"/components/")
	html = strings.ReplaceAll(html, "{{background}}", cfg.UI.Background)

	c.Header("Cache-Control", "no-cache")
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, html)
}

// ServeSetup serves the setup page with template replacements.
func ServeSetup(c *gin.Context, cfg *config.ConfigManager) {
	content, err := loadThemeFile(cfg, "setup.html")
	if err != nil {
		html := fallbackSetupHTML(cfg)
		c.Header("Cache-Control", "no-cache")
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(http.StatusOK, html)
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
	content, err := loadThemeFile(cfg, "admin", "index.html")
	if err != nil {
		html := fallbackAdminHTML(cfg)
		c.Header("Cache-Control", "no-cache")
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(http.StatusOK, html)
		return
	}

	c.Header("Cache-Control", "no-cache")
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, string(content))
}

// ServeUserPage serves user-facing static pages (login/register/dashboard/etc.)
func ServeUserPage(c *gin.Context, cfg *config.ConfigManager, pageName string) {
	content, err := loadThemeFile(cfg, pageName)
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

func fallbackBaseName(cfg *config.ConfigManager) string {
	if cfg != nil && cfg.Base != nil {
		if name := strings.TrimSpace(cfg.Base.Name); name != "" {
			return name
		}
	}
	return "FileCodeBox"
}

func fallbackBaseDescription(cfg *config.ConfigManager) string {
	if cfg != nil && cfg.Base != nil {
		if desc := strings.TrimSpace(cfg.Base.Description); desc != "" {
			return desc
		}
	}
	return "A lightweight file sharing service"
}

func fallbackPageExplain(cfg *config.ConfigManager) string {
	if cfg != nil && cfg.UI != nil {
		if explain := strings.TrimSpace(cfg.UI.PageExplain); explain != "" {
			return explain
		}
	}
	return "Service is running, but the selected theme assets were not found."
}

func fallbackIndexHTML(cfg *config.ConfigManager) string {
	name := html.EscapeString(fallbackBaseName(cfg))
	desc := html.EscapeString(fallbackBaseDescription(cfg))
	explain := html.EscapeString(fallbackPageExplain(cfg))

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>%s</title>
<style>
body{margin:0;padding:0;font-family:-apple-system,BlinkMacSystemFont,"Segoe UI",Roboto,"Helvetica Neue",Arial,sans-serif;background:#f5f5f5;color:#333;}
.wrapper{max-width:720px;margin:15vh auto;padding:32px;background:#fff;border-radius:12px;box-shadow:0 10px 30px rgba(0,0,0,0.08);}
h1{margin-top:0;font-size:2.25rem;}
p{line-height:1.6;}
.muted{color:#666;font-size:0.95rem;}
</style>
</head>
<body>
<div class="wrapper">
<h1>%s</h1>
<p class="muted">%s</p>
<p>%s</p>
<p class="muted">The configured theme directory is missing; static assets will load once it is restored.</p>
</div>
</body>
</html>`, name, name, desc, explain)
}

func fallbackSetupHTML(cfg *config.ConfigManager) string {
	name := html.EscapeString(fallbackBaseName(cfg))
	desc := html.EscapeString(fallbackBaseDescription(cfg))

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>%s - Setup</title>
<style>
body{margin:0;padding:0;font-family:-apple-system,BlinkMacSystemFont,"Segoe UI",Roboto,"Helvetica Neue",Arial,sans-serif;background:#0f172a;color:#e2e8f0;}
.card{max-width:640px;margin:15vh auto;padding:36px;background:rgba(15,23,42,0.92);border-radius:16px;box-shadow:0 12px 40px rgba(15,23,42,0.45);}
h1{margin-top:0;font-size:2rem;color:#38bdf8;}
ul{line-height:1.7;padding-left:1.2rem;}
a{color:#38bdf8;}
</style>
</head>
<body>
<div class="card">
<h1>%s 初始化</h1>
<p>%s</p>
<p>主题资源尚未就绪，请先完成配置文件中的 <code>ui.themes_select</code> 目录部署。</p>
<ul>
<li>确认主题目录已随构建产物一并分发</li>
<li>或在配置中切换到有效的主题路径</li>
<li>之后重新刷新本页面即可完成初始化流程</li>
</ul>
</div>
</body>
</html>`, name, name, desc)
}

func fallbackAdminHTML(cfg *config.ConfigManager) string {
	name := html.EscapeString(fallbackBaseName(cfg))

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>%s Admin</title>
<style>
body{display:flex;align-items:center;justify-content:center;height:100vh;margin:0;background:#1f2937;color:#f9fafb;font-family:-apple-system,BlinkMacSystemFont,"Segoe UI",Roboto,"Helvetica Neue",Arial,sans-serif;text-align:center;padding:24px;}
.panel{max-width:520px;}
h1{font-size:2.2rem;margin-bottom:0.5rem;}
p{line-height:1.6;color:#d1d5db;}
</style>
</head>
<body>
<div class="panel">
<h1>Admin theme missing</h1>
<p>Static assets for the admin console are unavailable. Restore the configured theme directory to load the full interface.</p>
</div>
</body>
</html>`, name)
}
