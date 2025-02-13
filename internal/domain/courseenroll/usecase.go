package courseenroll

import (
	"context"
	"time"

	"github.com/Stefanuswilfrid/course-backend/internal/apierror"
	"github.com/Stefanuswilfrid/course-backend/internal/schema"
	"github.com/google/uuid"
)

type UseCase struct {
	repo Repository
}

func NewUseCase(repo Repository) *UseCase {
	return &UseCase{repo: repo}
}

func (uc *UseCase) EnrollStudent(ctx context.Context, userID, courseID uuid.UUID) error {
	id, err := uuid.NewV7()
	if err != nil {
		return apierror.ErrInternalServer.Build()
	}
	enroll := schema.CourseEnroll{
		ID:        id,
		UserID:    userID,
		CourseID:  courseID,
		CreatedAt: time.Now(),
	}
	return uc.repo.Create(ctx, &enroll)
}

func (uc *UseCase) GetEnrollmentsByCourse(ctx context.Context, courseID uuid.UUID) ([]schema.User, error) {
	return uc.repo.GetUsersByCourseID(ctx, courseID)
}

func (uc *UseCase) GetEnrollmentsByUser(ctx context.Context, userID uuid.UUID) ([]schema.Course, error) {
	return uc.repo.GetCoursesByUserID(ctx, userID)
}

func (uc *UseCase) CheckEnrollment(ctx context.Context, userID, courseID uuid.UUID) (bool, error) {
	return uc.repo.IsEnrolled(ctx, userID, courseID)
}
