module github.com/nimbus

go 1.20

replace github.com/nimbus => ../nimbus

require (
	github.com/alecthomas/assert/v2 v2.3.0
	github.com/go-chi/chi v1.5.4
	github.com/google/uuid v1.3.0
	github.com/joeshaw/envdecode v0.0.0-20200121155833-099f1fc765bd
	github.com/mattn/go-sqlite3 v1.14.17
	github.com/stretchr/testify v1.8.4
	golang.org/x/sys v0.10.0
)

require (
	github.com/alecthomas/repr v0.2.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/hexops/gotextdiff v1.0.3 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	gorm.io/driver/sqlite v1.5.2 // indirect
	gorm.io/gorm v1.25.2 // indirect
)
