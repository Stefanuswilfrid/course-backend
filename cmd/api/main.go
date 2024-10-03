package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Stefanuswilfrid/course-backend/internal/config"
	"github.com/Stefanuswilfrid/course-backend/internal/domain/assignment"
	"github.com/Stefanuswilfrid/course-backend/internal/domain/attachment"
	"github.com/Stefanuswilfrid/course-backend/internal/domain/auth"
	"github.com/Stefanuswilfrid/course-backend/internal/domain/course"
	"github.com/Stefanuswilfrid/course-backend/internal/domain/courseenroll"
	"github.com/Stefanuswilfrid/course-backend/internal/domain/forum"
	"github.com/Stefanuswilfrid/course-backend/internal/domain/material"
	"github.com/Stefanuswilfrid/course-backend/internal/domain/notification"
	"github.com/Stefanuswilfrid/course-backend/internal/domain/review"
	"github.com/Stefanuswilfrid/course-backend/internal/domain/submission"
	"github.com/Stefanuswilfrid/course-backend/internal/domain/user"
	"github.com/Stefanuswilfrid/course-backend/internal/domain/wallet"
	"github.com/Stefanuswilfrid/course-backend/internal/middleware"
	"github.com/Stefanuswilfrid/course-backend/internal/schema"

	"github.com/joho/godotenv"
)

// 8080
// System: PostgreSQL
// Server: postgres (This is the hostname you defined in the Docker compose file under hostname: postgres-server, but Adminer needs to communicate with the Docker container, so you use postgres which is the default service name for communication.)
// Username: postgres (Defined in POSTGRESQL_USERNAME)
// Password: postgres (Defined in POSTGRESQL_PASSWORD)
// Database: db (Defined in POSTGRESQL_DATABASE)

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

	mailDialer := config.NewMailDialer()
	config.SetupMidtrans()

	engine := config.NewGin()
	engine.Use(middleware.CORS())

	uploader, err := config.InitializeS3()

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
	authRepo := auth.NewRepository()
	authUseCase := auth.NewUseCase(authRepo, userRepo, mailDialer)
	auth.NewRestController(engine, authUseCase)

	courseEnrollRepo := courseenroll.NewRepository(db)
	courseEnrollUseCase := courseenroll.NewUseCase(courseEnrollRepo)

	// Course
	courseRepo := course.NewRepository(db)
	courseUseCase := course.NewUseCase(courseRepo, walletRepo, *courseEnrollUseCase, userRepo, notificationRepo, mailDialer, uploader)
	course.NewRestController(engine, courseUseCase, walletUseCase)

	// Attachment
	attachmentRepo := attachment.NewRepository(db)
	attachmentUseCase := attachment.NewUseCase(attachmentRepo, uploader)
	attachment.NewRestController(engine, attachmentUseCase)

	assignmentRepo := assignment.NewRepository(db)
	assignmentUseCase := assignment.NewUseCase(assignmentRepo, attachmentUseCase)
	assignment.NewRestController(engine, assignmentUseCase, courseUseCase)

	// Submission
	submissionRepo := submission.NewRepository(db)
	submissionUseCase := submission.NewUseCase(submissionRepo, assignmentRepo, *attachmentUseCase, courseRepo,
		courseEnrollRepo, userRepo, notificationRepo, mailDialer)
	submission.NewRestController(engine, submissionUseCase)

	materialRepo := material.NewRepository(db)
	materialUsecase := material.NewUseCase(materialRepo, attachmentUseCase)
	material.NewRestController(engine, materialUsecase, courseUseCase)

	reviewRepo := review.NewRepository(db)
	reviewUseCase := review.NewUseCase(reviewRepo, courseRepo, courseEnrollUseCase)
	review.NewRestController(engine, reviewUseCase)

	// Forum
	forumRepo := forum.NewRepository(db)
	forumUseCase := forum.NewUseCase(forumRepo, courseEnrollUseCase, courseRepo)
	forum.NewRestController(engine, forumUseCase)

	if err := engine.Run(":" + config.Env.ApiPort); err != nil {
		log.Fatalln(err)
	}

}
