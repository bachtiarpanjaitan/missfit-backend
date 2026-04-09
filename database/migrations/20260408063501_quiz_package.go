package migrations

import (
	"missfit/app/facades"

	"github.com/goravel/framework/contracts/database/schema"
)

type M20260408063501QuizPackage struct{}

// Signature The unique signature for the migration.
func (r *M20260408063501QuizPackage) Signature() string {
	return "20260408063501_quiz_package"
}

// Up Run the migrations.
func (r *M20260408063501QuizPackage) Up() error {
	if !facades.Schema().HasTable("quiz_packages") {
		err := facades.Schema().Create("quiz_packages", func(table schema.Blueprint) {
			table.Uuid("id")
			table.String("title")
			table.Text("description")
			table.String("category")
			table.String("difficulty_level")
			table.Text("thumbnail_url").Nullable()
			table.Boolean("is_free").Default(false)
			table.Decimal("price").Default(0)
			table.String("currency", 3).Default("IDR")
			table.Integer("total_questions").Default(0)
			table.Integer("duration_minutes").Default(0)
			table.Integer("passing_score").Default(0)
			table.Integer("max_attempts").Default(0)
			table.Integer("total_taken").Default(0)
			table.Decimal("average_score").Default(0)
			table.Boolean("is_published").Default(false)
			table.Timestamp("published_at").Nullable()
			table.Uuid("created_by").Nullable()
			table.Timestamp("created_at")
			table.Timestamp("updated_at").Nullable()
			table.Timestamp("deleted_at").Nullable()

			table.Primary("id")
			table.Foreign("created_by").References("id").On("users")
			table.Index("category", "difficulty_level")

		})
		if err != nil {
			return err
		}
	}
	return nil
}

// Down Reverse the migrations.
func (r *M20260408063501QuizPackage) Down() error {
	return facades.Schema().DropIfExists("quiz_packages")
}
