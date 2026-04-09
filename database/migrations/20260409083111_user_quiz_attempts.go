package migrations

import (
	"missfit/app/facades"

	"github.com/goravel/framework/contracts/database/schema"
)

type M20260409083111UserQuizAttempts struct{}

// Signature The unique signature for the migration.
func (r *M20260409083111UserQuizAttempts) Signature() string {
	return "20260409083111_user_quiz_attempts"
}

// Up Run the migrations.
func (r *M20260409083111UserQuizAttempts) Up() error {
	if !facades.Schema().HasTable("user_quiz_attempts") {
		err := facades.Schema().Create("user_quiz_attempts", func(table schema.Blueprint) {
			table.Uuid("id")
			table.Uuid("user_id")
			table.Uuid("quiz_package_id")
			table.Timestamp("started_at").Nullable()
			table.Timestamp("completed_at").Nullable()
			table.Integer("score").Nullable()
			table.Integer("total_points").Nullable()
			table.Decimal("percentage").Nullable()
			table.Boolean("is_passed").Nullable()
			table.Integer("time_taken_seconds").Default(0)
			table.String("status", 20)
			table.Timestamps()
			table.SoftDeletes()

			table.Primary("id")
			table.Index("user_id")
			table.Index("quiz_package_id")
			table.Foreign("user_id").References("id").On("users")
			table.Foreign("quiz_package_id").References("id").On("quiz_packages")
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// Down Reverse the migrations.
func (r *M20260409083111UserQuizAttempts) Down() error {
	return facades.Schema().DropIfExists("user_quiz_answers")
}
