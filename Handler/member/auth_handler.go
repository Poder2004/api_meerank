package handlers

import (
	model "meerank/models" // <-- import model ที่สร้างไว้
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// --- JWT Configuration ---
// !!คำเตือน: ในแอปจริง ห้ามเก็บ Secret Key ไว้ในโค้ดเด็ดขาด!
// ควรเก็บไว้ใน Environment Variable
var jwtKey = []byte("your_very_secret_key_that_is_long_and_secure")

type Claims struct {
	UserID int64  `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// --- Register Handler ---

// RegisterHandler จัดการการสมัครสมาชิกใหม่
func RegisterHandler(c *gin.Context, db *gorm.DB) {
	// --- STEP 1: เพิ่ม age และ gender เข้าไปใน payload ---
	var payload struct {
		Name     string  `json:"name" binding:"required"`
		Phone    string  `json:"phone" binding:"required"`
		Password string  `json:"password" binding:"required"`
		Age      *int    `json:"age"`    // <-- เพิ่มเข้ามา (ใช้ pointer เพราะอาจไม่ส่งมาก็ได้)
		Gender   *string `json:"gender"` // <-- เพิ่มเข้ามา (ใช้ pointer เพราะอาจไม่ส่งมาก็ได้)
	}

	// 1. Bind and Validate JSON payload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
		return
	}

	// 2. Check if phone number already exists
	var existingUser model.User
	if err := db.Where("phone = ?", payload.Phone).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Phone number already registered"})
		return
	}

	// 3. Hash the password for security
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// --- STEP 2: เพิ่ม age และ gender ตอนสร้าง User object ---
	newUser := model.User{
		Name:     payload.Name,
		Phone:    &payload.Phone,
		Password: string(hashedPassword),
		Role:     model.RoleMember,
		Age:      payload.Age,    // <-- เพิ่มเข้ามา
		Gender:   payload.Gender, // <-- เพิ่มเข้ามา
	}

	// 5. Save the new user to the database
	if err := db.Create(&newUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// 6. Return success response
	c.JSON(http.StatusCreated, gin.H{"message": "Registration successful"})
}

// --- Login Handler ---

// LoginHandler จัดการการล็อกอิน
func LoginHandler(c *gin.Context, db *gorm.DB) {
	var payload struct {
		Phone    string `json:"phone" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	// 1. Bind and Validate JSON payload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
		return
	}

	// 2. Find the user by phone number
	var user model.User
	if err := db.Where("phone = ?", payload.Phone).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid phone number or password"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// 3. Compare the hashed password with the provided password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(payload.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid phone number or password"})
		return
	}

	// 4. Generate JWT Token
	expirationTime := time.Now().Add(24 * time.Hour) // Token หมดอายุใน 24 ชั่วโมง
	claims := &Claims{
		UserID: user.UID,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate token"})
		return
	}

	// 5. Return success response with the token
	// 5. Return success response with name and token
	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"name":    user.Name,
		"token":   tokenString,
		"role":    user.Role, // ✨ เพิ่มบรรทัดนี้เข้ามา
	})
}
