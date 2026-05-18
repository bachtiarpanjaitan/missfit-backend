package migrations

import (
	"lumos/app/facades"

	"github.com/goravel/framework/contracts/database/schema"
)

type M20260409093135Ranking struct{}

// Signature The unique signature for the migration.
func (r *M20260409093135Ranking) Signature() string {
	return "20260409093135_ranking"
}

// Up Run the migrations.
func (r *M20260409093135Ranking) Up() error {
	if !facades.Schema().HasTable("rankings") {
		err := facades.Schema().Create("rankings", func(table schema.Blueprint) {
			table.Uuid("id")
			table.Uuid("user_id")
			table.Uuid("quiz_package_id")
			table.Integer("total_points").Default(0)
			table.Timestamp("last_updated").Nullable()
			table.Timestamp("created_at")

			table.Primary("id")
			table.Index("user_id")
			table.Index("quiz_package_id")
			table.Foreign("user_id").References("id").On("users")
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// Down Reverse the migrations.
func (r *M20260409093135Ranking) Down() error {
	return facades.Schema().DropIfExists("rankings")
}
