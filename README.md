# LumaLogBackEnd

Gin + PostgreSQL API for LumaLog.

## Structure

```text
main.go                         # entrypoint and command dispatch
config/config.go                # environment configuration
database/database.go            # PostgreSQL connection and migration runner
database/migrations/001_init.sql # table definitions and schema migration
model/model.go                  # request/response/table structs
router/router.go                # Gin routes and CORS middleware
handler/*.go                    # HTTP handlers by business area
repository/*.go                 # database access helpers
service/*.go                    # business and statistics logic
patch/patch.go                  # development data patch scripts
util/util.go                    # shared formatting and parsing helpers
```

## Local Database

Create the database first:

```sql
CREATE DATABASE lumalogdb2026;
```

Default connection used by the server:

```text
host=localhost
port=5432
user=postgres
password=794859685
database=lumalogdb2026
```

The server creates tables automatically on startup.

## Run

```bash
go run .
```

## Patch Data

Run a development data patch:

```bash
go run . patch yu-hua-reading-246
```

This creates or updates:

```text
login email: demo@lumalog.local
password: 123456
item: 余华阅读
checkins: 246 consecutive daily checkins
```

Available patch names are shown when the patch name is missing or unknown.

Optional environment variables:

```text
PORT=8080
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=794859685
DB_NAME=lumalogdb2026
DATABASE_URL=postgres://postgres:794859685@localhost:5432/lumalogdb2026?sslmode=disable
JWT_SECRET=change-me
```
