package services

import (
	"errors"
	"math"
	"testing"
	"time"

	dbcontract "github.com/goravel/framework/contracts/database/db"
	ormcontract "github.com/goravel/framework/contracts/database/orm"
	frameworkfoundation "github.com/goravel/framework/foundation"
	mockdb "github.com/goravel/framework/mocks/database/db"
	mockorm "github.com/goravel/framework/mocks/database/orm"
	mockfoundation "github.com/goravel/framework/mocks/foundation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"missfit/app/dtos"
	"missfit/app/models"
)

func installMockApp(t *testing.T) *mockfoundation.Application {
	t.Helper()

	app := mockfoundation.NewApplication(t)
	original := frameworkfoundation.App
	frameworkfoundation.App = app
	t.Cleanup(func() {
		frameworkfoundation.App = original
	})

	return app
}

func expectOrmQueries(t *testing.T, app *mockfoundation.Application, queries ...ormcontract.Query) *mockorm.Orm {
	t.Helper()

	orm := mockorm.NewOrm(t)
	app.On("MakeOrm").Return(orm)

	index := 0
	orm.On("Query").Return(func() ormcontract.Query {
		if index >= len(queries) {
			t.Fatalf("unexpected facades.Orm().Query() call")
			return nil
		}

		query := queries[index]
		index++
		return query
	})

	t.Cleanup(func() {
		assert.Equal(t, len(queries), index, "not all expected ORM queries were used")
	})

	return orm
}

func fillQuizPackageOnFirst(query *mockorm.Query, quizPackage models.QuizPackage) {
	query.On("First", mock.Anything).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.QuizPackage)
		*dest = quizPackage
	}).Return(nil).Once()
}

func fillQuestionsOnFind(query *mockorm.Query, questions []models.QuizQuestion) {
	query.On("Find", mock.Anything).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*[]models.QuizQuestion)
		*dest = questions
	}).Return(nil).Once()
}

func TestCalculateQuizAttemptScore(t *testing.T) {
	startedAt := time.Date(2026, 5, 16, 10, 0, 0, 0, time.UTC)
	completedAt := startedAt.Add(90 * time.Second)
	pkg := &models.QuizPackage{PassingScore: 8}
	questions := []models.QuizQuestion{
		{
			Base:  models.Base{Id: "q-1"},
			Point: 5,
			Options: []models.QuizOption{
				{Base: models.Base{Id: "o-1"}, IsCorrect: true},
				{Base: models.Base{Id: "o-2"}, IsCorrect: false},
			},
		},
		{
			Base:  models.Base{Id: "q-2"},
			Point: 5,
			Options: []models.QuizOption{
				{Base: models.Base{Id: "o-3"}, IsCorrect: true},
				{Base: models.Base{Id: "o-4"}, IsCorrect: false},
			},
		},
		{
			Base:  models.Base{Id: "q-3"},
			Point: 5,
			Options: []models.QuizOption{
				{Base: models.Base{Id: "o-5"}, IsCorrect: true},
				{Base: models.Base{Id: "o-6"}, IsCorrect: false},
			},
		},
	}

	score := calculateQuizAttemptScore(pkg, questions, []dtos.QuizResultAnswer{
		{QuestionId: "q-1", AnswerId: "o-1"},
		{QuestionId: "q-2", AnswerId: "o-4"},
		{QuestionId: "q-3", AnswerId: "skipped"},
	}, startedAt, completedAt)

	assert.Equal(t, float64(5), score.TotalPoint)
	assert.Equal(t, float64(33.33333333333333), score.Percentage)
	assert.False(t, score.IsPassed)
	assert.Equal(t, "failed", score.Status)
	assert.Equal(t, int64(90), score.TimeTakenSeconds)
	assert.Equal(t, 1, score.CorrectAnswers)
	assert.Equal(t, 1, score.WrongAnswers)
	assert.Equal(t, 1, score.SkipAnswers)
}

func TestCalculateQuizAttemptScoreWithNoQuestions(t *testing.T) {
	score := calculateQuizAttemptScore(&models.QuizPackage{PassingScore: 1}, nil, nil, time.Now(), time.Now())

	assert.Equal(t, float64(0), score.TotalPoint)
	assert.Equal(t, float64(0), score.Percentage)
	assert.False(t, score.IsPassed)
	assert.Equal(t, "failed", score.Status)
	assert.False(t, math.IsNaN(score.Percentage))
}

func TestPackageServiceMyProgress(t *testing.T) {
	service := &PackageService{}

	progress, err := service.MyProgress("user-1")

	require.NoError(t, err)
	require.NotNil(t, progress)
}

func TestPackageServiceGetPackageById(t *testing.T) {
	app := installMockApp(t)
	query := mockorm.NewQuery(t)
	expectOrmQueries(t, app, query)

	expected := models.QuizPackage{
		Base:        models.Base{Id: "pkg-1"},
		Title:       "Paket A",
		IsPublished: true,
	}

	query.On("Where", "id", "pkg-1").Return(query).Once()
	query.On("Where", "is_published", true).Return(query).Once()
	fillQuizPackageOnFirst(query, expected)

	pkg, err := (&PackageService{}).GetPackageById("pkg-1", map[string]any{"is_published": true})

	require.NoError(t, err)
	assert.Equal(t, expected.Id, pkg.Id)
	assert.Equal(t, expected.Title, pkg.Title)
}

func TestPackageServiceGetActivePackage(t *testing.T) {
	app := installMockApp(t)
	query := mockorm.NewQuery(t)
	expectOrmQueries(t, app, query)

	expected := models.QuizPackage{Base: models.Base{Id: "pkg-active"}, IsPublished: true}

	query.On("Where", "is_active", true).Return(query).Once()
	fillQuizPackageOnFirst(query, expected)

	pkg, err := (&PackageService{}).GetActivePackage(true)

	require.NoError(t, err)
	assert.Equal(t, "pkg-active", pkg.Id)
}

func TestPackageServiceGetUserPackages(t *testing.T) {
	app := installMockApp(t)
	query := mockorm.NewQuery(t)
	expectOrmQueries(t, app, query)

	expected := []models.UserPurchasedPackage{
		{Base: models.Base{Id: "purchase-1"}, UserId: "user-1", QuizPackageId: "pkg-1", IsActive: true},
	}

	query.On("Table", "user_purchased_packages").Return(query).Once()
	query.On("Join", "left join quiz_packages on quiz_packages.id = user_purchased_packages.quiz_package_id").Return(query).Once()
	query.On("With", "QuizPackage").Return(query).Once()
	query.On("Where", "user_id", "user-1").Return(query).Once()
	query.On("Where", "is_active", true).Return(query).Once()
	query.On("Where", "quiz_packages.is_published", true).Return(query).Once()
	query.On("Offset", 10).Return(query).Once()
	query.On("Limit", 10).Return(query).Once()
	query.On("Order", "created_at desc").Return(query).Once()
	query.On("Find", mock.Anything).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*[]models.UserPurchasedPackage)
		*dest = expected
	}).Return(nil).Once()

	result, err := (&PackageService{}).GetUserPackages("user-1", dtos.PaginationParams{
		Page:  2,
		Limit: 10,
		Sort:  "created_at",
		Order: "desc",
	})

	require.NoError(t, err)
	require.Len(t, *result, 1)
	assert.Equal(t, "purchase-1", (*result)[0].Id)
}

func TestPackageServiceGetUserPurchasedPackage(t *testing.T) {
	app := installMockApp(t)
	query := mockorm.NewQuery(t)
	expectOrmQueries(t, app, query)

	expected := models.UserPurchasedPackage{
		Base:          models.Base{Id: "purchase-1"},
		UserId:        "user-1",
		QuizPackageId: "pkg-1",
	}

	query.On("Where", "user_id", "user-1").Return(query).Once()
	query.On("Where", "quiz_package_id", "pkg-1").Return(query).Once()
	query.On("First", mock.Anything).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.UserPurchasedPackage)
		*dest = expected
	}).Return(nil).Once()

	result, err := (&PackageService{}).GetUserPurchasedPackage("user-1", "pkg-1")

	require.NoError(t, err)
	assert.Equal(t, "purchase-1", result.Id)
}

func TestPackageServiceGetQuestionsByPackageId(t *testing.T) {
	app := installMockApp(t)
	query := mockorm.NewQuery(t)
	expectOrmQueries(t, app, query)

	expected := []models.QuizQuestion{
		{Base: models.Base{Id: "q-1"}, QuizPackageId: "pkg-1", QuestionOrder: 1},
		{Base: models.Base{Id: "q-2"}, QuizPackageId: "pkg-1", QuestionOrder: 2},
	}

	query.On("With", "Options").Return(query).Once()
	query.On("Where", "quiz_package_id", "pkg-1").Return(query).Once()
	query.On("Order", "question_order ASC").Return(query).Once()
	fillQuestionsOnFind(query, expected)

	result, err := (&PackageService{}).GetQuestionsByPackageId("pkg-1")

	require.NoError(t, err)
	require.Len(t, *result, 2)
	assert.Equal(t, "q-1", (*result)[0].Id)
}

func TestPackageServiceSubmitQuizResult(t *testing.T) {
	app := installMockApp(t)
	pkgQuery := mockorm.NewQuery(t)
	questionsQuery := mockorm.NewQuery(t)
	latestAttemptQuery := mockorm.NewQuery(t)
	rankingQuery := mockorm.NewQuery(t)
	fetchAttemptQuery := mockorm.NewQuery(t)
	expectOrmQueries(t, app, pkgQuery, questionsQuery, latestAttemptQuery, rankingQuery, fetchAttemptQuery)

	dbMock := mockdb.NewDB(t)
	tx := mockdb.NewTx(t)
	app.On("MakeDB").Return(dbMock).Once()
	dbMock.On("BeginTransaction").Return(tx, nil).Once()

	startedAt := time.Date(2026, 5, 16, 10, 0, 0, 0, time.UTC)
	completedAt := startedAt.Add(2 * time.Minute)
	pkg := models.QuizPackage{
		Base:         models.Base{Id: "pkg-1"},
		PassingScore: 6,
		MaxAttempts:  3,
	}
	questions := []models.QuizQuestion{
		{
			Base:  models.Base{Id: "q-1"},
			Point: 5,
			Options: []models.QuizOption{
				{Base: models.Base{Id: "o-1"}, IsCorrect: true},
				{Base: models.Base{Id: "o-2"}, IsCorrect: false},
			},
		},
		{
			Base:  models.Base{Id: "q-2"},
			Point: 5,
			Options: []models.QuizOption{
				{Base: models.Base{Id: "o-3"}, IsCorrect: true},
				{Base: models.Base{Id: "o-4"}, IsCorrect: false},
			},
		},
	}

	pkgQuery.On("Where", "id", "pkg-1").Return(pkgQuery).Once()
	fillQuizPackageOnFirst(pkgQuery, pkg)

	questionsQuery.On("With", "Options").Return(questionsQuery).Once()
	questionsQuery.On("Where", "quiz_package_id", "pkg-1").Return(questionsQuery).Once()
	questionsQuery.On("Order", "question_order ASC").Return(questionsQuery).Once()
	fillQuestionsOnFind(questionsQuery, questions)

	latestAttemptQuery.On("Where", "user_id = ? AND quiz_package_id = ?", "user-1", "pkg-1").Return(latestAttemptQuery).Once()
	latestAttemptQuery.On("Order", "created_at DESC").Return(latestAttemptQuery).Once()
	latestAttemptQuery.On("First", mock.Anything).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.UserQuizAttempt)
		*dest = models.UserQuizAttempt{
			Base:        models.Base{Id: "attempt-old"},
			TotalPoints: 2,
		}
	}).Return(nil).Once()

	attemptTable := mockdb.NewQuery(t)
	tx.On("Table", "user_quiz_attempts").Return(attemptTable).Once()
	attemptTable.On("Insert", mock.MatchedBy(func(value any) bool {
		attempt, ok := value.(*models.UserQuizAttempt)
		if !ok {
			return false
		}

		return attempt.UserId == "user-1" &&
			attempt.QuizPackageId == "pkg-1" &&
			attempt.TotalPoints == 5 &&
			attempt.Percentage == 50 &&
			attempt.Status == "failed" &&
			attempt.CorrectAnswers == 1 &&
			attempt.WrongAnswers == 0 &&
			attempt.SkipAnswers == 1 &&
			attempt.TimeTakenSeconds == 120 &&
			attempt.IsPassed != nil &&
			!*attempt.IsPassed
	})).Return(nil, nil).Once()

	rankingQuery.On("Where", "user_id = ? AND quiz_package_id = ?", "user-1", "pkg-1").Return(rankingQuery).Once()
	rankingQuery.On("First", mock.Anything).Return(errors.New("not found")).Once()

	rankingTable := mockdb.NewQuery(t)
	tx.On("Table", "rankings").Return(rankingTable).Once()
	rankingTable.On("Insert", mock.MatchedBy(func(value any) bool {
		ranking, ok := value.(*models.Ranking)
		return ok && ranking.UserId == "user-1" && ranking.QuizPackageId == "pkg-1" && ranking.TotalPoints == 5
	})).Return(nil, nil).Once()

	userTable := mockdb.NewQuery(t)
	tx.On("Table", "users").Return(userTable).Once()
	userTable.On("Where", "id = ?", "user-1").Return(userTable).Once()
	userTable.On("Update", mock.MatchedBy(func(value any) bool {
		data, ok := value.(map[string]any)
		return ok &&
			data["total_points"] == float64(13) &&
			data["total_quizzes_completed"] == 3
	})).Return(nil, nil).Once()

	answerTable1 := mockdb.NewQuery(t)
	answerTable2 := mockdb.NewQuery(t)
	answerTables := []dbcontract.Query{answerTable1, answerTable2}
	answerIndex := 0
	tx.On("Table", "user_quiz_answers").Return(func(name string) dbcontract.Query {
		if answerIndex >= len(answerTables) {
			t.Fatalf("unexpected user_quiz_answers insert")
			return nil
		}
		table := answerTables[answerIndex]
		answerIndex++
		return table
	})
	t.Cleanup(func() {
		assert.Equal(t, len(answerTables), answerIndex)
	})

	answerTable1.On("Insert", mock.MatchedBy(func(value any) bool {
		answer, ok := value.(*models.UserQuizAnswer)
		return ok && answer.SelectedOptionId == "o-1" && answer.IsCorrect && answer.PointsEarned == 5
	})).Return(nil, nil).Once()
	answerTable2.On("Insert", mock.MatchedBy(func(value any) bool {
		answer, ok := value.(*models.UserQuizAnswer)
		return ok && answer.SelectedOptionId == "skipped" && !answer.IsCorrect && answer.PointsEarned == 0
	})).Return(nil, nil).Once()

	tx.On("Commit").Return(nil).Once()

	fetchAttemptQuery.On("Where", "id", mock.Anything).Return(fetchAttemptQuery).Once()
	fetchAttemptQuery.On("With", "QuizPackage").Return(fetchAttemptQuery).Once()
	fetchAttemptQuery.On("First", mock.Anything).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*models.UserQuizAttempt)
		*dest = models.UserQuizAttempt{
			Base:          models.Base{Id: "attempt-new"},
			UserId:        "user-1",
			QuizPackageId: "pkg-1",
			TotalPoints:   5,
			Status:        "failed",
		}
	}).Return(nil).Once()

	user := &models.User{
		Base:                  models.Base{Id: "user-1"},
		TotalPoints:           10,
		TotalQuizzesCompleted: 3,
	}
	result, err := (&PackageService{}).SubmitQuizResult(dtos.QuizResult{
		UserId:      "user-1",
		PackageId:   "pkg-1",
		Score:       50,
		StartedAt:   startedAt,
		CompletedAt: completedAt,
		Answers: []dtos.QuizResultAnswer{
			{QuestionId: "q-1", AnswerId: "o-1"},
			{QuestionId: "q-2", AnswerId: "skipped"},
		},
	}, user)

	require.NoError(t, err)
	assert.Equal(t, "attempt-new", result.Id)
	assert.Equal(t, float64(13), user.TotalPoints)
	assert.Equal(t, 3, user.TotalQuizzesCompleted)
}

func TestPackageServiceGetUserResults(t *testing.T) {
	app := installMockApp(t)
	resultsQuery := mockorm.NewQuery(t)
	questionsAQuery1 := mockorm.NewQuery(t)
	questionsAQuery2 := mockorm.NewQuery(t)
	questionsBQuery := mockorm.NewQuery(t)
	expectOrmQueries(t, app, resultsQuery, questionsAQuery1, questionsAQuery2, questionsBQuery)

	attempts := []models.UserQuizAttempt{
		{QuizPackageId: "pkg-a", Percentage: 80, TotalPoints: 8, Status: "passed"},
		{QuizPackageId: "pkg-a", Percentage: 60, TotalPoints: 6, Status: "failed"},
		{QuizPackageId: "pkg-b", Percentage: 100, TotalPoints: 4, Status: "passed"},
	}

	resultsQuery.On("With", "QuizPackage").Return(resultsQuery).Once()
	resultsQuery.On("Where", "user_id", "user-1").Return(resultsQuery).Once()
	resultsQuery.On("Order", "created_at DESC").Return(resultsQuery).Once()
	resultsQuery.On("Find", mock.Anything).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*[]models.UserQuizAttempt)
		*dest = attempts
	}).Return(nil).Once()

	questionsA := []models.QuizQuestion{{Point: 5}, {Point: 5}}
	questionsB := []models.QuizQuestion{{Point: 4}}
	for _, query := range []*mockorm.Query{questionsAQuery1, questionsAQuery2} {
		query.On("With", "Options").Return(query).Once()
		query.On("Where", "quiz_package_id", "pkg-a").Return(query).Once()
		query.On("Order", "question_order ASC").Return(query).Once()
		fillQuestionsOnFind(query, questionsA)
	}
	questionsBQuery.On("With", "Options").Return(questionsBQuery).Once()
	questionsBQuery.On("Where", "quiz_package_id", "pkg-b").Return(questionsBQuery).Once()
	questionsBQuery.On("Order", "question_order ASC").Return(questionsBQuery).Once()
	fillQuestionsOnFind(questionsBQuery, questionsB)

	results, err := (&PackageService{}).GetUserResults("user-1")

	require.NoError(t, err)
	require.Len(t, results, 2)

	byPackage := map[string]dtos.MyQuizResult{}
	for _, result := range results {
		byPackage[result.QuizPackageId] = result
	}

	assert.Equal(t, float64(70), byPackage["pkg-a"].AvgScore)
	assert.Equal(t, float64(8), byPackage["pkg-a"].BestScore)
	assert.Equal(t, float64(10), byPackage["pkg-a"].HighestScore)
	assert.Equal(t, 1, byPackage["pkg-a"].Passed)
	assert.Equal(t, 2, byPackage["pkg-a"].TotalAttempts)

	assert.Equal(t, float64(100), byPackage["pkg-b"].AvgScore)
	assert.Equal(t, float64(4), byPackage["pkg-b"].BestScore)
	assert.Equal(t, float64(4), byPackage["pkg-b"].HighestScore)
	assert.Equal(t, 1, byPackage["pkg-b"].Passed)
	assert.Equal(t, 1, byPackage["pkg-b"].TotalAttempts)
}

func TestPackageServiceHasMaxAttempts(t *testing.T) {
	app := installMockApp(t)
	pkgQuery := mockorm.NewQuery(t)
	countQuery := mockorm.NewQuery(t)
	expectOrmQueries(t, app, pkgQuery, countQuery)

	pkgQuery.On("Where", "id", "pkg-1").Return(pkgQuery).Once()
	fillQuizPackageOnFirst(pkgQuery, models.QuizPackage{
		Base:        models.Base{Id: "pkg-1"},
		MaxAttempts: 2,
	})

	countQuery.On("Model", mock.Anything).Return(countQuery).Once()
	countQuery.On("Where", "user_id", "user-1").Return(countQuery).Once()
	countQuery.On("Where", "quiz_package_id", "pkg-1").Return(countQuery).Once()
	countQuery.On("Count").Return(int64(2), nil).Once()

	hasMaxAttempts, err := (&PackageService{}).HasMaxAttempts("user-1", "pkg-1")

	require.NoError(t, err)
	assert.True(t, hasMaxAttempts)
}

func TestPackageServiceGetGlobalRankings(t *testing.T) {
	app := installMockApp(t)
	query := mockorm.NewQuery(t)
	expectOrmQueries(t, app, query)

	expected := []dtos.Ranking{{UserId: "user-1", Username: "alice", TotalPoints: 100, Rank: 1}}

	query.On("Table", "rankings").Return(query).Once()
	query.On("Join", "JOIN users ON users.id = rankings.user_id").Return(query).Once()
	query.On("Select", mock.Anything).Return(query).Once()
	query.On("Group", "rankings.user_id, users.username, users.avatar_url, users.name").Return(query).Once()
	query.On("Order", "total_points DESC").Return(query).Once()
	query.On("Limit", 10).Return(query).Once()
	query.On("Find", mock.Anything).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*[]dtos.Ranking)
		*dest = expected
	}).Return(nil).Once()

	result, err := (&PackageService{}).GetGlobalRankings(10)

	require.NoError(t, err)
	require.Len(t, *result, 1)
	assert.Equal(t, "alice", (*result)[0].Username)
}

func TestPackageServiceGetMyRank(t *testing.T) {
	app := installMockApp(t)
	query := mockorm.NewQuery(t)
	expectOrmQueries(t, app, query)

	query.On("Raw", mock.Anything, "user-1").Return(query).Once()
	query.On("Scan", mock.Anything).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*dtos.Ranking)
		*dest = dtos.Ranking{UserId: "user-1", Rank: 2, TotalPoints: 80}
	}).Return(nil).Once()

	result, err := (&PackageService{}).GetMyRank("user-1")

	require.NoError(t, err)
	assert.Equal(t, 2, result.Rank)
	assert.Equal(t, float64(80), result.TotalPoints)
}

func TestPackageServiceGetPackageRank(t *testing.T) {
	app := installMockApp(t)
	query := mockorm.NewQuery(t)
	expectOrmQueries(t, app, query)

	query.On("Raw", mock.Anything, "pkg-1").Return(query).Once()
	query.On("Scan", mock.Anything).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*[]dtos.Ranking)
		*dest = []dtos.Ranking{{UserId: "user-1", Rank: 1, TotalPoints: 90}}
	}).Return(nil).Once()

	result, err := (&PackageService{}).GetPackageRank("pkg-1")

	require.NoError(t, err)
	require.Len(t, result["pkg-1"], 1)
	assert.Equal(t, "user-1", result["pkg-1"][0].UserId)
}

func TestPackageServiceGetPurchaseHistory(t *testing.T) {
	app := installMockApp(t)
	countQuery := mockorm.NewQuery(t)
	itemsQuery := mockorm.NewQuery(t)
	expectOrmQueries(t, app, countQuery, itemsQuery)

	purchasedAt := time.Date(2026, 5, 16, 10, 0, 0, 0, time.UTC)
	expected := []dtos.PurchaseHistoryItem{
		{
			TransactionId: "trx-1",
			OrderId:       "order-1",
			PackageId:     "pkg-1",
			PackageTitle:  "Paket A",
			Amount:        150000,
			Currency:      "IDR",
			Status:        "settlement",
			PurchasedDate: purchasedAt,
		},
	}

	countQuery.On("Raw", mock.Anything, "user-1").Return(countQuery).Once()
	countQuery.On("Scan", mock.Anything).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*int64)
		*dest = 1
	}).Return(nil).Once()

	itemsQuery.On("Raw", mock.Anything, "user-1", 10, 20).Return(itemsQuery).Once()
	itemsQuery.On("Scan", mock.Anything).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*[]dtos.PurchaseHistoryItem)
		*dest = expected
	}).Return(nil).Once()

	items, total, err := (&PackageService{}).GetPurchaseHistory("user-1", dtos.PaginationParams{
		Page:  3,
		Limit: 10,
	})

	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	require.Len(t, items, 1)
	assert.Equal(t, "trx-1", items[0].TransactionId)
}

func TestPackageServiceGetPurchaseHistoryReturnsEmptySlice(t *testing.T) {
	app := installMockApp(t)
	countQuery := mockorm.NewQuery(t)
	itemsQuery := mockorm.NewQuery(t)
	expectOrmQueries(t, app, countQuery, itemsQuery)

	countQuery.On("Raw", mock.Anything, "user-1").Return(countQuery).Once()
	countQuery.On("Scan", mock.Anything).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*int64)
		*dest = 0
	}).Return(nil).Once()

	itemsQuery.On("Raw", mock.Anything, "user-1", 10, 0).Return(itemsQuery).Once()
	itemsQuery.On("Scan", mock.Anything).Return(nil).Once()

	items, total, err := (&PackageService{}).GetPurchaseHistory("user-1", dtos.PaginationParams{
		Page:  1,
		Limit: 10,
	})

	require.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.NotNil(t, items)
	assert.Empty(t, items)
}
