package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// SetupGuardConfig controls how the setup guard middleware behaves.
type SetupGuardConfig struct {
	// IsInitialized returns the current initialization status.
	IsInitialized func() (bool, error)
	// SetupPath denotes the setup entry path, defaults to /setup.
	SetupPath string
	// RedirectPath denotes the path to redirect to once initialized, defaults to /.
	RedirectPath string
	// AllowPaths lists exact paths that should remain accessible before initialization.
	AllowPaths []string
	// AllowPrefixes lists path prefixes that should remain accessible before initialization.
	AllowPrefixes []string
}

// SetupGuard ensures only setup resources are accessible before initialization
// and blocks setup routes after initialization is complete.
func SetupGuard(cfg SetupGuardConfig) gin.HandlerFunc {
	setupPath := cfg.SetupPath
	if setupPath == "" {
		setupPath = "/setup"
	}
	redirectPath := cfg.RedirectPath
	if redirectPath == "" {
		redirectPath = "/"
	}

	allowPaths := map[string]struct{}{
		setupPath:       {},
		setupPath + "/": {},
	}

	for _, p := range cfg.AllowPaths {
		allowPaths[p] = struct{}{}
	}

	allowPrefixes := []string{setupPath + "/"}
	allowPrefixes = append(allowPrefixes, cfg.AllowPrefixes...)

	return func(c *gin.Context) {
		initialized := false
		if cfg.IsInitialized != nil {
			var err error
			initialized, err = cfg.IsInitialized()
			if err != nil {
				logrus.WithError(err).Warn("setup guard: failed to determine initialization state")
				// Fail closed on error so users can still reach setup for recovery.
				initialized = false
			}
		}

		path := c.Request.URL.Path

		if initialized {
			if path == setupPath || strings.HasPrefix(path, setupPath+"/") {
				switch c.Request.Method {
				case http.MethodGet, http.MethodHead:
					c.Redirect(http.StatusFound, redirectPath)
				case http.MethodOptions:
					c.Status(http.StatusNoContent)
				default:
					c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
						"code":    http.StatusForbidden,
						"message": "系统已初始化，禁止重新初始化",
					})
				}
				c.Abort()
				return
			}

			c.Next()
			return
		}

		if _, ok := allowPaths[path]; ok {
			c.Next()
			return
		}
		for _, prefix := range allowPrefixes {
			if strings.HasPrefix(path, prefix) {
				c.Next()
				return
			}
		}

		switch c.Request.Method {
		case http.MethodGet, http.MethodHead:
			c.Redirect(http.StatusFound, setupPath)
		case http.MethodOptions:
			// Allow CORS preflight to complete without redirect loops.
			c.Status(http.StatusNoContent)
		default:
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"code":    http.StatusForbidden,
				"message": "系统未初始化，请访问 /setup 完成初始化",
			})
		}
		c.Abort()
	}
}
