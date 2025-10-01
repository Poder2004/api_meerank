package middleware

import (
	// ✨ 1. Import handlers/member เพื่อใช้ Claims และ jwtKey จากที่เดียว ✨
	handlers "meerank/Handler/member"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// ❌ ไม่ต้องประกาศ jwtKey และ Claims struct ที่นี่แล้ว ❌

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			return
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token format"})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// ✨ 2. ใช้ handlers.Claims ที่มี UserID เป็น string ✨
		claims := &handlers.Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			// ✨ 3. ใช้ handlers.jwtKey ตัวเดียวกับตอนสร้าง Token ✨
			return handlers.JwtKey, nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}

		// ✨ 4. ส่งต่อข้อมูลที่ถูกต้องไปให้ Handler ตัวถัดไป ✨
		c.Set("uid", claims.UserID)
		c.Set("role", claims.Role) // <-- เพิ่มการส่ง role ไปด้วย

		c.Next()
	}
}
