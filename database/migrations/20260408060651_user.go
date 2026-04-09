package migrations

import (
	"missfit/app/facades"

	"github.com/goravel/framework/contracts/database/schema"
)

type M20260408060651User struct{}

// Signature The unique signature for the migration.
func (r *M20260408060651User) Signature() string {
	return "20260408060651_user"
}

// Up Run the migrations.
func (r *M20260408060651User) Up() error {
	if !facades.Schema().HasTable("users") {
		err := facades.Schema().Create("users", func(table schema.Blueprint) {
			table.Uuid("id")
			table.String("name")
			table.String("email")
			table.String("password")
			table.String("username")
			table.String("full_name")
			table.String("avatar_url")
			table.String("bio").Nullable()
			table.Date("date_of_birth")
			table.String("gender")
			table.String("phone").Nullable()
			table.Integer("total_points").Default(0)
			table.Integer("total_quizzes_completed").Default(0)
			table.String("auth_provider")
			table.String("auth_provider_id")
			table.Boolean("is_verified").Default(false)
			table.Boolean("is_active").Default(false)
			table.Timestamp("last_login_at").Nullable()
			table.Timestamps()
			table.SoftDeletes()

			table.Primary("id")
			table.Unique("email", "username")
			table.Index("username", "email")
			table.FullText("full_name")
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// Down Reverse the migrations.
func (r *M20260408060651User) Down() error {
	return facades.Schema().DropIfExists("users")
}
