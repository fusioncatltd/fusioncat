export TESTSERVER_URL ?= http://localhost:8080/

# Load .env file if it exists
-include .env
export

run:
	go run main.go

update_docs_run:
	swag init; go run main.go;

test:
	go clean -testcache && go test ./tests  -v

truncate-remote-db:
	@echo "Truncating all tables except schema_migrations..."
	@psql postgresql://$(PG_USER):$(PG_PASSWORD)@$(PG_HOST):$(PG_PORT)/$(PG_DB_NAME)?sslmode=$(PG_SSLMODE) -c "DO \$$\$$ DECLARE r RECORD; BEGIN FOR r IN (SELECT tablename FROM pg_tables WHERE schemaname = 'public' AND tablename != 'schema_migrations') LOOP EXECUTE 'TRUNCATE TABLE ' || quote_ident(r.tablename) || ' CASCADE'; END LOOP; END \$$\$$;"
	@echo "Database truncated successfully!"

clean-test: truncate-remote-db test

