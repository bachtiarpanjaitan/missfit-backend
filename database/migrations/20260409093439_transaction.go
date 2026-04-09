package migrations

import (
	"missfit/app/facades"

	"github.com/goravel/framework/contracts/database/schema"
)

type M20260409093439Transaction struct{}

// Signature The unique signature for the migration.
func (r *M20260409093439Transaction) Signature() string {
	return "20260409093439_transaction"
}

// Up Run the migrations.
func (r *M20260409093439Transaction) Up() error {
	if !facades.Schema().HasTable("rankings") {
		err := facades.Schema().Create("rankings", func(table schema.Blueprint) {
			table.Uuid("id")
			table.Uuid("user_id")
			table.Uuid("quiz_package_id")
			table.Decimal("amount")
			table.String("currency")
			table.String("payment_method")
			table.String("payment_provider")
			table.String("transaction_reference")
			table.String("status")
			table.String("payment_url").Nullable()
			table.Timestamp("paid_at").Nullable()
			table.Jsonb("metadata").Nullable()
			table.String("receipt_url").Nullable()
			table.Text("notes").Nullable()
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
func (r *M20260409093439Transaction) Down() error {
	return nil
}
