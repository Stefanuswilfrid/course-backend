package forum

import (
	"net/http"

	"github.com/Stefanuswilfrid/course-backend/internal/apierror"
)

var (
	ErrDiscussionNotFound = apierror.NewApiErrorBuilder().
				WithHttpStatus(http.StatusNotFound).
				WithMessage("DISCUSSION_NOT_FOUND")

	ErrReplyNotFound = apierror.NewApiErrorBuilder().
				WithHttpStatus(http.StatusNotFound).
				WithMessage("REPLY_NOT_FOUND")
)
