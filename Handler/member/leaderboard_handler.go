package handlers

import (
	"context"
	"log"
	"meerank/models"
	"net/http"

	"cloud.google.com/go/firestore"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/iterator"
)

// LeaderboardEntry struct สำหรับข้อมูลที่จะส่งกลับไป
// ✨ แก้ไข: เอา Rank ออก เพราะ Frontend จะเป็นคนจัดการลำดับเอง
type LeaderboardEntry struct {
	Name       string `json:"name"`
	NumberTree int    `json:"number_tree"`
	Score      int    `json:"score"`
}

// GetLeaderboardHandler ดึงข้อมูลผู้ใช้มาจัดอันดับจาก Firestore
func GetLeaderboardHandler(c *gin.Context, client *firestore.Client) {
	ctx := context.Background()
	var leaderboard []LeaderboardEntry

	// 1. สร้าง Query เพื่อดึงข้อมูลผู้ใช้ 10 อันดับแรก
	// ✨ เพิ่ม .Where(...) เพื่อกรองเอาเฉพาะ member เท่านั้น
	iter := client.Collection(models.CollectionUsers).
		Where("role", "==", models.RoleMember). // <-- เพิ่มบรรทัดนี้
		OrderBy("number_tree", firestore.Desc).
		OrderBy("score", firestore.Desc).
		Limit(10).
		Documents(ctx)
	defer iter.Stop()

	// 2. วนลูปเพื่ออ่านข้อมูลและสร้างผลลัพธ์ (ไม่ต้องนับ Rank แล้ว)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Printf("Failed to iterate leaderboard data: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch leaderboard data"})
			return
		}

		var user models.User
		if err := doc.DataTo(&user); err != nil {
			log.Printf("Failed to convert user data for leaderboard: %v", err)
			continue
		}

		leaderboard = append(leaderboard, LeaderboardEntry{
			Name:       user.Name,
			NumberTree: user.NumberTree,
			Score:      user.Score,
		})
	}

	// 3. ส่งข้อมูลที่ได้กลับไปให้ Frontend
	c.JSON(http.StatusOK, leaderboard)
}
