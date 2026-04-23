OPENAPI_SPEC := api/v3/itsreg.swagger.yaml
OAPI_CODEGEN_SERVER_PACKAGE := apiv3
HTTP_API_DIR := internal/api/v3
SWAGGER_UI_VERSION := v5.32.4

all: openapi-gen swagger-ui-gen

.PHONY: openapi-gen
openapi-gen: $(HTTP_API_DIR)/openapi_api.gen.go $(HTTP_API_DIR)/openapi_types.gen.go

$(HTTP_API_DIR)/openapi_types.gen.go: $(OPENAPI_SPEC)
	oapi-codegen -generate types -o $@ -package $(OAPI_CODEGEN_SERVER_PACKAGE) $^

$(HTTP_API_DIR)/openapi_api.gen.go: $(OPENAPI_SPEC)
	oapi-codegen -generate chi-server -o $@ -package $(OAPI_CODEGEN_SERVER_PACKAGE) $^

.PHONY: swagger-ui-gen
swagger-ui-gen: $(OPENAPI_SPEC)
	SWAGGER_UI_VERSION=$(SWAGGER_UI_VERSION) ./scripts/generate-swagger-ui.sh $<
