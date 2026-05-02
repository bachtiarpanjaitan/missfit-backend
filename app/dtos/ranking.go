package dtos

type Ranking struct {
	Rank         int
	UserId       string  `gorm:"column:user_id"`
	Username     string  `gorm:"column:username"`
	Name         string  `gorm:"column:name"`
	UserAvatar   string  `gorm:"column:user_avatar"`
	TotalPoints  float64 `gorm:"column:total_points"`
	QuizzesTaken int     `gorm:"column:quizzes_taken"`
}
