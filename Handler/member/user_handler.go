package handlers

import (
	"meerank/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetUserProfileHandler ดึงข้อมูลโปรไฟล์ผู้ใช้ (ชื่อ, คะแนน, นาที)
func GetUserProfileHandler(c *gin.Context, db *gorm.DB) {
	// 1. ดึง uid จาก URL parameter
	uidStr := c.Param("uid")
	uid, err := strconv.ParseInt(uidStr, 10, 64) // แปลง string เป็น int64
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// 2. ค้นหาผู้ใช้ในฐานข้อมูลด้วย uid
	var user model.User
	if err := db.First(&user, uid).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// 3. สร้างข้อมูลที่จะตอบกลับ (Response)
	response := gin.H{
		"name":   user.Name,
		"score":  user.Score,
		"minute": user.Minute,
	}

	// 4. ส่งข้อมูลกลับไปเป็น JSON
	c.JSON(http.StatusOK, response)
}