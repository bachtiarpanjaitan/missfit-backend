package migrations

import (
	"lumos/app/facades"

	"github.com/goravel/framework/contracts/database/schema"
)

type M20260409081647QuizOption struct{}

// Signature The unique signature for the migration.
func (r *M20260409081647QuizOption) Signature() string {
	return "20260409081647_quiz_option"
}

// Up Run the migrations.
func (r *M20260409081647QuizOption) Up() error {
	if !facades.Schema().HasTable("quiz_options") {
		err := facades.Schema().Create("quiz_options", func(table schema.Blueprint) {
			table.Uuid("id")
			table.Uuid("quiz_question_id")
			table.String("option_text")
			table.String("option_image_url").Nullable()
			table.Integer("option_order").Default(0)
			table.Boolean("is_correct").Default(false)
			table.Timestamp("created_at")
			table.Timestamp("updated_at").Nullable()
			table.Timestamp("deleted_at").Nullable()

			table.Primary("id")
			table.Index("quiz_question_id")
			table.Index("option_order")
			table.Foreign("quiz_question_id").References("id").On("quiz_questions")
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// Down Reverse the migrations.
func (r *M20260409081647QuizOption) Down() error {
	return facades.Schema().DropIfExists("quiz_options")
}
