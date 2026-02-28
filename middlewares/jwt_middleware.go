// All rights reserved Samyak-Setu

package middlewares

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/samyaksetu/backend/services"
)

// JWTAuth is a middleware that verifies the JWT token in the Authorization header.
// If valid, it sets "farmerId", "farmerPhone", and "farmerName" in the Gin context.
// Protected routes can then use c.GetString("farmerId") to identify the logged-in user.
func JWTAuth(jwtService *services.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		// Expect format: "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header must be: Bearer <token>"})
			c.Abort()
			return
		}

		claims, err := jwtService.ValidateToken(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Inject farmer info into request context for downstream handlers
		c.Set("farmerId", claims.FarmerID)
		c.Set("farmerPhone", claims.Phone)
		c.Set("farmerName", claims.Name)

		c.Next()
	}
}
