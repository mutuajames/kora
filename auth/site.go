package auth

import (
	"database/sql"

	"github.com/gin-gonic/gin"
)

// SiteGuard is a unified middleware that enforces authentication, CSRF,
// and site context resolution for workspace and API routes.
// It wraps AuthMiddleware + CSRFMiddleware + site context injection.
type SiteGuard struct {
	sessionMgr *SessionManager
}

// NewSiteGuard creates a SiteGuard.
func NewSiteGuard(db *sql.DB) *SiteGuard {
	return &SiteGuard{
		sessionMgr: NewSessionManager(db),
	}
}

// Middleware returns the combined site guard middleware.
// It runs: Auth → CSRF → inject site context.
func (g *SiteGuard) Middleware(skipCSRF bool) gin.HandlerFunc {
	auth := AuthMiddleware(g.sessionMgr)
	csrf := CSRFMiddleware()

	return func(c *gin.Context) {
		// Auth check (skips login, ping, and public paths).
		auth(c)
		if c.IsAborted() {
			return
		}

		// Inject site context from SiteRouter into handlers.
		// This ensures every handler can read c.MustGet("site_db") and c.MustGet("site_registry").
		if db, exists := c.Get("site_db"); exists {
			if siteDB, ok := db.(*sql.DB); ok {
				c.Set("db", siteDB)
			}
		}
		if reg, exists := c.Get("site_registry"); exists {
			c.Set("registry", reg)
		}

		// CSRF check (skip for GET/HEAD/OPTIONS, or if explicitly skipped).
		if !skipCSRF {
			csrf(c)
			if c.IsAborted() {
				return
			}
		}

		c.Next()
	}
}

// SiteDB returns the site's database from the request context.
func SiteDB(c *gin.Context) *sql.DB {
	db, _ := c.Get("site_db")
	if db == nil {
		return nil
	}
	return db.(*sql.DB)
}

// SiteRegistry returns the site's DocType registry from the request context.
func SiteRegistry(c *gin.Context) interface{} {
	reg, _ := c.Get("site_registry")
	return reg
}
