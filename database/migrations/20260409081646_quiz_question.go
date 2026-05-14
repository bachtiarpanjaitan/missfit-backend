package migrations

import (
	"missfit/app/facades"

	"github.com/goravel/framework/contracts/database/schema"
)

type M20260409081646QuizQuestion struct{}

// Signature The unique signature for the migration.
func (r *M20260409081646QuizQuestion) Signature() string {
	return "20260409081646_QuizQuestion"
}

// Up Run the migrations.
func (r *M20260409081646QuizQuestion) Up() error {
	if !facades.Schema().HasTable("quiz_questions") {
		err := facades.Schema().Create("quiz_questions", func(table schema.Blueprint) {
			table.Uuid("id")
			table.Uuid("quiz_package_id")
			table.Text("question_text")
			table.String("question_image_url").Nullable()
			table.Integer("question_order").Default(0)
			table.String("question_type").Nullable()
			table.Integer("number_of_options").Default(0)
			table.Text("explanation").Nullable()
			table.Integer("point").Default(0)
			table.Timestamp("created_at")
			table.Timestamp("updated_at").Nullable()
			table.Timestamp("deleted_at").Nullable()

			table.Primary("id")
			table.Index("quiz_package_id")
			table.Index("question_order")
			// table.Foreign("quiz_package_id").
			// 	References("id").
			// 	On("quiz_packages")
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// Down Reverse the migrations.
func (r *M20260409081646QuizQuestion) Down() error {
	return facades.Schema().DropIfExists("quiz_questions")
}
