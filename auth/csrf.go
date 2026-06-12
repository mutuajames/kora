package auth

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"github.com/gin-gonic/gin"
)

// CSRFSecure controls the Secure flag on CSRF cookies. Set true in production with TLS.
var CSRFSecure = false

// SetCSRFSecure sets the Secure flag for CSRF cookies.
func SetCSRFSecure(secure bool) { CSRFSecure = secure }

// CSRFMiddleware protects state-changing requests with a double-submit cookie pattern.
// On first GET, a CSRF token is set as a cookie (HttpOnly=false so JS can read it).
// On POST/PUT/DELETE, the X-Kora-CSRF-Token header must match the cookie value.
func CSRFMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip safe methods.
		if c.Request.Method == "GET" || c.Request.Method == "HEAD" || c.Request.Method == "OPTIONS" {
			ensureCSRFCookie(c)
			c.Next()
			return
		}

		// For state-changing methods, verify the token.
		cookieToken, err := c.Cookie("kora_csrf")
		if err != nil || cookieToken == "" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":   "csrf_token_required",
				"message": "CSRF token is required for state-changing requests.",
			})
			return
		}

		headerToken := c.GetHeader("X-Kora-CSRF-Token")
		if headerToken == "" {
			// Also check X-CSRF-Token for compatibility.
			headerToken = c.GetHeader("X-CSRF-Token")
		}

		if headerToken != cookieToken {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":   "csrf_token_mismatch",
				"message": "CSRF token mismatch.",
			})
			return
		}

		c.Next()
	}
}

// ensureCSRFCookie sets a CSRF cookie if one doesn't exist.
func ensureCSRFCookie(c *gin.Context) {
	_, err := c.Cookie("kora_csrf")
	if err == nil {
		return // Already exists.
	}

	token := generateCSRFToken()
	c.SetCookie("kora_csrf", token, 86400, "/", "", CSRFSecure, false)
	c.Set("csrf_token", token)
}

func generateCSRFToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}
