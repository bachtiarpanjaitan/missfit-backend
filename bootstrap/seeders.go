package bootstrap

import (
	"github.com/goravel/framework/contracts/database/seeder"

	"lumos/database/seeders"
)

func Seeders() []seeder.Seeder {
	return []seeder.Seeder{
		&seeders.UserSeeder{},
		&seeders.QuizPackageSeeder{},
		&seeders.QuizQuestionSeeder{},
		&seeders.QuizOptionSeeder{},
	}
}
