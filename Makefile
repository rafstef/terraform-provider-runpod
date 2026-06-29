default: test

# Unit tests — fast, hermetic (httptest stubs), no external dependencies.
# Acceptance tests are gated behind TF_ACC and are skipped here.
test:
	go test ./...

# Acceptance / integration tests via terraform-plugin-testing (real
# plan/apply/refresh cycles). Requires TF_ACC=1. The *_riab tests additionally
# require RIAB_ACC=1 and a running runpod-in-a-box, with:
#   RUNPOD_BASE_URL    (e.g. http://localhost:8081/v1)
#   RUNPOD_GRAPHQL_URL (e.g. http://localhost:4000/graphql)
#   RUNPOD_API_KEY     (a valid token / test JWT)
# This lane is intentionally "red until fixed": it exercises real behavior and
# will fail on currently-open provider bugs.
testacc:
	TF_ACC=1 go test ./... -v -timeout 30m

.PHONY: default test testacc
