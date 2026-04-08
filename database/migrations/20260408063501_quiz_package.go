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
			table.Text("thumbnail_url")
			table.Boolean("is_free")
			table.Decimal("price")
			table.String("currency", 3)
			table.Integer("total_questions")
			table.Integer("duration_minutes")
			table.Integer("passing_score")
			table.Integer("max_attempts")
			table.Integer("total_taken")
			table.Decimal("average_score")
			table.Boolean("is_published")
			table.Timestamp("published_at")
			table.Uuid("created_by")
			table.Timestamp("created_at")
			table.Timestamp("updated_at")
			table.Timestamp("deleted_at")

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
