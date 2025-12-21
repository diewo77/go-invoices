module github.com/diewo77/go-invoices/view

go 1.25.5

require (
	github.com/diewo77/go-invoices/auth v0.0.0
	github.com/diewo77/go-invoices/i18n v0.0.0
)

replace (
	github.com/diewo77/go-invoices/auth => ../auth
	github.com/diewo77/go-invoices/i18n => ../i18n
)
