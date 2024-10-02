package review

import (
	"net/http"

	"github.com/Stefanuswilfrid/course-backend/internal/apierror"
)

var (
	ErrCourseAlreadyReviewed = apierror.NewApiErrorBuilder().
					WithHttpStatus(http.StatusConflict).
					WithMessage("COURSE_ALREADY_REVIEWED")

	ErrReviewNotFound = apierror.NewApiErrorBuilder().
				WithHttpStatus(http.StatusNotFound).
				WithMessage("REVIEW_NOT_FOUND")
)
