package handlers

import (
	"meerank/models" // หรือ path ไปยัง model ของคุณ
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// LeaderboardEntry struct สำหรับข้อมูลที่จะส่งกลับไป
// เราสร้าง struct ใหม่เพื่อความชัดเจน และเลือกส่งแค่ข้อมูลที่จำเป็น
type LeaderboardEntry struct {
	Rank       int    `json:"rank"`
	Name       string `json:"name"`
	NumberTree int    `json:"number_tree"`
	Score      int    `json:"score"`
}

// GetLeaderboardHandler ดึงข้อมูลผู้ใช้มาจัดอันดับ
func GetLeaderboardHandler(c *gin.Context, db *gorm.DB) {
    // 1. สร้าง Struct เพื่อรับข้อมูลเฉพาะที่เราต้องการจาก DB
    var leaderboardData []struct {
        Name       string `json:"name"`
        NumberTree int    `json:"number_tree"`
        Minute     int    `json:"minute"`
    }

    // 2. แก้ไข Query ให้เรียงลำดับตาม number_tree และ minute
    // - Order("number_tree DESC, minute DESC"): เรียงจากมากไปน้อย
    // - Limit(10): เอาแค่ 10 อันดับแรก
    err := db.Model(&model.User{}).
        Select("name, number_tree, minute").
        Order("number_tree DESC, minute DESC").
        Limit(10).
        Scan(&leaderboardData).Error

    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch leaderboard data"})
        return
    }

    // 3. ส่งข้อมูลที่ได้กลับไปให้ Frontend
    c.JSON(http.StatusOK, leaderboardData)
}