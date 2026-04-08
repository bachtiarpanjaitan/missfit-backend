package migrations

import (
	"missfit/app/facades"

	"github.com/goravel/framework/contracts/database/schema"
)

type M20260408063502UserPurchasedPackage struct{}

// Signature The unique signature for the migration.
func (r *M20260408063502UserPurchasedPackage) Signature() string {
	return "20260408063502_user_purchased_package"
}

// Up Run the migrations.
func (r *M20260408063502UserPurchasedPackage) Up() error {
	if !facades.Schema().HasTable("user_purchased_packages") {
		err := facades.Schema().Create("user_purchased_packages", func(table schema.Blueprint) {
			table.Uuid("id")
			table.Uuid("user_id")
			table.Uuid("quiz_package_id")
			table.String("transaction_id")
			table.Timestamp("purchased_date")
			table.Boolean("is_active")
			table.Timestamp("expired_date")
			table.Timestamp("created_at")
			table.Timestamp("updated_at")
			table.Timestamp("deleted_at")

			table.Primary("id")
			table.Index("user_id", "quiz_package_id")
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
func (r *M20260408063502UserPurchasedPackage) Down() error {
	return facades.Schema().DropIfExists("user_purchased_packages")
}
