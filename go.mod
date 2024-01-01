module github.com/sfs

go 1.20

replace github.com/sfs => ../sfs

require (
	github.com/alecthomas/assert/v2 v2.3.0
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/go-chi/chi/v5 v5.0.10
	github.com/google/uuid v1.3.0
	github.com/joeshaw/envdecode v0.0.0-20200121155833-099f1fc765bd
	github.com/joho/godotenv v1.5.1
	github.com/mattn/go-sqlite3 v1.14.17
)

require (
	github.com/alecthomas/repr v0.2.0 // indirect
	github.com/hexops/gotextdiff v1.0.3 // indirect
)
