package routers

import (
	// handlersadmin "meerank/Handler/"
	handlers "meerank/Handler/member"
	"meerank/middleware"
	model "meerank/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SetupRouter ฟังก์ชันสำหรับตั้งค่า routes ของแอป
func SetupRouter(r *gin.Engine, db *gorm.DB) {

	r.POST("/register", func(c *gin.Context) {
		handlers.RegisterHandler(c, db)
	})

	r.POST("/login", func(c *gin.Context) {
		handlers.LoginHandler(c, db)
	})

	r.GET("/user/:uid", func(c *gin.Context) {
		handlers.GetUserProfileHandler(c, db)
	})

	r.GET("/leaderboard", func(c *gin.Context) { handlers.GetLeaderboardHandler(c, db) })

	// --- Protected Routes (ต้องล็อกอิน) ---
	profileGroup := r.Group("/profile")
	profileGroup.Use(middleware.AuthMiddleware()) // <-- ใช้ Middleware กับทุกเส้นทางในกลุ่มนี้
	{
		// GET /profile/me
		profileGroup.GET("/me", func(c *gin.Context) { handlers.GetMyProfileHandler(c, db) })

		// PUT /profile/me
		profileGroup.PUT("/me", func(c *gin.Context) { handlers.UpdateMyProfileHandler(c, db) })

		// --- เพิ่ม Route ใหม่สำหรับอัปเดตกิจกรรมตรงนี้ ---
		// POST /profile/activity
		profileGroup.POST("/activity", func(c *gin.Context) { handlers.UpdateUserActivityHandler(c, db) })
		// POST /profile/tree
		profileGroup.POST("/tree", func(c *gin.Context) { handlers.AddTreeHandler(c, db) })

	}

	// --- Admin Routes (สำหรับ Admin เท่านั้น) ---
	adminGroup := r.Group("/admin")
	// ✨ ใช้ Middleware 2 ชั้น: 1. ต้องล็อกอิน, 2. ต้องเป็น Admin
	adminGroup.Use(middleware.AuthMiddleware())
	adminGroup.Use(middleware.RoleMiddleware(model.RoleAdmin))
	{
		// ตัวอย่างเส้นทางสำหรับ Admin
		adminGroup.GET("/dashboard", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "Welcome to Admin Dashboard!"})
		})
	}
}
