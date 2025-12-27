module github.com/diewo77/go-invoices

go 1.25.5

require (
	github.com/diewo77/go-gate v0.0.0
	github.com/diewo77/go-invoices/auth v0.0.0
	github.com/diewo77/go-invoices/httpx v0.0.0-00010101000000-000000000000
	github.com/diewo77/go-invoices/i18n v0.0.0
	github.com/diewo77/go-invoices/view v0.0.0
	github.com/diewo77/go-pdf v0.0.0
	github.com/joho/godotenv v1.5.1
	gorm.io/driver/postgres v1.5.11
	gorm.io/driver/sqlite v1.6.0
	gorm.io/gorm v1.30.0
)

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/pgx/v5 v5.5.5 // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/mattn/go-sqlite3 v1.14.22 // indirect
	github.com/stretchr/testify v1.8.4 // indirect
	golang.org/x/crypto v0.17.0 // indirect
	golang.org/x/sync v0.9.0 // indirect
	golang.org/x/text v0.20.0 // indirect
)

replace (
	github.com/diewo77/go-gate => ../go-gate
	github.com/diewo77/go-invoices/auth => ../auth
	github.com/diewo77/go-invoices/httpx => ../httpx
	github.com/diewo77/go-invoices/i18n => ../i18n
	github.com/diewo77/go-invoices/validation => ../validation
	github.com/diewo77/go-invoices/view => ../view
	github.com/diewo77/go-pdf => ../go-pdf
)
