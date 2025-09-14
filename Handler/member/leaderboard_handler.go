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
	var users []model.User

	// 1. ดึงข้อมูลผู้ใช้ทั้งหมดจากฐานข้อมูล
	// - Order("score DESC"): เรียงลำดับจากคะแนน (score) มากไปน้อย
	// - Limit(100): จำกัดให้ดึงข้อมูลมาแค่ 100 อันดับแรก (เพื่อประสิทธิภาพ)
	// - Select(...): เลือกดึงมาเฉพาะคอลัมน์ที่จำเป็นคือ name, number_tree, score
	result := db.Select("name", "number_tree", "score").
		Order("score DESC").
		Limit(100).
		Find(&users)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch leaderboard data"})
		return
	}

	// 2. จัดรูปแบบข้อมูลเพื่อส่งกลับ (พร้อมเพิ่ม Rank เข้าไป)
	leaderboard := make([]LeaderboardEntry, len(users))
	for i, user := range users {
		leaderboard[i] = LeaderboardEntry{
			Rank:       i + 1, // i เริ่มจาก 0 เลยต้อง +1
			Name:       user.Name,
			NumberTree: user.NumberTree,
			Score:      user.Score,
		}
	}

	// 3. ส่งข้อมูลกลับไปเป็น JSON
	c.JSON(http.StatusOK, leaderboard)
}