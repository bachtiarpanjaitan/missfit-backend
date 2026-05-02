package services

import (
	"math"
	"missfit/app/dtos"
	"missfit/app/facades"
	"missfit/app/models"
	"missfit/app/utils"
	"time"
)

type PackageServiceInterface interface {
	GetPackageById(id string, filters map[string]any) (*models.QuizPackage, error)
	GetActivePackage(isActive bool) (*models.QuizPackage, error)
	GetUserPurchasedPackage(userId string, packageId string) (*models.UserPurchasedPackage, error)
	GetUserPackages(userId string, pagination dtos.PaginationParams) (*[]models.UserPurchasedPackage, error)
	GetQuestionsByPackageId(packageId string) (*[]models.QuizQuestion, error)
	SubmitQuizResult(quizResult dtos.QuizResult) (*models.UserQuizAttempt, error)
	GetUserResults(userId string) ([]dtos.MyQuizResult, error)
	HasMaxAttempts(userId string, packageId string) (bool, error)
	GetGlobalRankings(limit int) (*[]dtos.Ranking, error)
	GetMyRank(packageId string) (*dtos.Ranking, error)
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
	var totalWeight float64 = 0
	var percentage float64 = 0
	var is_passed bool = false
	var time_taken_seconds int64 = utils.DiffSeconds(quizResult.StartedAt, quizResult.CompletedAt)
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

	var correctAnswers int = 0
	var wrongAnswers int = 0
	var skipAnswers int = 0
	for _, question := range *questions {
		totalWeight += float64(question.Point)
		if answer_map[question.Id] == "skipped" {
			skipAnswers++
			continue
		}
		for _, option := range question.Options {
			if option.IsCorrect && answer_map[question.Id] == option.Id {
				totalPoint += float64(question.Point)
				correctAnswers++
			}

			if option.IsCorrect {
				if answer_map[question.Id] != option.Id {
					wrongAnswers++
				}
			}
		}
	}
	// println(utils.ToJson(totalPoint), totalWeight)
	// panic("stop")
	percentage = (totalPoint / totalWeight) * 100
	if totalPoint >= float64(pkg.PassingScore) {
		is_passed = true
		status = "passed"
	} else {
		is_passed = false
		status = "failed"
	}

	// println(totalPoint, totalWeight, percentage)
	// panic("stop")
	quizAttempt := models.UserQuizAttempt{
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
		CorrectAnswers:   correctAnswers,
		WrongAnswers:     wrongAnswers,
		SkipAnswers:      skipAnswers,
	}

	err = facades.Orm().Query().Model(&models.UserQuizAttempt{}).Create(&quizAttempt)
	if err != nil {
		return nil, err
	}

	//save ranking
	ranking := models.Ranking{}
	exist := facades.Orm().Query().Where("user_id = ? AND quiz_package_id = ?", quizResult.UserId, quizResult.PackageId).First(&ranking)
	if exist == nil && ranking.Id != "" {
		ranking.TotalPoints = totalPoint
		ranking.LastUpdated = time.Now()
		facades.Orm().Query().Where("id = ?", ranking.Id).Save(&ranking)
	} else {
		ranking = models.Ranking{
			UserId:        quizResult.UserId,
			QuizPackageId: quizResult.PackageId,
			TotalPoints:   totalPoint,
			CreatedAt:     time.Now(),
			LastUpdated:   time.Now(),
		}
		err = facades.Orm().Query().Model(&models.Ranking{}).Create(&ranking)
		if err != nil {
			return nil, err
		}
	}

	UserQuizAttempModel := models.UserQuizAttempt{}
	quizAttempData := facades.Orm().Query().Where("id", quizAttempt.Id).With("QuizPackage").First(&UserQuizAttempModel)
	if quizAttempData != nil {
		return &UserQuizAttempModel, nil
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
			UserQuizAttemptId: quizAttempt.Id,
			SelectedOptionId:  answer_map[question.Id],
			IsCorrect:         correct,
			PointsEarned:      point,
		})
	}
	err = facades.Orm().Query().Model(&models.UserQuizAnswer{}).Create(&userAnswers)
	if err != nil {
		return nil, err
	}

	return &UserQuizAttempModel, nil
}

func (s *PackageService) GetUserResults(userId string) ([]dtos.MyQuizResult, error) {
	var userResults []models.UserQuizAttempt

	err := facades.Orm().Query().
		With("QuizPackage").
		Where("user_id", userId).
		Order("created_at DESC").
		Find(&userResults)

	if err != nil {
		return nil, err
	}

	if err != nil {
		return nil, err
	}

	// map untuk grouping per package
	stats := make(map[string]*dtos.MyQuizResult)

	for _, r := range userResults {
		pkgId := r.QuizPackageId
		highestScore := 0
		questions, err := s.GetQuestionsByPackageId(pkgId)
		if err != nil {
			return nil, err
		}

		for _, question := range *questions {
			highestScore += int(question.Point)
		}
		// init kalau belum ada
		if stats[pkgId] == nil {
			stats[pkgId] = &dtos.MyQuizResult{
				QuizPackageId: pkgId,
				HighestScore:  float64(highestScore),
			}
		}

		stat := stats[pkgId]

		// akumulasi
		stat.AvgScore += r.Percentage
		stat.TotalAttempts++

		// best score
		if r.TotalPoints > stat.BestScore {
			stat.BestScore = r.TotalPoints
		}

		// completed
		if r.Status == "passed" {
			stat.Passed++
		}
	}

	// hitung average + convert ke slice
	var result []dtos.MyQuizResult

	for _, stat := range stats {
		if stat.TotalAttempts > 0 {
			stat.AvgScore = math.Round(stat.AvgScore / float64(stat.TotalAttempts))
		}

		result = append(result, *stat)
	}

	return result, nil
}

func (s *PackageService) HasMaxAttempts(userId string, packageId string) (bool, error) {
	pkg, err := s.GetPackageById(packageId, nil)
	if err != nil {
		return false, err
	}
	if pkg == nil {
		return false, nil
	}
	var count int64
	count, err = facades.Orm().Query().Model(&models.UserQuizAttempt{}).Where("user_id", userId).Where("quiz_package_id", packageId).Count()
	if err != nil {
		return false, err
	}
	return count >= int64(pkg.MaxAttempts), nil
}

func (s *PackageService) GetGlobalRankings(limit int) (*[]dtos.Ranking, error) {
	rankings := []dtos.Ranking{}
	err := facades.Orm().Query().
		Table("rankings").
		Join("JOIN users ON users.id = rankings.user_id").
		Select(`
        rankings.user_id,
        users.username,
				users.name as name,
        users.avatar_url as user_avatar,
        SUM(rankings.total_points) as total_points,
				RANK() OVER (ORDER BY SUM(rankings.total_points) DESC) as rank
    `).
		Group("rankings.user_id, users.username, users.avatar_url, users.name").
		Order("total_points DESC").
		Limit(limit).
		Find(&rankings)
	if err != nil {
		return nil, err
	}
	return &rankings, nil
}

func (s *PackageService) GetMyRank(userId string) (*dtos.Ranking, error) {
	ranking := dtos.Ranking{}
	err := facades.Orm().Query().Raw(`
    SELECT *
    FROM (
        SELECT 
            rankings.user_id,
            users.username,
            users.avatar_url,
            SUM(rankings.total_points) as total_points,
            RANK() OVER (ORDER BY SUM(rankings.total_points) DESC) as rank
        FROM rankings
        JOIN users ON users.id = rankings.user_id
        GROUP BY rankings.user_id, users.username, users.avatar_url
    ) ranked
    WHERE ranked.user_id = ?
`, userId).Scan(&ranking)
	if err != nil {
		return nil, err
	}
	return &ranking, nil
}
