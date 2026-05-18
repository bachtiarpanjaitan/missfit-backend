package services

import (
	"lumos/app/dtos"
	"lumos/app/facades"
	"lumos/app/models"
	"lumos/app/utils"
	"math"
	"time"
)

type PackageServiceInterface interface {
	GetPackageById(id string, filters map[string]any) (*models.QuizPackage, error)
	GetActivePackage(isActive bool) (*models.QuizPackage, error)
	GetUserPurchasedPackage(userId string, packageId string) (*models.UserPurchasedPackage, error)
	GetUserPackages(userId string, pagination dtos.PaginationParams) (*[]models.UserPurchasedPackage, error)
	GetQuestionsByPackageId(packageId string) (*[]models.QuizQuestion, error)
	SubmitQuizResult(quizResult dtos.QuizResult, user *models.User) (*models.UserQuizAttempt, error)
	GetUserResults(userId string) ([]dtos.MyQuizResult, error)
	HasMaxAttempts(userId string, packageId string) (bool, error)
	GetGlobalRankings(limit int) (*[]dtos.Ranking, error)
	GetMyRank(packageId string) (*dtos.Ranking, error)
	GetPackageRank(packageId string) (map[string][]dtos.Ranking, error)
	GetPurchaseHistory(userId string, pagination dtos.PaginationParams) ([]dtos.PurchaseHistoryItem, int64, error)
}

type PackageService struct {
}

func NewPackageService() PackageServiceInterface {
	return &PackageService{}
}

type quizAttemptScore struct {
	TotalPoint       float64
	Percentage       float64
	IsPassed         bool
	Status           string
	TimeTakenSeconds int64
	CorrectAnswers   int
	WrongAnswers     int
	SkipAnswers      int
}

func quizAnswerMap(answers []dtos.QuizResultAnswer) map[string]string {
	answerMap := make(map[string]string)
	for _, answer := range answers {
		answerMap[answer.QuestionId] = answer.AnswerId
	}
	return answerMap
}

func calculateQuizAttemptScore(pkg *models.QuizPackage, questions []models.QuizQuestion, answers []dtos.QuizResultAnswer, startedAt, completedAt time.Time) quizAttemptScore {
	answerMap := quizAnswerMap(answers)

	var totalPoint float64
	var totalWeight float64
	var correctAnswers int
	var wrongAnswers int
	var skipAnswers int

	for _, question := range questions {
		totalWeight += float64(question.Point)
		if answerMap[question.Id] == "skipped" {
			skipAnswers++
			continue
		}

		for _, option := range question.Options {
			if option.IsCorrect && answerMap[question.Id] == option.Id {
				totalPoint += float64(question.Point)
				correctAnswers++
			}

			if option.IsCorrect && answerMap[question.Id] != option.Id {
				wrongAnswers++
			}
		}
	}

	percentage := float64(0)
	if totalWeight > 0 {
		percentage = (totalPoint / totalWeight) * 100
	}

	isPassed := totalPoint >= float64(pkg.PassingScore)
	status := "failed"
	if isPassed {
		status = "passed"
	}

	return quizAttemptScore{
		TotalPoint:       totalPoint,
		Percentage:       percentage,
		IsPassed:         isPassed,
		Status:           status,
		TimeTakenSeconds: utils.DiffSeconds(startedAt, completedAt),
		CorrectAnswers:   correctAnswers,
		WrongAnswers:     wrongAnswers,
		SkipAnswers:      skipAnswers,
	}
}

func (s *PackageService) MyProgress(userId string) (*dtos.UserProgress, error) {
	return &dtos.UserProgress{}, nil
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
	err := facades.Orm().Query().
		Table("user_purchased_packages").
		Join("left join quiz_packages on quiz_packages.id = user_purchased_packages.quiz_package_id").
		With("QuizPackage").
		Where("user_id", userId).
		Where("is_active", true).
		Where("quiz_packages.is_published", true).
		Offset(offset).Limit(pagination.Limit).Order(pagination.Sort + " " + pagination.Order).Find(&userPackages)
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

func (s *PackageService) SubmitQuizResult(quizResult dtos.QuizResult, user *models.User) (*models.UserQuizAttempt, error) {
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

	answerMap := quizAnswerMap(quizResult.Answers)
	score := calculateQuizAttemptScore(pkg, *questions, quizResult.Answers, quizResult.StartedAt, quizResult.CompletedAt)
	totalPoint := score.TotalPoint

	//db transaction
	tx, err := facades.DB().BeginTransaction()
	if err != nil {
		return nil, err
	}

	latestAttemptModel := models.UserQuizAttempt{}
	facades.Orm().Query().
		Where("user_id = ? AND quiz_package_id = ?", quizResult.UserId, quizResult.PackageId).
		Order("created_at DESC").
		First(&latestAttemptModel)

	quizAttempt := models.UserQuizAttempt{
		Base: models.Base{
			Id:        utils.GenerateId(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		UserId:           quizResult.UserId,
		QuizPackageId:    quizResult.PackageId,
		StartedAt:        utils.ToDate(quizResult.StartedAt),
		CompletedAt:      utils.ToDate(quizResult.CompletedAt),
		Score:            quizResult.Score,
		TotalPoints:      totalPoint,
		Percentage:       score.Percentage,
		IsPassed:         &score.IsPassed,
		TimeTakenSeconds: score.TimeTakenSeconds,
		Status:           score.Status,
		CorrectAnswers:   score.CorrectAnswers,
		WrongAnswers:     score.WrongAnswers,
		SkipAnswers:      score.SkipAnswers,
	}
	_, err = tx.Table("user_quiz_attempts").Insert(&quizAttempt)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	//save ranking
	ranking := models.Ranking{}
	exist := facades.Orm().Query().Where("user_id = ? AND quiz_package_id = ?", quizResult.UserId, quizResult.PackageId).First(&ranking)
	if exist == nil && ranking.Id != "" {
		ranking.TotalPoints = totalPoint
		ranking.LastUpdated = time.Now()

		// facades.Orm().Query().Where("id = ?", ranking.Id).Save(&ranking)
		_, err = tx.Table("rankings").Where("id = ?", ranking.Id).Update(map[string]any{
			"total_points": ranking.TotalPoints,
			"last_updated": ranking.LastUpdated,
		})
		if err != nil {
			tx.Rollback()
			return nil, err
		}
	} else {
		ranking = models.Ranking{
			Id:            utils.GenerateId(),
			UserId:        quizResult.UserId,
			QuizPackageId: quizResult.PackageId,
			TotalPoints:   totalPoint,
			CreatedAt:     time.Now(),
			LastUpdated:   time.Now(),
		}
		_, err = tx.Table("rankings").Insert(&ranking)
		// err = facades.Orm().Query().Model(&models.Ranking{}).Create(&ranking)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	if latestAttemptModel.Id == "" {
		user.TotalQuizzesCompleted += 1
	}

	//update user profil
	if latestAttemptModel.Id != "" {
		user.TotalPoints = (user.TotalPoints - latestAttemptModel.TotalPoints) + totalPoint
	} else {
		user.TotalPoints += totalPoint
	}

	// err = facades.Orm().Query().Where("id", user.Id).Save(&user)
	_, err = tx.Table("users").Where("id = ?", user.Id).Update(map[string]any{
		"total_points":            user.TotalPoints,
		"total_quizzes_completed": user.TotalQuizzesCompleted,
	})

	if err != nil {
		tx.Rollback()
		return nil, err
	}

	for _, question := range *questions {
		var correct bool = false
		var point float64 = 0

		for _, option := range question.Options {
			if option.IsCorrect && answerMap[question.Id] == option.Id {
				correct = true
				point = float64(question.Point)
			}
		}

		newAnswer := models.UserQuizAnswer{
			Base: models.Base{
				Id:        utils.GenerateId(),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			UserQuizAttemptId: quizAttempt.Id,
			SelectedOptionId:  answerMap[question.Id],
			IsCorrect:         correct,
			PointsEarned:      point,
		}

		// println(utils.ToJson(&newAnswer))
		_, err = tx.Table("user_quiz_answers").Insert(&newAnswer)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	UserQuizAttemptModel := models.UserQuizAttempt{}
	quizAttempData := facades.Orm().Query().Where("id", quizAttempt.Id).With("QuizPackage").First(&UserQuizAttemptModel)
	if quizAttempData != nil {
		return &quizAttempt, nil
	}

	return &UserQuizAttemptModel, nil
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
        LEFT JOIN users ON users.id = rankings.user_id
        GROUP BY rankings.user_id, users.username, users.avatar_url
    ) ranked
    WHERE ranked.user_id = ?
`, userId).Scan(&ranking)
	if err != nil {
		return nil, err
	}
	return &ranking, nil
}

func (s *PackageService) GetPackageRank(packageId string) (map[string][]dtos.Ranking, error) {
	var rankings []dtos.Ranking

	err := facades.Orm().Query().Raw(`
        SELECT 
            r.user_id,
            COALESCE(u.username, '') as username,
						COALESCE(u.name, '') as name,
            COALESCE(u.avatar_url, '') as user_avatar,
            SUM(r.total_points) as total_points,
            RANK() OVER (ORDER BY SUM(r.total_points) DESC) as rank
        FROM rankings r
        LEFT JOIN users u ON u.id = r.user_id
        WHERE r.quiz_package_id = ?
        GROUP BY r.user_id, u.username, u.name, u.avatar_url
    `, packageId).Scan(&rankings)

	if err != nil {
		return nil, err
	}

	result := make(map[string][]dtos.Ranking)
	result[packageId] = rankings

	return result, nil
}

// GetPurchaseHistory mengambil riwayat pembelian paket berbayar milik user.
// Data diambil dari tabel transactions JOIN quiz_packages.
// Return: items, totalCount, error
func (s *PackageService) GetPurchaseHistory(userId string, pagination dtos.PaginationParams) ([]dtos.PurchaseHistoryItem, int64, error) {
	var items []dtos.PurchaseHistoryItem
	offset := (pagination.Page - 1) * pagination.Limit

	// Hitung total data untuk informasi pagination
	var totalCount int64
	countErr := facades.Orm().Query().Raw(`
		SELECT COUNT(*) FROM transactions WHERE user_id = ?
	`, userId).Scan(&totalCount)
	if countErr != nil {
		return nil, 0, countErr
	}

	err := facades.Orm().Query().Raw(`
		SELECT
			t.id as transaction_id,
			t.order_id,
			t.quiz_package_id as package_id,
			COALESCE(qp.title, 'Paket Dihapus') as package_title,
			t.amount,
			t.currency,
			t.payment_method,
			t.status,
			t.created_at as purchased_date,
			t.paid_at
		FROM transactions t
		LEFT JOIN quiz_packages qp ON qp.id = t.quiz_package_id
		WHERE t.user_id = ?
		ORDER BY t.created_at DESC
		LIMIT ? OFFSET ?
	`, userId, pagination.Limit, offset).Scan(&items)

	if err != nil {
		return nil, 0, err
	}

	if items == nil {
		items = []dtos.PurchaseHistoryItem{}
	}

	return items, totalCount, nil
}
