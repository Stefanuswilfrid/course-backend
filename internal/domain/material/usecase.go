package material

import (
	"context"

	"log"

	"github.com/Stefanuswilfrid/course-backend/internal/apierror"
	"github.com/Stefanuswilfrid/course-backend/internal/domain/attachment"
	"github.com/Stefanuswilfrid/course-backend/internal/schema"
	"github.com/google/uuid"
)

type UseCase struct {
	repo              Repository
	attachmentUseCase *attachment.UseCase // Add this line
}

func NewUseCase(repo Repository, attachmentUseCase *attachment.UseCase) *UseCase {
	return &UseCase{repo: repo, attachmentUseCase: attachmentUseCase}
}

func (uc *UseCase) CreateMaterial(ctx context.Context, req CreateMaterialRequest) error {
	// Create a new material instance

	courseId, err := uuid.Parse(req.CourseID)

	if err != nil {
		return apierror.ErrInternalServer.Build()

	}
	id, err := uuid.NewV7()
	if err != nil {
		log.Println("Error generating UUID: ", err)
		return apierror.ErrInternalServer.Build()
	}
	mat := schema.Material{
		ID:          id,
		CourseID:    courseId,
		Title:       req.Title,
		Description: req.Description,
	}

	return uc.repo.Create(ctx, &mat)
}

func (uc *UseCase) GetMaterialByID(ctx context.Context, id uuid.UUID) (*schema.Material, error) {
	return uc.repo.GetByID(ctx, id)
}

func (uc *UseCase) GetAllMaterials(ctx context.Context) ([]*schema.Material, error) {
	return uc.repo.GetAll(ctx)
}

func (uc *UseCase) UpdateMaterial(ctx context.Context, req UpdateMaterialRequest, id uuid.UUID) error {
	mat, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return ErrMaterialNotFound.Build()
	}

	// Update the material fields from the request
	if req.Title != nil {
		mat.Title = *req.Title
	}
	if req.Description != nil {
		mat.Description = *req.Description
	}

	return uc.repo.Update(ctx, mat)
}

func (uc *UseCase) AddAttachment(ctx context.Context, id uuid.UUID, req AttachmentInput) error {
	mat, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return ErrMaterialNotFound.Build()
	}

	if req.File != nil {

		attachment, err := uc.attachmentUseCase.CreateAttachment(ctx, req.File, req.Description, id)
		if err != nil {

			return ErrS3UploadFail.Build()
		}
		mat.Attachments = append(mat.Attachments, attachment)
	}

	return uc.repo.Update(ctx, mat)
}

func (uc *UseCase) DeleteMaterial(ctx context.Context, id uuid.UUID) error {
	return uc.repo.Delete(ctx, id)
}
