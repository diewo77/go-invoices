module github.com/diewo77/billing-app

go 1.25.5

require (
	github.com/diewo77/billing-app/auth v0.0.0
	github.com/diewo77/billing-app/httpx v0.0.0
	github.com/diewo77/billing-app/i18n v0.0.0
	github.com/diewo77/billing-app/pdf v0.0.0
	github.com/diewo77/billing-app/validation v0.0.0
	github.com/diewo77/billing-app/view v0.0.0
	github.com/golang-migrate/migrate/v4 v4.19.1
	github.com/joho/godotenv v1.5.1
	golang.org/x/crypto v0.46.0
	gorm.io/driver/postgres v1.6.0
	gorm.io/driver/sqlite v1.6.0
	gorm.io/gorm v1.31.1
)

require (
	github.com/boombuler/barcode v1.0.1 // indirect
	github.com/f-amaral/go-async v0.3.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/hhrutter/lzw v1.0.0 // indirect
	github.com/hhrutter/tiff v1.0.1 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.6.0 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/johnfercher/go-tree v1.0.5 // indirect
	github.com/johnfercher/maroto/v2 v2.3.3 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/lib/pq v1.10.9 // indirect
	github.com/mattn/go-runewidth v0.0.15 // indirect
	github.com/mattn/go-sqlite3 v1.14.22 // indirect
	github.com/pdfcpu/pdfcpu v0.6.0 // indirect
	github.com/phpdave11/gofpdf v1.4.3 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/rivo/uniseg v0.4.4 // indirect
	github.com/rogpeppe/go-internal v1.14.1 // indirect
	golang.org/x/image v0.18.0 // indirect
	golang.org/x/sync v0.19.0 // indirect
	golang.org/x/text v0.32.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

replace github.com/diewo77/billing-app/auth => ../auth

replace github.com/diewo77/billing-app/httpx => ../httpx

replace github.com/diewo77/billing-app/i18n => ../i18n

replace github.com/diewo77/billing-app/pdf => ../pdf

replace github.com/diewo77/billing-app/validation => ../validation

replace github.com/diewo77/billing-app/view => ../view
