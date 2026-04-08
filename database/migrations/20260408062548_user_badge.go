package migrations

import (
	"missfit/app/facades"

	"github.com/goravel/framework/contracts/database/schema"
)

type M20260408062548UserBadge struct{}

// Signature The unique signature for the migration.
func (r *M20260408062548UserBadge) Signature() string {
	return "20260408062548_user_badge"
}

// Up Run the migrations.
func (r *M20260408062548UserBadge) Up() error {
	if !facades.Schema().HasTable("user_badges") {
		err := facades.Schema().Create("user_badges", func(table schema.Blueprint) {
			table.Uuid("id")
			table.Uuid("user_id")
			table.String("badge_name")
			table.String("badge_icon_url")
			table.Text("description")
			table.Timestamp("earned_at")

			table.Primary("id")
			table.Index("user_id", "badge_name")
			table.Foreign("user_id").References("id").On("users")
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// Down Reverse the migrations.
func (r *M20260408062548UserBadge) Down() error {
	return facades.Schema().DropIfExists("user_badges")
}
