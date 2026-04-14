OPENAPI_SPEC := api/openapi/itsreg.swagger.yaml
OAPI_CODEGEN_SERVER_PACKAGE := apiv3
HTTP_API_DIR := internal/api/v3

all: openapi-gen

.PHONY: openapi-gen
openapi-gen: $(HTTP_API_DIR)/openapi_api.gen.go $(HTTP_API_DIR)/openapi_types.gen.go

$(HTTP_API_DIR)/openapi_types.gen.go: $(OPENAPI_SPEC)
	oapi-codegen -generate types -o $@ -package $(OAPI_CODEGEN_SERVER_PACKAGE) $^

$(HTTP_API_DIR)/openapi_api.gen.go: $(OPENAPI_SPEC)
	oapi-codegen -generate chi-server -o $@ -package $(OAPI_CODEGEN_SERVER_PACKAGE) $^
