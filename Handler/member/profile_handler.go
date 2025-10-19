package handlers

import (
	"context"
	"errors"
	"log"
	"meerank/models"
	"net/http"

	"cloud.google.com/go/firestore"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GetMyProfileHandler ดึงข้อมูลโปรไฟล์ของ user ที่ล็อกอินอยู่
func GetMyProfileHandler(c *gin.Context, client *firestore.Client) {
	// 1. ดึง uid (string) ที่ได้จาก Middleware
	uidValue, exists := c.Get("uid")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
		return
	}

	uid, ok := uidValue.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid User ID format in token"})
		return
	}

	ctx := context.Background()

	// 2. ดึง document จาก Firestore
	doc, err := client.Collection(models.CollectionUsers).Doc(uid).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// 3. แปลงข้อมูลและส่งกลับ
	var user models.User
	if err := doc.DataTo(&user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process user data"})
		return
	}
	user.ID = doc.Ref.ID // เพิ่ม ID เข้าไปใน struct ก่อนส่งกลับ

	c.JSON(http.StatusOK, user)
}

// UpdateMyProfileHandler อัปเดตข้อมูลโปรไฟล์ของ user ที่ล็อกอินอยู่
func UpdateMyProfileHandler(c *gin.Context, client *firestore.Client) {
	// 1. ดึง uid (string) จาก Middleware
	uidValue, exists := c.Get("uid")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
		return
	}
	uid, ok := uidValue.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid User ID format in token"})
		return
	}

	var payload struct {
		Name   *string `json:"name"`
		Phone  *string `json:"phone"`
		Age    *int    `json:"age"`
		Gender *string `json:"gender"`
	}

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data"})
		return
	}

	ctx := context.Background()

	// 2. สร้างรายการอัปเดตเฉพาะ field ที่ส่งมา
	var updates []firestore.Update
	if payload.Name != nil {
		updates = append(updates, firestore.Update{Path: "name", Value: *payload.Name})
	}
	if payload.Phone != nil {
		updates = append(updates, firestore.Update{Path: "phone", Value: *payload.Phone})
	}
	if payload.Age != nil {
		updates = append(updates, firestore.Update{Path: "age", Value: *payload.Age})
	}
	if payload.Gender != nil {
		updates = append(updates, firestore.Update{Path: "gender", Value: *payload.Gender})
	}

	// 3. ถ้าไม่มีข้อมูลให้อัปเดต ก็ไม่ต้องทำอะไร
	if len(updates) == 0 {
		c.JSON(http.StatusOK, gin.H{"message": "No fields to update"})
		return
	}

	// 4. บันทึกการเปลี่ยนแปลงลง Firestore
	_, err := client.Collection(models.CollectionUsers).Doc(uid).Update(ctx, updates)
	if err != nil {
		log.Printf("Failed to update profile: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Profile updated successfully"})
}

// UpdateUserActivityHandler อัปเดตคะแนนและนาทีจากการออกกำลังกาย
func UpdateUserActivityHandler(c *gin.Context, client *firestore.Client) {
	// 1. ดึง uid (string) จาก Middleware
	uidValue, exists := c.Get("uid")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
		return
	}
	uid, ok := uidValue.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid User ID format in token"})
		return
	}

	var payload struct {
		Minute int `json:"minute"`
		Score  int `json:"score"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data, 'minute' and 'score' are required"})
		return
	}

	ctx := context.Background()

	// 2. ใช้ firestore.Increment เพื่อบวกค่าแบบ Atomic
	updates := []firestore.Update{
		{Path: "minute", Value: firestore.Increment(payload.Minute)},
		{Path: "score", Value: firestore.Increment(payload.Score)},
	}

	// 3. อัปเดตข้อมูลใน Firestore
	_, err := client.Collection(models.CollectionUsers).Doc(uid).Update(ctx, updates)
	if err != nil {
		log.Printf("Failed to update user activity: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user activity"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Score and minute updated successfully"})
}

// AddTreeHandler เพิ่มจำนวนต้นไม้สะสมของผู้ใช้
func AddTreeHandler(c *gin.Context, client *firestore.Client) {
	// 1. ดึง uid (string) จาก Middleware
	uidValue, exists := c.Get("uid")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
		return
	}
	uid, ok := uidValue.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid User ID format in token"})
		return
	}

	ctx := context.Background()

	// 2. ใช้ firestore.Increment เพื่อบวกค่า number_tree เพิ่มไป 1
	update := firestore.Update{Path: "number_tree", Value: firestore.Increment(1)}

	_, err := client.Collection(models.CollectionUsers).Doc(uid).Update(ctx, []firestore.Update{update})
	if err != nil {
		log.Printf("Failed to update tree count: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update tree count"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tree count updated successfully"})
}

// WaterTreeHandler จัดการการรดน้ำต้นไม้ (ใช้ Transaction)
func WaterTreeHandler(c *gin.Context, client *firestore.Client) {
	// 1. ดึง uid (string) จาก Middleware
	uidValue, exists := c.Get("uid")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
		return
	}
	uid, ok := uidValue.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid User ID format in token"})
		return
	}

	var payload struct {
		Amount int `json:"amount"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	if payload.Amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Amount must be positive"})
		return
	}

	ctx := context.Background()
	userRef := client.Collection(models.CollectionUsers).Doc(uid)

	var finalUser models.User

	// 2. ใช้ Transaction เพื่อความปลอดภัยของข้อมูล
	err := client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		doc, err := tx.Get(userRef)
		if err != nil {
			return err
		}

		var user models.User
		if err := doc.DataTo(&user); err != nil {
			return err
		}

		// ตรรกะทางธุรกิจ
		if user.Score < payload.Amount {
			return errors.New("not enough score")
		}

		user.Score -= payload.Amount
		user.TreeProgress += payload.Amount

		if user.TreeProgress >= 1000 {
			user.NumberTree += 1
			user.TreeProgress -= 1000
		}

		// เก็บข้อมูลล่าสุดเพื่อส่งกลับ
		finalUser = user

		// ทำการอัปเดตใน transaction
		return tx.Update(userRef, []firestore.Update{
			{Path: "score", Value: user.Score},
			{Path: "tree_progress", Value: user.TreeProgress},
			{Path: "number_tree", Value: user.NumberTree},
		})
	})

	// 3. จัดการผลลัพธ์ของ Transaction
	if err != nil {
		if err.Error() == "not enough score" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Not enough score"})
			return
		}
		log.Printf("WaterTree transaction failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "Tree watered successfully",
		"new_score":     finalUser.Score,
		"tree_progress": finalUser.TreeProgress,
		"number_tree":   finalUser.NumberTree,
	})
}
