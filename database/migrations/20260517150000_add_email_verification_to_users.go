package migrations

import (
	"time"

	"missfit/app/facades"

	"github.com/goravel/framework/contracts/database/schema"
)

type M20260517150000AddEmailVerificationToUsers struct{}

// Signature The unique signature for the migration.
func (r *M20260517150000AddEmailVerificationToUsers) Signature() string {
	return "20260517150000_add_email_verification_to_users"
}

// Up Run the migrations.
func (r *M20260517150000AddEmailVerificationToUsers) Up() error {
	if facades.Schema().HasTable("users") {
		err := facades.Schema().Table("users", func(table schema.Blueprint) {
			table.String("email_verification_token").Nullable()
			table.Timestamp("email_verification_token_expires_at").Nullable()
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// Down Reverse the migrations.
func (r *M20260517150000AddEmailVerificationToUsers) Down() error {
	if facades.Schema().HasTable("users") {
		err := facades.Schema().Table("users", func(table schema.Blueprint) {
			table.DropColumn("email_verification_token")
			table.DropColumn("email_verification_token_expires_at")
		})
		if err != nil {
			return err
		}
	}
	return nil
}
