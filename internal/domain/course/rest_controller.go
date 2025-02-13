package course

import (
	"net/http"

	"github.com/Stefanuswilfrid/course-backend/internal/apierror"
	"github.com/Stefanuswilfrid/course-backend/internal/domain/wallet"
	"github.com/Stefanuswilfrid/course-backend/internal/middleware"
	"github.com/Stefanuswilfrid/course-backend/internal/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type RestController struct {
	uc  *UseCase
	wuc *wallet.UseCase
}

func NewRestController(router *gin.Engine, uc *UseCase, wuc *wallet.UseCase) {

	controller := &RestController{uc: uc, wuc: wuc}

	courseGroup := router.Group("/v1/courses")
	{
		courseGroup.GET("", controller.GetAll())
		courseGroup.GET("/:id", controller.GetByID())
		courseGroup.POST("",
			middleware.Authenticate(),
			middleware.RequireEmailVerified(),
			middleware.RequireRole("instructor"),
			controller.Create(),
		)
		courseGroup.PUT("/:id", middleware.Authenticate(), middleware.RequireRole("instructor"), controller.Update())
		courseGroup.POST("/buy/:id", middleware.Authenticate(), middleware.RequireEmailVerified(), middleware.RequireRole("student"), controller.BuyCourse())
		courseGroup.GET("/instructor/:id", middleware.Authenticate(), controller.GetInstructorCourse())
		courseGroup.DELETE("/:id",
			middleware.Authenticate(),
			middleware.RequireRole("instructor"),
			controller.Delete(),
		)

		courseGroup.GET("/popularity", controller.GetByPopularity())
		courseGroup.GET("/mycourse", middleware.Authenticate(), controller.GetUserEnrollments())
		courseGroup.GET("/usersEnroll/:courseId", middleware.Authenticate(), controller.GetCourseEnrollments())
		courseGroup.GET("/progress/:courseId", middleware.Authenticate(), middleware.RequireEmailVerified(), controller.GetStudentProgress())
		courseGroup.GET("/search", controller.SearchCourses())
		courseGroup.GET("/filter", controller.FilterCourse())
	}

}

func (c *RestController) GetAll() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req PaginationRequest
		if err := ctx.ShouldBindQuery(&req); err != nil {
			response.NewRestResponse(http.StatusBadRequest, "Invalid pagination parameters", nil).Send(ctx)
			return
		}
		result, err := c.uc.GetAll(ctx, req.Page, req.Limit)
		if err != nil {
			response.NewRestResponse(http.StatusInternalServerError, "Failed to retrieve courses", nil).Send(ctx)
			return
		}
		response.NewRestResponse(http.StatusOK, "Courses retrieved successfully", result).Send(ctx)
	}
}

func (c *RestController) GetByPopularity() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req PaginationRequest
		if err := ctx.ShouldBindQuery(&req); err != nil {
			response.NewRestResponse(http.StatusBadRequest, "Invalid pagination parameters", nil).Send(ctx)
			return
		}
		result, err := c.uc.GetCourseByPopularity(ctx, req.Page, req.Limit)
		if err != nil {
			response.NewRestResponse(http.StatusInternalServerError, "Failed to retrieve courses", nil).Send(ctx)
			return
		}
		response.NewRestResponse(http.StatusOK, "Courses retrieved successfully", result).Send(ctx)
	}
}

func (c *RestController) SearchCourses() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req SearchPaginationRequest
		if err := ctx.ShouldBindQuery(&req); err != nil {
			response.NewRestResponse(http.StatusBadRequest, "Invalid request parameters: "+err.Error(), nil).Send(ctx)
			return
		}

		result, err := c.uc.SearchCoursesByTitle(ctx, req.Title, req.Page, req.Limit)
		if err != nil {
			response.NewRestResponse(http.StatusInternalServerError, "Failed to search courses", err.Error()).Send(ctx)
			return
		}

		response.NewRestResponse(http.StatusOK, "Courses found", result).Send(ctx)
	}
}

func (c *RestController) GetInstructorCourse() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		instructorID, err := uuid.Parse(ctx.Param("id"))
		if err != nil {
			response.NewRestResponse(http.StatusBadRequest, "Invalid Instructor ID", nil).Send(ctx)
			return
		}

		var req PaginationRequest
		if err := ctx.ShouldBindQuery(&req); err != nil {
			response.NewRestResponse(http.StatusBadRequest, "Invalid pagination parameters", nil).Send(ctx)
			return
		}

		result, err := c.uc.GetByInstructorID(ctx, instructorID, req.Page, req.Limit)
		if err != nil {
			response.NewRestResponse(apierror.GetHttpStatus(err), err.Error(), apierror.GetPayload(err)).Send(ctx)
			return
		}

		if len(result.Courses) == 0 {
			response.NewRestResponse(http.StatusOK, "No courses found for this instructor", nil).Send(ctx)
			return
		}

		response.NewRestResponse(http.StatusOK, "Courses retrieved successfully", result).Send(ctx)
	}
}

func (c *RestController) GetCourseEnrollments() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		courseID, err := uuid.Parse(ctx.Param("courseId"))
		if err != nil {
			response.NewRestResponse(http.StatusBadRequest, "Invalid course ID", nil).Send(ctx)
			return
		}

		users, err := c.uc.GetEnrollmentsByCourse(ctx, courseID)
		if err != nil {
			response.NewRestResponse(http.StatusInternalServerError, "Failed to retrieve enrollments", err.Error()).Send(ctx)
			return
		}

		if len(users) == 0 {
			response.NewRestResponse(http.StatusOK, "No enrollments found for this course", nil).Send(ctx)
			return
		}

		response.NewRestResponse(http.StatusOK, "Enrollments retrieved successfully", users).Send(ctx)
	}
}

func (c *RestController) GetUserEnrollments() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		studentID, exists := ctx.Get("user.id")
		if !exists {
			response.NewRestResponse(http.StatusInternalServerError, "Failed to get student ID from context", nil).Send(ctx)
			return
		}
		courses, err := c.uc.GetEnrollmentsByUser(ctx, studentID.(string))
		if err != nil {
			response.NewRestResponse(http.StatusInternalServerError, "Failed to retrieve enrollments", err.Error()).Send(ctx)
			return
		}

		if len(courses) == 0 {
			response.NewRestResponse(http.StatusOK, "User not have purchased course", nil).Send(ctx)
			return
		}

		response.NewRestResponse(http.StatusOK, "Enrollments retrieved successfully", courses).Send(ctx)
	}
}

func (c *RestController) BuyCourse() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id, err := uuid.Parse(ctx.Param("id"))
		if err != nil {
			response.NewRestResponse(http.StatusBadRequest, "Invalid ID", nil).Send(ctx)
			return
		}

		studentID, exists := ctx.Get("user.id")
		if !exists {
			response.NewRestResponse(http.StatusInternalServerError, "Failed to get student ID from context", nil).Send(ctx)
			return
		}

		err = c.uc.BuyCourse(ctx, id, studentID.(string))
		if err != nil {
			response.NewRestResponse(apierror.GetHttpStatus(err), err.Error(), apierror.GetPayload(err)).Send(ctx)
			return
		}

		response.NewRestResponse(http.StatusOK, "Buy Course successfully", nil).Send(ctx)
	}
}

func (c *RestController) GetByID() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id, err := uuid.Parse(ctx.Param("id"))
		if err != nil {
			response.NewRestResponse(http.StatusBadRequest, "Invalid ID", nil).Send(ctx)
			return
		}

		course, err := c.uc.GetByID(ctx, id)
		if err != nil {
			response.NewRestResponse(http.StatusInternalServerError, err.Error(), nil).Send(ctx)
			return
		}
		response.NewRestResponse(http.StatusOK, "Course retrieved successfully", course).Send(ctx)
	}
}

func (c *RestController) Create() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		var req CreateCourseRequest
		if err := ctx.ShouldBind(&req); err != nil {
			response.NewRestResponse(http.StatusBadRequest, "Invalid course data: "+err.Error(), nil).Send(ctx)
			return
		}

		imageFile, errImage := ctx.FormFile("image")
		if errImage != nil && errImage != http.ErrMissingFile {
			response.NewRestResponse(http.StatusBadRequest, "Could not retrieve image file", nil).Send(ctx)
			return
		}

		syllabusFile, errSyllabus := ctx.FormFile("syllabus")
		if errSyllabus != nil && errSyllabus != http.ErrMissingFile {
			response.NewRestResponse(http.StatusBadRequest, "Could not retrieve syllabus file", nil).Send(ctx)
			return
		}

		instructorID, exists := ctx.Get("user.id")
		if !exists {
			response.NewRestResponse(http.StatusInternalServerError, "Failed to get instructor ID from context", nil).Send(ctx)
			return
		}

		err := c.uc.Create(ctx, req, imageFile, syllabusFile, instructorID.(string))
		if err != nil {
			response.NewRestResponse(apierror.GetHttpStatus(err), err.Error(), apierror.GetPayload(err)).Send(ctx)
			return
		}

		response.NewRestResponse(http.StatusCreated, "Course created successfully", nil).Send(ctx)
	}
}

func (c *RestController) Update() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		id, err := uuid.Parse(ctx.Param("id"))
		if err != nil {
			response.NewRestResponse(http.StatusBadRequest, "Invalid ID", nil).Send(ctx)
			return
		}

		var req UpdateCourseRequest
		if err := ctx.ShouldBind(&req); err != nil {
			response.NewRestResponse(http.StatusBadRequest, "Invalid course data: "+err.Error(), nil).Send(ctx)
			return
		}

		err = c.checkCourseOwnership(ctx, id)
		if err != nil {
			response.NewRestResponse(apierror.GetHttpStatus(err), err.Error(), apierror.GetPayload(err)).Send(ctx)
			return
		}

		imageFile, _ := ctx.FormFile("image")
		syllabusFile, _ := ctx.FormFile("syllabus")

		updatedCourse, err := c.uc.Update(ctx.Request.Context(), req, id, imageFile, syllabusFile)
		if err != nil {
			response.NewRestResponse(apierror.GetHttpStatus(err), err.Error(), apierror.GetPayload(err)).Send(ctx)
			return
		}

		response.NewRestResponse(http.StatusOK, "Course updated successfully", updatedCourse).Send(ctx)
	}
}

func (c *RestController) Delete() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id, err := uuid.Parse(ctx.Param("id"))
		if err != nil {
			response.NewRestResponse(http.StatusBadRequest, "Invalid ID", nil).Send(ctx)
			return
		}

		err = c.checkCourseOwnership(ctx, id)
		if err != nil {
			response.NewRestResponse(apierror.GetHttpStatus(err), err.Error(), apierror.GetPayload(err)).Send(ctx)
			return
		}

		err = c.uc.Delete(ctx, id)
		if err != nil {
			response.NewRestResponse(apierror.GetHttpStatus(err), err.Error(), apierror.GetPayload(err)).Send(ctx)
			return
		}
		response.NewRestResponse(http.StatusOK, "Course deleted successfully", nil).Send(ctx)
	}
}

func (c *RestController) checkCourseOwnership(ctx *gin.Context, courseID uuid.UUID) error {
	instructorID, exists := ctx.Get("user.id")
	if !exists {
		return ErrUnauthorizedAccess.Build()
	}
	course, err := c.uc.GetByID(ctx, courseID)
	if err != nil {
		return err
	}

	if course.InstructorID.String() != instructorID {
		return ErrNotOwnerAccess.Build()
	}
	return nil
}

func (c *RestController) GetStudentProgress() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		var req CourseProgress
		if err := ctx.ShouldBindQuery(&req); err != nil {
			response.NewRestResponse(http.StatusBadRequest, "Invalid user id: "+err.Error(), nil).Send(ctx)
			return
		}
		courseID, err := uuid.Parse(ctx.Param("courseId"))
		if err != nil {
			response.NewRestResponse(http.StatusBadRequest, "Invalid ID", nil).Send(ctx)
			return
		}

		progress, err := c.uc.GetUserCourseProgress(ctx, courseID, req.UserId)
		if err != nil {
			response.NewRestResponse(apierror.GetHttpStatus(err), err.Error(), apierror.GetPayload(err)).Send(ctx)
			return
		}

		response.NewRestResponse(http.StatusOK, "Student progress retrieve successfully", progress).Send(ctx)
	}
}

func (c *RestController) FilterCourse() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		var req FilterCoursesRequest
		if err := ctx.ShouldBindQuery(&req); err != nil {
			response.NewRestResponse(http.StatusBadRequest, "Invalid user id: "+err.Error(), nil).Send(ctx)
			return
		}

		result, err := c.uc.FilterCourses(ctx, req)
		if err != nil {
			response.NewRestResponse(apierror.GetHttpStatus(err), err.Error(), apierror.GetPayload(err)).Send(ctx)
			return
		}

		response.NewRestResponse(http.StatusOK, "Course retrieve successfully", result).Send(ctx)
	}
}
