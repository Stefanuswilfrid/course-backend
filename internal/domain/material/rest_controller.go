// file: material/rest_controller.go

package material

import (
	"net/http"

	"github.com/Stefanuswilfrid/course-backend/internal/apierror"
	"github.com/Stefanuswilfrid/course-backend/internal/domain/course"
	"github.com/Stefanuswilfrid/course-backend/internal/middleware"
	"github.com/Stefanuswilfrid/course-backend/internal/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type RestController struct {
	useCase       *UseCase
	courseUseCase *course.UseCase
}

func NewRestController(r *gin.Engine, uc *UseCase, cuc *course.UseCase) {
	c := &RestController{useCase: uc, courseUseCase: cuc}

	materialGroup := r.Group("/v1/materials")
	{
		materialGroup.POST("", middleware.Authenticate(), middleware.RequireRole("instructor"), c.create)
		materialGroup.GET("/:id", c.getByID)
		materialGroup.GET("/course/:id", c.getMaterialByCourse)
		materialGroup.GET("", c.getAll)
		materialGroup.PUT("/:id", middleware.Authenticate(), middleware.RequireRole("instructor"), c.update)
		materialGroup.DELETE("/:id", middleware.Authenticate(), middleware.RequireRole("instructor"), c.delete)
		materialGroup.POST("addAttachment/:id", middleware.Authenticate(), middleware.RequireRole("instructor"), c.addAttachment)
	}

}

func (c *RestController) create(ctx *gin.Context) {

	var req CreateMaterialRequest
	if err := ctx.ShouldBind(&req); err != nil {
		response.NewRestResponse(http.StatusBadRequest, "Invalid material data: "+err.Error(), nil).Send(ctx)
		return
	}

	userID, exists := ctx.Get("user.id")
	if !exists {
		response.NewRestResponse(http.StatusInternalServerError, "User ID not found in request context", nil).Send(ctx)
		return
	}

	// Check if the course exists and is owned by the current user
	courseId, err := uuid.Parse(req.CourseID)

	if err != nil {
		err2 := apierror.ErrInternalServer.Build()
		response.NewRestResponse(apierror.GetHttpStatus(err2), err2.Error(), err.Error()).Send(ctx)
	}

	course, err := c.courseUseCase.GetByID(ctx, courseId)
	if err != nil {
		response.NewRestResponse(http.StatusInternalServerError, "Failed to fetch course: "+err.Error(), nil).Send(ctx)
		return
	}

	if course.InstructorID.String() != userID {
		response.NewRestResponse(http.StatusForbidden, "Only the owner of the course can add materials", nil).Send(ctx)
		return
	}

	if err := c.useCase.CreateMaterial(ctx, req); err != nil {
		response.NewRestResponse(apierror.GetHttpStatus(err), err.Error(), apierror.GetPayload(err)).Send(ctx)
		return
	}
	response.NewRestResponse(http.StatusCreated, "Create material successfully", nil).Send(ctx)
}

func (c *RestController) getByID(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		err = apierror.ErrInvalidParamId.Build()
		response.NewRestResponse(apierror.GetHttpStatus(err), err.Error(), err.Error()).Send(ctx)
		return
	}
	mat, err := c.useCase.GetMaterialByID(ctx, id)
	if err != nil {
		response.NewRestResponse(apierror.GetHttpStatus(err), err.Error(), apierror.GetPayload(err)).Send(ctx)
		return

	}
	response.NewRestResponse(http.StatusOK, "All Material Retrieve", mat).Send(ctx)
}

func (c *RestController) getAll(ctx *gin.Context) {
	mats, err := c.useCase.GetAllMaterials(ctx)
	if err != nil {
		response.NewRestResponse(apierror.GetHttpStatus(err), err.Error(), apierror.GetPayload(err)).Send(ctx)
		return
	}
	response.NewRestResponse(http.StatusOK, "All Material Retrieve", mats).Send(ctx)
}

func (c *RestController) getMaterialByCourse(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		err = apierror.ErrInvalidParamId.Build()
		response.NewRestResponse(apierror.GetHttpStatus(err), err.Error(), err.Error()).Send(ctx)
		return
	}

	course, err := c.courseUseCase.GetByID(ctx, id)
	if err != nil {
		response.NewRestResponse(http.StatusInternalServerError, "Failed to fetch material: "+err.Error(), nil).Send(ctx)
		return
	}

	response.NewRestResponse(http.StatusOK, "All Course Material Retrieve", course.Materials).Send(ctx)

}
func (c *RestController) update(ctx *gin.Context) {

	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		err = apierror.ErrInvalidParamId.Build()
		response.NewRestResponse(apierror.GetHttpStatus(err), err.Error(), err.Error()).Send(ctx)
		return
	}

	err = c.verifyMaterialOwnership(ctx, id)
	if err != nil {
		response.NewRestResponse(apierror.GetHttpStatus(err), err.Error(), nil).Send(ctx)
		return
	}

	var req UpdateMaterialRequest
	if err := ctx.ShouldBind(&req); err != nil {
		response.NewRestResponse(http.StatusBadRequest, "Invalid material data: "+err.Error(), nil).Send(ctx)
		return

	}

	if err := c.useCase.UpdateMaterial(ctx, req, id); err != nil {
		response.NewRestResponse(apierror.GetHttpStatus(err), err.Error(), apierror.GetPayload(err)).Send(ctx)
		return
	}
	response.NewRestResponse(http.StatusCreated, "Update Material successfully", nil).Send(ctx)
}

func (c *RestController) delete(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		err = apierror.ErrInvalidParamId.Build()
		response.NewRestResponse(apierror.GetHttpStatus(err), err.Error(), err.Error()).Send(ctx)
		return
	}

	err = c.verifyMaterialOwnership(ctx, id)
	if err != nil {
		response.NewRestResponse(apierror.GetHttpStatus(err), err.Error(), nil).Send(ctx)
		return
	}

	if err := c.useCase.DeleteMaterial(ctx, id); err != nil {
		response.NewRestResponse(apierror.GetHttpStatus(err), err.Error(), apierror.GetPayload(err)).Send(ctx)
		return
	}
	response.NewRestResponse(http.StatusOK, "Delete Material successfully", nil).Send(ctx)
}

func (c *RestController) addAttachment(ctx *gin.Context) {

	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		err = apierror.ErrInvalidParamId.Build()
		response.NewRestResponse(apierror.GetHttpStatus(err), err.Error(), err.Error()).Send(ctx)
		return
	}

	err = c.verifyMaterialOwnership(ctx, id)
	if err != nil {
		response.NewRestResponse(apierror.GetHttpStatus(err), err.Error(), nil).Send(ctx)
		return
	}

	var req AttachmentInput
	if err := ctx.ShouldBind(&req); err != nil {
		response.NewRestResponse(http.StatusBadRequest, "Invalid attachment data: "+err.Error(), nil).Send(ctx)
		return
	}

	if err := c.useCase.AddAttachment(ctx, id, req); err != nil {
		response.NewRestResponse(apierror.GetHttpStatus(err), err.Error(), apierror.GetPayload(err)).Send(ctx)
		return
	}
	response.NewRestResponse(http.StatusOK, "Add attachment successfully", nil).Send(ctx)
}

func (c *RestController) verifyMaterialOwnership(ctx *gin.Context, materialID uuid.UUID) error {

	mat, err := c.useCase.GetMaterialByID(ctx, materialID)
	if err != nil {
		return err
	}

	courseData, err := c.courseUseCase.GetByID(ctx, mat.CourseID)
	if err != nil {
		return err
	}
	instructorID, exists := ctx.Get("user.id")
	if !exists {
		return ErrUnauthorizedAccess.Build()
	}
	if courseData.InstructorID.String() != instructorID {
		return ErrNotOwnerAccess.Build()
	}

	return nil
}
