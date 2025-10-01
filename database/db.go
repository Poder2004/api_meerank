package database

import (
	"context"
	"log"
	"os"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
	"github.com/joho/godotenv"
	"google.golang.org/api/option"
)

func SetupFirestoreClient() (*firestore.Client, error) {
	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: .env file not found, will use environment variables from OS")
	}

	credentialsPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	if credentialsPath == "" {
		log.Fatal("GOOGLE_APPLICATION_CREDENTIALS environment variable not set.")
	}
	
	// ✨ 1. อ่านค่า Project ID จาก Environment Variable ✨
	projectID := os.Getenv("GOOGLE_PROJECT_ID")
	if projectID == "" {
		log.Fatal("GOOGLE_PROJECT_ID environment variable not set.")
	}

	ctx := context.Background()
	opt := option.WithCredentialsFile(credentialsPath)

	// ✨ 2. สร้าง Config เพื่อระบุ Project ID โดยตรง ✨
	conf := &firebase.Config{
		ProjectID: projectID,
	}

	// ✨ 3. ส่ง Config เข้าไปตอนสร้าง App ✨
	app, err := firebase.NewApp(ctx, conf, opt)
	if err != nil {
		log.Printf("Error initializing Firebase app: %v\n", err)
		return nil, err
	}

	client, err := app.Firestore(ctx)
	if err != nil {
		log.Printf("Error creating Firestore client: %v\n", err)
		return nil, err
	}
	
	log.Println("Successfully connected to Firestore.")
	return client, nil
}