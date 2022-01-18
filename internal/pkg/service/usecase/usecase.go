package usecase

import (
	"forum/internal/models"
	"forum/internal/pkg/service/repository"
)

type ServiceUsecase struct {
	sr *repository.ServiceRepository
}

func NewServiceUsecase(sr *repository.ServiceRepository) *ServiceUsecase {
	return &ServiceUsecase{
		sr: sr,
	}
}

func (su *ServiceUsecase) GetStatus() (*models.Status, error) {
	return su.sr.CountAll()
}

func (su *ServiceUsecase) Clear() error {
	return su.sr.Clear()
}
