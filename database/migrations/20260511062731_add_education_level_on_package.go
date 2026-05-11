package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/facades"
)

type M20260511062731AddEducationLevelOnPackage struct{}

// Signature The unique signature for the migration.
func (r *M20260511062731AddEducationLevelOnPackage) Signature() string {
	return "20260511062731_add_education_level_on_package"
}

// Up Run the migrations.
func (r *M20260511062731AddEducationLevelOnPackage) Up() error {
	facades.Schema().Table("quiz_packages", func(table schema.Blueprint) {
		table.String("education_level").Nullable()
	})
	return nil
}

// Down Reverse the migrations.
func (r *M20260511062731AddEducationLevelOnPackage) Down() error {
	facades.Schema().Table("quiz_packages", func(table schema.Blueprint) {
		table.DropColumn("education_level")
	})
	return nil
}
