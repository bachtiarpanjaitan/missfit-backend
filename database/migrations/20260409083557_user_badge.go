package migrations

import (
	"lumos/app/facades"

	"github.com/goravel/framework/contracts/database/schema"
)

type M20260409083557UserBadge struct{}

// Signature The unique signature for the migration.
func (r *M20260409083557UserBadge) Signature() string {
	return "20260409083557_user_badge"
}

// Up Run the migrations.
func (r *M20260409083557UserBadge) Up() error {
	if !facades.Schema().HasTable("user_badges") {
		err := facades.Schema().Create("user_badges", func(table schema.Blueprint) {
			table.Uuid("id")
			table.Uuid("user_id")
			table.String("badge_name")
			table.String("badge_icon_url").Nullable()
			table.Text("description").Nullable()
			table.Timestamp("earned_at").Nullable()
			table.Timestamps()
			table.SoftDeletes()

			table.Primary("id")
			table.Index("user_id")
			table.Index("badge_name")
			table.Foreign("user_id").References("id").On("users")
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// Down Reverse the migrations.
func (r *M20260409083557UserBadge) Down() error {
	return facades.Schema().DropIfExists("user_badges")
}
