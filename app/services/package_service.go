package services

import (
	"missfit/app/dtos"
	"missfit/app/facades"
	"missfit/app/models"
)

type PackageServiceInterface interface {
	GetPackageById(id string, filters map[string]any) (*models.QuizPackage, error)
	GetActivePackage(isActive bool) (*models.QuizPackage, error)
	GetUserPurchasedPackage(userId string, packageId string) (*models.UserPurchasedPackage, error)
	GetUserPackages(userId string, pagination dtos.PaginationParams) (*[]models.UserPurchasedPackage, error)
}

type PackageService struct {
}

func NewPackageService() PackageServiceInterface {
	return &PackageService{}
}

func (s *PackageService) GetPackageById(id string, filters map[string]any) (*models.QuizPackage, error) {
	quizPackage := models.QuizPackage{}
	query := facades.Orm().Query().Where("id", id)
	for key, value := range filters {
		query = query.Where(key, value)
	}
	err := query.First(&quizPackage)
	if err != nil {
		return nil, err
	}
	return &quizPackage, nil
}

func (s *PackageService) GetActivePackage(isActive bool) (*models.QuizPackage, error) {
	quizPackage := models.QuizPackage{}
	err := facades.Orm().Query().Where("is_active", isActive).First(&quizPackage)
	return &quizPackage, err
}

func (s *PackageService) GetUserPackages(userId string, pagination dtos.PaginationParams) (*[]models.UserPurchasedPackage, error) {
	userPackages := []models.UserPurchasedPackage{}
	offset := (pagination.Page - 1) * pagination.Limit
	err := facades.Orm().Query().With("QuizPackage").Where("user_id", userId).Where("is_active", true).Offset(offset).Limit(pagination.Limit).Order(pagination.Sort + " " + pagination.Order).Find(&userPackages)
	if err != nil {
		return nil, err
	}
	return &userPackages, nil
}

func (s *PackageService) GetUserPurchasedPackage(userId string, packageId string) (*models.UserPurchasedPackage, error) {
	userPurchasedPackage := models.UserPurchasedPackage{}
	err := facades.Orm().Query().Where("user_id", userId).Where("quiz_package_id", packageId).First(&userPurchasedPackage)
	if err != nil {
		return nil, err
	}
	return &userPurchasedPackage, nil
}
