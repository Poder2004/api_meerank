package handlers

import (
	"context"
	"fmt"
	"log"
	"meerank/models"
	"net/http"

	"cloud.google.com/go/firestore"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// --- 1. Get All Users (UID and Name only) ---

// UserSummary เป็น struct สำหรับส่งข้อมูลแบบย่อ
type UserSummary struct {
	UID  string `json:"uid"`
	Name string `json:"name"`
}

// GetAllUsersSummaryHandler ดึงรายชื่อผู้ใช้ทั้งหมด (เฉพาะ UID และ Name)
// GetAllUsersSummaryHandler ดึงรายชื่อผู้ใช้ทั้งหมด (ยกเว้น admin ที่เรียก)
func GetAllUsersSummaryHandler(c *gin.Context, client *firestore.Client) {
	// ✨ 1. ดึง UID ของ Admin ที่ล็อกอินอยู่ออกจาก Context ✨
	adminUIDValue, exists := c.Get("uid")
	if !exists {
		// กรณีนี้ไม่น่าเกิดขึ้นถ้าผ่าน Middleware มาได้ แต่ใส่ไว้เพื่อความปลอดภัย
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Admin UID not found in token"})
		return
	}
	adminUID, _ := adminUIDValue.(string)

	var users []UserSummary
	ctx := context.Background()

	// 2. ดึง document ทั้งหมดจาก collection users
	iter := client.Collection(models.CollectionUsers).Documents(ctx)
	defer iter.Stop()

	// 3. วนลูปเพื่ออ่านข้อมูล
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Printf("Failed to iterate users: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not retrieve users"})
			return
		}

		// ✨ 4. เพิ่มเงื่อนไข: ถ้า UID ไม่ใช่ของ Admin คนปัจจุบัน ให้เพิ่มเข้าไปใน list ✨
		if doc.Ref.ID != adminUID {
			var user models.User
			if err := doc.DataTo(&user); err == nil {
				users = append(users, UserSummary{
					UID:  doc.Ref.ID,
					Name: user.Name,
				})
			}
		}
	}

	// 5. ส่งข้อมูลกลับไป
	c.JSON(http.StatusOK, users)
}

// --- 2. Get Full User Profile by UID ---

// GetFullUserProfileHandler ดึงข้อมูลทั้งหมดของผู้ใช้จาก UID ที่ระบุ
func GetFullUserProfileHandler(c *gin.Context, client *firestore.Client) {
	// 1. ดึง uid จาก URL parameter
	uid := c.Param("uid")
	if uid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	ctx := context.Background()

	// 2. ค้นหา document ด้วย uid
	doc, err := client.Collection(models.CollectionUsers).Doc(uid).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		log.Printf("Failed to get user by ID: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// 3. แปลงข้อมูลและใส่ ID กลับเข้าไปใน struct
	var user models.User
	if err := doc.DataTo(&user); err != nil {
		log.Printf("Failed to process user data: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process user data"})
		return
	}
	user.ID = doc.Ref.ID // อย่าลืมใส่ Document ID กลับเข้าไป

	// 4. ส่งข้อมูลทั้งหมดกลับไป
	c.JSON(http.StatusOK, user)
}

// ResetAllUsersStatsHandler รีเซ็ตค่า minute, score, number_tree, tree_progress ของผู้ใช้ทุกคนให้เป็น 0
func ResetAllUsersStatsHandler(c *gin.Context, client *firestore.Client) {
	ctx := context.Background()
	userCount := 0

	// ✨ ใช้ Batched Writes เพื่อประสิทธิภาพสูงสุดในการอัปเดตข้อมูลจำนวนมาก
	batch := client.Batch()

	// 1. ดึงข้อมูลผู้ใช้ทั้งหมดใน Collection "users"
	iter := client.Collection(models.CollectionUsers).Documents(ctx)
	defer iter.Stop()

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Printf("Failed to iterate users for reset: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user data"})
			return
		}

		// 2. เตรียมข้อมูลที่จะอัปเดต
		updates := []firestore.Update{
			{Path: "minute", Value: 0},
			{Path: "score", Value: 0},
			{Path: "number_tree", Value: 0},
			{Path: "tree_progress", Value: 0},
		}

		// 3. เพิ่ม Operation การอัปเดตเข้าไปใน Batch
		batch.Update(doc.Ref, updates)
		userCount++

		// Firestore Batched Writes มีขีดจำกัดที่ 500 operations ต่อครั้ง
		// หากมีผู้ใช้จำนวนมาก เราจะ commit ทุกๆ 500 คน แล้วเริ่ม batch ใหม่
		if userCount%500 == 0 {
			if _, err := batch.Commit(ctx); err != nil {
				log.Printf("Batch commit failed during reset: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user stats"})
				return
			}
			// เริ่ม Batch ใหม่สำหรับรอบต่อไป
			batch = client.Batch()
		}
	}

	// 4. Commit Batch สุดท้าย (สำหรับผู้ใช้ที่เหลือที่ยังไม่ถึง 500)
	if _, err := batch.Commit(ctx); err != nil {
		log.Printf("Final batch commit failed during reset: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to finalize updating user stats"})
		return
	}

	// 5. ส่งคำตอบกลับเมื่อสำเร็จ
	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Successfully reset stats for %d users.", userCount),
	})
}
