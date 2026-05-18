package migrations

import (
	"lumos/app/facades"

	"github.com/goravel/framework/contracts/database/schema"
)

type M20260409093439Transaction struct{}

// Signature The unique signature for the migration.
func (r *M20260409093439Transaction) Signature() string {
	return "20260409093439_transaction"
}

// Up Run the migrations.
func (r *M20260409093439Transaction) Up() error {
	// BUG FIX: sebelumnya salah cek "rankings", seharusnya "transactions"
	if !facades.Schema().HasTable("transactions") {
		err := facades.Schema().Create("transactions", func(table schema.Blueprint) {
			table.Uuid("id")
			table.Uuid("user_id")
			table.Uuid("quiz_package_id")

			// Referensi order yang kita kirim ke Midtrans (e.g. "ORDER-xxx-xxx")
			table.String("order_id")

			// ID transaksi yang di-generate oleh Midtrans setelah pembayaran selesai
			table.String("midtrans_transaction_id").Nullable()

			table.Decimal("amount")
			table.String("currency").Default("IDR")

			// Metode pembayaran yang dipilih user (gopay, qris, credit_card, dll)
			table.String("payment_method").Nullable()

			// Provider pembayaran (selalu "midtrans" untuk sekarang)
			table.String("payment_provider").Default("midtrans")

			// Status: pending | settlement | capture | deny | cancel | expire | failure
			table.String("status").Default("pending")

			// Snap token dari Midtrans untuk membuka halaman pembayaran
			table.String("snap_token").Nullable()

			// URL redirect ke halaman Midtrans Snap
			table.Text("payment_url").Nullable()

			// Waktu pembayaran dikonfirmasi (terisi saat status = settlement/capture)
			table.Timestamp("paid_at").Nullable()

			// Menyimpan raw response notification dari Midtrans untuk audit
			table.Jsonb("metadata").Nullable()

			table.Timestamps()
			table.SoftDeletes()

			table.Primary("id")
			table.Index("user_id")
			table.Index("quiz_package_id")
			table.Unique("order_id")
			table.Index("order_id")
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
	return facades.Schema().DropIfExists("transactions")
}
