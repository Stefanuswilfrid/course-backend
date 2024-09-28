package user

import (
	"net/http"

	"github.com/Stefanuswilfrid/course-backend/internal/apierror"
)

var (
	ErrUserNotFound = apierror.NewApiErrorBuilder().
		WithHttpStatus(http.StatusNotFound).
		WithMessage("USER_NOT_FOUND")
)
