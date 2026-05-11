package bootstrap

import (
	"github.com/goravel/framework/contracts/database/schema"

	"missfit/database/migrations"
)

func Migrations() []schema.Migration {
	return []schema.Migration{
		&migrations.M20210101000001CreateJobsTable{},
		&migrations.M20260408060651User{},
		&migrations.M20260408063501QuizPackage{},
		&migrations.M20260408063502UserPurchasedPackage{},
		&migrations.M20260409081646QuizQuestion{},
		&migrations.M20260409081647QuizOption{},
		&migrations.M20260409083111UserQuizAttempts{},
		&migrations.M20260409082702UserQuizAnswer{},
		&migrations.M20260409083557UserBadge{},
		&migrations.M20260409093135Ranking{},
		&migrations.M20260409093439Transaction{},
		&migrations.M20260511062731AddEducationLevelOnPackage{},
	}
}
