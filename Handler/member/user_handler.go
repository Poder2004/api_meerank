package handlers

import (
	"context"
	"log"
	"meerank/models"
	"net/http"

	"cloud.google.com/go/firestore"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GetUserProfileHandler ดึงข้อมูลโปรไฟล์ผู้ใช้จาก Firestore
func GetUserProfileHandler(c *gin.Context, client *firestore.Client) {
	// 1. ดึง uid (Document ID) จาก URL parameter (เป็น string)
	uid := c.Param("uid")

	ctx := context.Background()

	// 2. ค้นหา document ใน collection "users" ด้วย uid
	doc, err := client.Collection("users").Doc(uid).Get(ctx)
	if err != nil {
		// 2.1 ตรวจสอบว่าเป็น error "ไม่พบข้อมูล" หรือไม่
		if status.Code(err) == codes.NotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		// 2.2 ถ้าเป็น error อื่นๆ
		log.Printf("Failed to get user from Firestore: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// 3. แปลงข้อมูลจาก document ไปยัง struct model.User
	var user models.User // <-- บรรทัดนี้ที่เคย error
	if err := doc.DataTo(&user); err != nil {
		log.Printf("Failed to map user data: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process user data"})
		return
	}

	// 4. สร้างข้อมูลที่จะตอบกลับ (Response)
	response := gin.H{
		"name":   user.Name,
		"score":  user.Score,
		"minute": user.Minute,
	}

	// 5. ส่งข้อมูลกลับไปเป็น JSON
	c.JSON(http.StatusOK, response)
}
