package handlers

import (
	model "meerank/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// GetMyProfileHandler ดึงข้อมูลโปรไฟล์ของ user ที่ล็อกอินอยู่
func GetMyProfileHandler(c *gin.Context, db *gorm.DB) {
	// ดึง uid ที่ได้จาก Middleware
	uid, exists := c.Get("uid")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
		return
	}

	var user model.User
	if err := db.First(&user, uid).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"name":   user.Name,
		"phone":  user.Phone,
		"age":    user.Age,
		"gender": user.Gender,
		"score":  user.Score,  // <-- เพิ่ม
		"minute": user.Minute, // <-- เพิ่ม
	})
}

// UpdateMyProfileHandler อัปเดตข้อมูลโปรไฟล์ของ user ที่ล็อกอินอยู่
func UpdateMyProfileHandler(c *gin.Context, db *gorm.DB) {
	// ดึง uid ที่ได้จาก Middleware
	uid, exists := c.Get("uid")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
		return
	}

	var payload struct {
		Name     string  `json:"name"`
		Phone    string  `json:"phone"`
		Password string  `json:"password"` // รับรหัสผ่านใหม่ (ถ้ามี)
		Age      *int    `json:"age"`
		Gender   *string `json:"gender"`
	}

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data"})
		return
	}

	var user model.User
	if err := db.First(&user, uid).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// อัปเดตข้อมูล
	user.Name = payload.Name
	user.Phone = &payload.Phone
	user.Age = payload.Age
	user.Gender = payload.Gender

	// ถ้ามีการส่งรหัสผ่านใหม่มาด้วย ให้ทำการ hash แล้วอัปเดต
	if payload.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
			return
		}
		user.Password = string(hashedPassword)
	}

	// บันทึกการเปลี่ยนแปลงลงฐานข้อมูล
	if err := db.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Profile updated successfully"})
}

// UpdateUserActivityHandler อัปเดตคะแนนและนาทีจากการออกกำลังกาย
func UpdateUserActivityHandler(c *gin.Context, db *gorm.DB) {
	// 1. ดึง uid ที่ได้จาก Middleware
	uid, exists := c.Get("uid")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
		return
	}

	// 2. รับข้อมูลนาทีและคะแนนที่จะบวกเพิ่มจาก JSON
	var payload struct {
		Minute int `json:"minute"`
		Score  int `json:"score"`
	}

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data, 'minute' and 'score' are required"})
		return
	}

	// 3. อัปเดตข้อมูลในฐานข้อมูล
	// เราใช้ gorm.Expr เพื่อบอกให้ GORM ทำการบวกค่าในระดับ Database
	// ซึ่งปลอดภัยและมีประสิทธิภาพกว่าการดึงค่าเก่ามาบวกในโค้ด Go
	err := db.Model(&model.User{}).Where("uid = ?", uid).Updates(map[string]interface{}{
		"minute": gorm.Expr("minute + ?", payload.Minute),
		"score":  gorm.Expr("score + ?", payload.Score),
	}).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user activity"})
		return
	}

	// 4. ตอบกลับว่าสำเร็จ
	c.JSON(http.StatusOK, gin.H{"message": "Score and minute updated successfully"})
}

// AddTreeHandler เพิ่มจำนวนต้นไม้สะสมของผู้ใช้
func AddTreeHandler(c *gin.Context, db *gorm.DB) {
	// 1. ดึง uid ที่ได้จาก Middleware
	uid, exists := c.Get("uid")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
		return
	}

	// 2. อัปเดตจำนวนต้นไม้ในฐานข้อมูล
	// ใช้ gorm.Expr เพื่อบวกค่า number_tree เพิ่มไป 1
	err := db.Model(&model.User{}).Where("uid = ?", uid).Update("number_tree", gorm.Expr("number_tree + ?", 1)).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update tree count"})
		return
	}

	// 3. ตอบกลับว่าสำเร็จ
	c.JSON(http.StatusOK, gin.H{"message": "Tree count updated successfully"})
}
