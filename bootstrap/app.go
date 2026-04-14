package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"

	"missfit/config"
	"missfit/routes"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithSeeders(Seeders).
		WithMigrations(Migrations).
		WithRouting(func() {
			routes.Web()
			routes.Grpc()
		}).
		WithProviders(Providers).
		WithConfig(config.Boot).
		Create()
}
