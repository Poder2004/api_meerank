package handlers

import (
	"context"
	"log"
	"net/http"
	"time"

	"meerank/models"

	"cloud.google.com/go/firestore"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/api/iterator"
)

// --- JWT Configuration ---
// !!คำเตือน: ในแอปจริง ห้ามเก็บ Secret Key ไว้ในโค้ดเด็ดขาด!
// ควรเก็บไว้ใน Environment Variable
var JwtKey = []byte("your_very_secret_key_that_is_long_and_secure")

type Claims struct {
	UserID string `json:"user_id"` // <-- ID ของ Firestore เป็น string
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// --- Register Handler (สำหรับ Firebase) ---

// RegisterHandler จัดการการสมัครสมาชิกใหม่ด้วย Firestore
func RegisterHandler(c *gin.Context, client *firestore.Client) {
	// 1. รับข้อมูลจาก JSON payload (ไม่มี password)
	var payload struct {
		Name   string  `json:"name" binding:"required"`
		Phone  string  `json:"phone" binding:"required"`
		Age    *int    `json:"age"`
		Gender *string `json:"gender"`
	}

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
		return
	}

	ctx := context.Background()

	// 2. ตรวจสอบว่าเบอร์โทรศัพท์นี้เคยลงทะเบียนแล้วหรือยัง
	// สร้าง query เพื่อค้นหา document ที่มี field 'phone' ตรงกับ payload
	iter := client.Collection(models.CollectionUsers).Where("phone", "==", payload.Phone).Limit(1).Documents(ctx)

	defer iter.Stop()

	// ตรวจสอบว่า query เจอผลลัพธ์หรือไม่
	_, err := iter.Next()
	if err != iterator.Done { // ถ้าไม่เจอ error 'Done' แปลว่ามีข้อมูลอยู่แล้ว
		c.JSON(http.StatusConflict, gin.H{"error": "Phone number already registered"})
		return
	}

	// 3. สร้างข้อมูลผู้ใช้ใหม่เพื่อบันทึกลง Firestore
	newUser := models.User{
		Name:   payload.Name,
		Phone:  &payload.Phone,
		Role:   models.RoleMember, // กำหนด role เริ่มต้น
		Age:    payload.Age,
		Gender: payload.Gender,
		// ไม่มี Password อีกต่อไป
	}

	// 4. บันทึกผู้ใช้ใหม่ลงใน collection "users"
	// Firestore จะสร้าง ID ให้โดยอัตโนมัติ
	_, _, err = client.Collection(models.CollectionUsers).Add(ctx, newUser)
	if err != nil {
		log.Printf("Failed to create user in Firestore: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// 5. ส่งคำตอบกลับไป
	c.JSON(http.StatusCreated, gin.H{"message": "Registration successful"})
}

// --- Login Handler (สำหรับ Firebase) ---

// LoginHandler จัดการการล็อกอิน, คำนวณวัน, และอัปเดตเวลาล่าสุด
func LoginHandler(c *gin.Context, client *firestore.Client) {
	// 1. รับข้อมูลจาก JSON payload (มีแค่เบอร์โทร)
	var payload struct {
		Phone string `json:"phone" binding:"required"`
	}

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
		return
	}

	ctx := context.Background()

	// 2. ค้นหาผู้ใช้จากเบอร์โทรศัพท์ใน Firestore
	var user models.User
	var docID string

	iter := client.Collection("users").Where("phone", "==", payload.Phone).Limit(1).Documents(ctx)
	defer iter.Stop()

	doc, err := iter.Next()
	if err == iterator.Done {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Phone number not found"})
		return
	}
	if err != nil {
		log.Printf("Error querying user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	if err := doc.DataTo(&user); err != nil {
		log.Printf("Error converting user data: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	docID = doc.Ref.ID

	// --- V V V ส่วนที่เพิ่มเข้ามา V V V ---

	// 3. คำนวณจำนวนวันที่ไม่ได้ล็อกอิน
	var daysSinceLastLogin int
	now := time.Now()

	if user.LastLoginAt == nil {
		// ถ้าไม่เคยมีประวัติการล็อกอินเลย (ครั้งแรก) ให้ถือว่าเป็น 0 วัน
		daysSinceLastLogin = 0
	} else {
		// คำนวณโดยใช้เฉพาะวันที่ (ไม่สนใจเวลา) เพื่อความแม่นยำ
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		lastLoginDay := time.Date(user.LastLoginAt.Year(), user.LastLoginAt.Month(), user.LastLoginAt.Day(), 0, 0, 0, 0, user.LastLoginAt.Location())

		duration := today.Sub(lastLoginDay)
		daysSinceLastLogin = int(duration.Hours() / 24)
	}

	// 4. อัปเดตเวลาล็อกอินล่าสุดใน Firestore ให้เป็นเวลาปัจจุบัน
	// ทำขั้นตอนนี้หลังจากคำนวณเสร็จแล้ว
	updateData := []firestore.Update{
		{Path: "last_login_at", Value: now},
	}
	if _, err := doc.Ref.Update(ctx, updateData); err != nil {
		// บันทึก error แต่ไม่ต้องหยุดการทำงาน เพื่อให้ผู้ใช้ยังล็อกอินได้
		log.Printf("Failed to update last login time for user %s: %v", docID, err)
	}

	// --- ^ ^ ^ สิ้นสุดส่วนที่เพิ่มเข้ามา ^ ^ ^ ---

	// 5. สร้าง JWT Token
	expirationTime := time.Now().Add(12 * time.Hour)
	claims := &Claims{
		UserID: docID,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(JwtKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate token"})
		return
	}

	// 6. ส่งคำตอบกลับพร้อม Token และจำนวนวันที่ไม่ได้ล็อกอิน
	c.JSON(http.StatusOK, gin.H{
		"message":               "Login successful",
		"name":                  user.Name,
		"token":                 tokenString,
		"role":                  user.Role,
		"days_since_last_login": daysSinceLastLogin, // <-- เพิ่มค่านี้เข้าไปใน response
	})
}
