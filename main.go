package main

import (
	"log"
	"meerank/database"
	"meerank/routers"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// 1. เปลี่ยนไปเรียกใช้ฟังก์ชันเชื่อมต่อ Firestore
	firestoreClient, err := database.SetupFirestoreClient()
	if err != nil {
		// ใช้ log.Fatalf จะแสดง error และหยุดการทำงานทันที
		log.Fatalf("Failed to connect to Firestore: %v", err)
	}
	// 2. เพิ่ม defer เพื่อปิดการเชื่อมต่อเมื่อจบการทำงาน
	defer firestoreClient.Close()

	r := gin.Default()

	// ✨ ส่วนของการตั้งค่า CORS ของคุณถูกต้องดีแล้ว ไม่ต้องแก้ไขครับ ✨
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true 
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	r.Use(cors.New(config))

	// 3. ส่ง firestoreClient (ตัวใหม่) เข้าไปใน SetupRouter แทนที่ db (ตัวเก่า)
	routers.SetupRouter(r, firestoreClient)

	// รันเซิร์ฟเวอร์ (แนะนำให้ระบุ port)
	r.Run(":8080") 
}