package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RoleMiddleware สร้าง Middleware ที่ต้องการ Role ที่กำหนด
func RoleMiddleware(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// ดึง role ที่ถูกตั้งค่าไว้โดย AuthMiddleware
		role, exists := c.Get("role")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Role not found in token"})
			return
		}

		// แปลงค่า role เป็น string แล้วเปรียบเทียบ
		userRole, ok := role.(string)
		if !ok || userRole != requiredRole {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
			return
		}

		c.Next()
	}
}