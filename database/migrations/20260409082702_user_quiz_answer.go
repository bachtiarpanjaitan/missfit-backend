package migrations

import (
	"missfit/app/facades"

	"github.com/goravel/framework/contracts/database/schema"
)

type M20260409082702UserQuizAnswer struct{}

// Signature The unique signature for the migration.
func (r *M20260409082702UserQuizAnswer) Signature() string {
	return "20260409082702_user_quiz_answer"
}

// Up Run the migrations.
func (r *M20260409082702UserQuizAnswer) Up() error {
	if !facades.Schema().HasTable("user_quiz_answers") {
		err := facades.Schema().Create("user_quiz_answers", func(table schema.Blueprint) {
			table.Uuid("id")
			table.Uuid("user_quiz_attempt_id")
			table.Uuid("selected_option_id")
			table.Boolean("is_correct").Default(false)
			table.Integer("points_earned").Default(0)
			table.Timestamp("created_at")
			table.Timestamp("updated_at").Nullable()

			table.Primary("id")
			table.Index("user_quiz_attempt_id")
			table.Index("selected_option_id")
			table.Foreign("user_quiz_attempt_id").References("id").On("user_quiz_attempts")
			table.Foreign("selected_option_id").References("id").On("quiz_options")
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// Down Reverse the migrations.
func (r *M20260409082702UserQuizAnswer) Down() error {
	return facades.Schema().DropIfExists("user_quiz_answers")
}
