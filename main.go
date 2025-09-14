package main

import (
	"meerank/database"
	"meerank/routers"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	db, err := database.SetupDatabaseConnection()
	if err != nil {
		panic("Failed to connect to the database")
	}

	r := gin.Default()

	// ✨ แก้ไขการตั้งค่า CORS ตรงนี้ให้สมบูรณ์ ✨
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true // อนุญาตทุก Origin (สำหรับทดสอบ)

	// ❗ บรรทัดสำคัญที่สุดที่ขาดไป คือบรรทัดนี้ ❗
	// อนุญาตให้เบราว์เซอร์ส่ง Header เหล่านี้เข้ามาได้
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	
	r.Use(cors.New(config)) // ใช้งาน Middleware

	routers.SetupRouter(r, db)

	r.Run()
}