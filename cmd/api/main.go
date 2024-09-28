package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Stefanuswilfrid/course-backend/internal/config"
	"github.com/Stefanuswilfrid/course-backend/internal/middleware"
	"github.com/Stefanuswilfrid/course-backend/internal/schema"

	"github.com/joho/godotenv"
)

// go run ./cmd/api
func main() {
	err := godotenv.Load()
	apiEnv := os.Getenv("ENV")
	if err != nil && apiEnv == "" {
		log.Println("fail to load env", err)
	}
	config.LoadEnv()

	db := config.NewPostgresql(
		&schema.Notification{},
		&schema.Wallet{},
		&schema.MidtransTransaction{},
		&schema.User{},
		&schema.Course{},
		&schema.Material{},
		&schema.Assignment{},
		&schema.Submission{},
		&schema.Attachment{},
		&schema.Review{},
		&schema.CourseEnroll{},
		&schema.ForumDiscussion{},
		&schema.ForumReply{},
	)

	engine := config.NewGin()
	engine.Use(middleware.CORS())

	if err != nil {
		log.Println("fail to connect s3 bucket", err)
	}

	fmt.Println("Hello, world.")

	notificationRepo := notification.NewRepository(db)
	notificationUseCase := notification.NewUseCase(notificationRepo)
	notification.NewRestController(engine, notificationUseCase)

	// Wallet
	walletRepo := wallet.NewRepository(db)
	walletUseCase := wallet.NewUseCase(walletRepo, nil)
	midtUseCase := wallet.NewMidtransUseCase(walletUseCase)
	walletUseCase.MidtUc = midtUseCase
	wallet.NewRestController(engine, walletUseCase, midtUseCase)

	// User
	userRepo := user.NewRepository(db, walletRepo)
	userUseCase := user.NewUseCase(userRepo, uploader)
	user.NewRestController(engine, userUseCase)

	// Auth
	authRepo := auth.NewRepository(rds)
	authUseCase := auth.NewUseCase(authRepo, userRepo, mailDialer)
	auth.NewRestController(engine, authUseCase)

}
