package courseenroll

import (
	"net/http"

	"github.com/Stefanuswilfrid/course-backend/internal/apierror"
)

var (
	ErrNotEnrolled = apierror.NewApiErrorBuilder().
		WithHttpStatus(http.StatusBadRequest).
		WithMessage("COURSE_NOT_ENROLLED")
)
