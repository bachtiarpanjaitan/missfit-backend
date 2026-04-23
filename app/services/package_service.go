package services

import (
	"missfit/app/dtos"
	"missfit/app/facades"
	"missfit/app/models"
	"missfit/app/utils"
)

type PackageServiceInterface interface {
	GetPackageById(id string, filters map[string]any) (*models.QuizPackage, error)
	GetActivePackage(isActive bool) (*models.QuizPackage, error)
	GetUserPurchasedPackage(userId string, packageId string) (*models.UserPurchasedPackage, error)
	GetUserPackages(userId string, pagination dtos.PaginationParams) (*[]models.UserPurchasedPackage, error)
	GetQuestionsByPackageId(packageId string) (*[]models.QuizQuestion, error)
	SubmitQuizResult(quizResult dtos.QuizResult) (*models.UserQuizAttempt, error)
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

func (s *PackageService) GetQuestionsByPackageId(packageId string) (*[]models.QuizQuestion, error) {
	questions := []models.QuizQuestion{}
	err := facades.Orm().Query().
		With("Options").
		Where("quiz_package_id", packageId).
		Order("question_order ASC").
		Find(&questions)
	if err != nil {
		return nil, err
	}
	return &questions, nil
}

func (s *PackageService) SubmitQuizResult(quizResult dtos.QuizResult) (*models.UserQuizAttempt, error) {
	var totalPoint float64 = 0
	var percentage float64 = 0
	var is_passed bool = false
	var time_taken_seconds, _ = utils.DiffSeconds(quizResult.StartedAt, quizResult.CompletedAt)
	var status string = "pending"

	pkg, err := s.GetPackageById(quizResult.PackageId, nil)
	if err != nil {
		return nil, err
	}
	if pkg == nil {
		return nil, nil
	}

	questions, err := s.GetQuestionsByPackageId(pkg.Id)
	if err != nil {
		return nil, err
	}

	answer_map := make(map[string]string)
	for _, answer := range quizResult.Answers {
		answer_map[answer.QuestionId] = answer.AnswerId
	}

	for _, question := range *questions {
		for _, option := range question.Options {
			if option.IsCorrect && answer_map[question.Id] == option.Id {
				totalPoint += float64(question.Point)
				percentage = (totalPoint / float64(len(*questions))) * 100
			}
		}
	}

	if totalPoint >= float64(pkg.PassingScore) {
		is_passed = true
		status = "passed"
	} else {
		is_passed = false
		status = "failed"
	}

	userQuizAttempt := models.UserQuizAttempt{
		UserId:           quizResult.UserId,
		QuizPackageId:    quizResult.PackageId,
		StartedAt:        utils.ToDate(quizResult.StartedAt),
		CompletedAt:      utils.ToDate(quizResult.CompletedAt),
		Score:            quizResult.Score,
		TotalPoints:      totalPoint,
		Percentage:       percentage,
		IsPassed:         &is_passed,
		TimeTakenSeconds: time_taken_seconds,
		Status:           status,
	}
	var attemps models.UserQuizAttempt
	err = facades.Orm().Query().Model(&models.UserQuizAttempt{}).Create(&userQuizAttempt)
	if err != nil {
		return nil, err
	}

	var userAnswers []models.UserQuizAnswer
	for _, question := range *questions {
		var correct bool = false
		var point float64 = 0
		for _, option := range question.Options {
			if option.IsCorrect && answer_map[question.Id] == option.Id {
				correct = true
				point = float64(question.Point)
			}
		}
		userAnswers = append(userAnswers, models.UserQuizAnswer{
			UserQuizAttemptId: userQuizAttempt.Id,
			SelectedOptionId:  answer_map[question.Id],
			IsCorrect:         correct,
			PointsEarned:      point,
		})
	}

	err = facades.Orm().Query().Model(&models.UserQuizAnswer{}).Create(&userAnswers)
	if err != nil {
		return nil, err
	}

	return &attemps, nil
}
