package bootstrap

import (
	"github.com/goravel/framework/contracts/database/schema"

	"missfit/database/migrations"
)

func Migrations() []schema.Migration {
	return []schema.Migration{
		&migrations.M20210101000001CreateJobsTable{},
		&migrations.M20260408060651User{},
		&migrations.M20260408062548UserBadge{},
		&migrations.M20260408063501QuizPackage{},
		&migrations.M20260408063502UserPurchasedPackage{},
	}
}
