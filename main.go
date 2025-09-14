package main

import (
	"meerank/database"
	"meerank/routers"

	"github.com/gin-contrib/cors" // ✨ 1. Import CORS
	"github.com/gin-gonic/gin"
)

func main() {
	db, err := database.SetupDatabaseConnection()
	if err != nil {
		panic("Failed to connect to the database")
	}

	r := gin.Default()

	// ✨ 2. ใช้งาน CORS Middleware ✨
	// ใช้ config ที่อนุญาตทุกอย่างไปก่อนเพื่อง่ายต่อการทดสอบ
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	// หรือจะระบุเฉพาะโดเมนของ Frontend ก็ได้ เช่น
	// config.AllowOrigins = []string{"http://localhost:4200", "https://your-angular-app.vercel.app"}

	r.Use(cors.New(config)) // 👈 ใช้งาน Middleware ที่นี่

	// เรียกใช้ routes จาก package routers (เหมือนเดิม)
	routers.SetupRouter(r, db)

	// เริ่มรันเซิร์ฟเวอร์
	// Render.com จะจัดการ port ให้เอง เราอาจจะลบ :8080 ออกหรือปล่อยไว้ก็ได้
	r.Run()
}
