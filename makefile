MIGRATE_DIR := migrations

.PHONY: openapi_http
openapi_http:
	@./scripts/openapi-http.sh bots internal/api/http http

$(MIGRATE_DIR)/%.sql:
	@migrate create -ext sql -seq -digits 3 -dir $(MIGRATE_DIR)/ $*
