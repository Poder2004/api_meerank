package routers

import (
	// handlersadmin "meerank/Handler/"
	handlersadmin "meerank/Handler/admin"
	handlers "meerank/Handler/member"
	"meerank/middleware"
	"meerank/models"

	"cloud.google.com/go/firestore" // 1. เปลี่ยน import จาก gorm
	"github.com/gin-gonic/gin"
)

// SetupRouter ฟังก์ชันสำหรับตั้งค่า routes ของแอป
// 2. เปลี่ยนพารามิเตอร์จาก db *gorm.DB เป็น client *firestore.Client
func SetupRouter(r *gin.Engine, client *firestore.Client) {

	// 3. เปลี่ยนการส่ง db เป็น client ในทุกๆ handler
	r.POST("/register", func(c *gin.Context) {
		handlers.RegisterHandler(c, client)
	})

	r.POST("/login", func(c *gin.Context) {
		handlers.LoginHandler(c, client)
	})

	r.GET("/user/:uid", func(c *gin.Context) {
		handlers.GetUserProfileHandler(c, client)
	})

	r.GET("/leaderboard", func(c *gin.Context) { handlers.GetLeaderboardHandler(c, client) })

	// --- Protected Routes (ต้องล็อกอิน) ---
	profileGroup := r.Group("/profile")
	profileGroup.Use(middleware.AuthMiddleware())
	{
		profileGroup.GET("/me", func(c *gin.Context) { handlers.GetMyProfileHandler(c, client) })
		profileGroup.PUT("/me", func(c *gin.Context) { handlers.UpdateMyProfileHandler(c, client) })
		profileGroup.POST("/activity", func(c *gin.Context) { handlers.UpdateUserActivityHandler(c, client) })
		profileGroup.POST("/tree", func(c *gin.Context) { handlers.AddTreeHandler(c, client) })
		profileGroup.POST("/tree/water", func(c *gin.Context) { handlers.WaterTreeHandler(c, client) })
	}

	// --- Admin Routes (สำหรับ Admin เท่านั้น) ---
	adminGroup := r.Group("/admin")
	adminGroup.Use(middleware.AuthMiddleware())
	adminGroup.Use(middleware.RoleMiddleware(models.RoleAdmin))
	{
		// GET /admin/users -> ดึงรายชื่อผู้ใช้ทั้งหมด (แบบย่อ)
		adminGroup.GET("/users", func(c *gin.Context) {
			handlersadmin.GetAllUsersSummaryHandler(c, client)
		})

		// GET /admin/users/:uid -> ดึงข้อมูลผู้ใช้ 1 คน (แบบเต็ม)
		adminGroup.GET("/users/:uid", func(c *gin.Context) {
			handlersadmin.GetFullUserProfileHandler(c, client)
		})

		adminGroup.POST("/users/reset-stats", func(c *gin.Context) {
			handlersadmin.ResetAllUsersStatsHandler(c, client)
		})
	}
}
