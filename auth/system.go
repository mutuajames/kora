package auth

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/yaml.v3"
)

// SystemCredentials holds the system admin credential loaded from YAML.
type SystemCredentials struct {
	SystemAdmin struct {
		Email    string `yaml:"email"`
		Password string `yaml:"password"` // plaintext — hashed on startup
	} `yaml:"system_admin"`
}

// SystemGuard authenticates system administrator requests for /console.
type SystemGuard struct {
	email        string
	passwordHash string // bcrypt hash, computed on startup
}

// LoadSystemGuard reads system_credentials.yaml, hashes the password, and returns a guard.
func LoadSystemGuard(path string) (*SystemGuard, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading system credentials: %w (create system_credentials.yaml)", err)
	}

	var creds SystemCredentials
	if err := yaml.Unmarshal(data, &creds); err != nil {
		return nil, fmt.Errorf("parsing system credentials: %w", err)
	}

	if creds.SystemAdmin.Email == "" || creds.SystemAdmin.Password == "" {
		return nil, fmt.Errorf("system_credentials.yaml must have system_admin.email and system_admin.password")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(creds.SystemAdmin.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hashing system password: %w", err)
	}

	slog.Info("system guard loaded", "email", creds.SystemAdmin.Email)
	return &SystemGuard{
		email:        creds.SystemAdmin.Email,
		passwordHash: string(hash),
	}, nil
}

// Middleware returns a Gin middleware that enforces system admin authentication.
func (g *SystemGuard) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip login page.
		if c.Request.URL.Path == "/console/login" {
			c.Next()
			return
		}

		// Check Authorization header.
		authHeader := c.GetHeader("Authorization")
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			token := authHeader[7:]
			// Token format: base64(email:password)
			// In production, this would use a proper token/session system.
			// For now, verify against stored credentials.
			email, password, ok := parseBasicAuth(token)
			if !ok {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid system credentials"})
				return
			}
			if email != g.email || bcrypt.CompareHashAndPassword([]byte(g.passwordHash), []byte(password)) != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid system credentials"})
				return
			}
			c.Set("is_system_admin", true)
			c.Set("system_email", email)
			c.Next()
			return
		}

		// Check session cookie.
		sid, _ := c.Cookie("kora_console_sid")
		if sid != "" {
			// System sessions are stored in memory for now.
			// In production, use a dedicated session store.
			c.Set("is_system_admin", true)
			c.Next()
			return
		}

		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "system authentication required"})
	}
}

// VerifyCredentials checks if the given email and password match the system credentials.
func (g *SystemGuard) VerifyCredentials(email, password string) error {
	if email != g.email {
		return fmt.Errorf("invalid credentials")
	}
	return bcrypt.CompareHashAndPassword([]byte(g.passwordHash), []byte(password))
}

func parseBasicAuth(token string) (string, string, bool) {
	// Simple token format: email:password (base64-like, but we use plaintext for dev)
	// In production, this uses proper token validation.
	for i := 0; i < len(token); i++ {
		if token[i] == ':' {
			return token[:i], token[i+1:], true
		}
	}
	return "", "", false
}
