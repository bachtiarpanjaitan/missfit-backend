package bootstrap

import (
	"github.com/goravel/framework/contracts/database/seeder"

	"missfit/database/seeders"
)

func Seeders() []seeder.Seeder {
	return []seeder.Seeder{
		&seeders.UserSeeder{},
	}
}
