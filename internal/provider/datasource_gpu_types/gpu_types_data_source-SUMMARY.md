## PLAN COMPLETE

**Plan:** gpu_types-REST-migration
**Tasks:** 5/5
**SUMMARY:** internal/provider/datasource_gpu_types/

**Commits:**
- 5abbb5a: feat(datasource-gpu-types): migrate from GraphQL to v2 REST endpoint

**Duration:** ~5 minutes

## Summary of Changes

### Files Modified

1. **internal/provider/datasource_gpu_types/gpu_types_data_source.go**
   - Migrated from GraphQL `gpuTypes` query to v2 REST endpoint `GET /v2/gpu`
   - Updated base URL from `https://rest.runpod.io/v1` to `https://api.runpod.io/v2`
   - Added `net/http` imports for direct REST calls
   - Rewrote `Read()` method to use HTTP GET request with proper headers
   - Implemented v2 response envelope parsing: `{data: {gpus: [...]}, meta: {...}}`

2. **internal/provider/datasource_gpu_types/gpu_types_data_source_test.go**
   - Added comprehensive test coverage for REST operations:
     - `TestGpuTypesRead_PopulatesState`: Tests single GPU type parsing
     - `TestGpuTypesRead_MultipleGpus`: Tests multiple GPU types parsing
     - `TestGpuTypesRead_ApiError`: Tests HTTP error handling (400, 500, etc.)
     - `TestGpuTypesRead_MissingGpusField`: Tests missing field validation
   - Updated existing test to use v2 REST format

### REST Endpoint Details

- **Endpoint**: `GET /v2/gpu`
- **Base URL**: `https://api.runpod.io/v2`
- **Authorization**: `Bearer {API_KEY}` header
- **Response Format**: v2 REST envelope with `data` wrapper

### Field Mapping (GraphQL → v2 REST)

| GraphQL Field | v2 REST Field | Type |
|--------------|---------------|------|
| `id` | `id` | string |
| `displayName` | `displayName` | string |
| `manufacturer` | `manufacturer` | string |
| `cuda_cores` | `cuda_cores` | float64 |
| `memory_in_gb` | `memory_in_gb` | float64 |
| `community_price` | `community_price` | float64 |
| `secure_price` | `secure_price` | float64 |
| `secure_cloud` | `secure_cloud` | bool |

### Test Results

```
=== RUN   TestGpuTypesRead_PopulatesState
--- PASS: TestGpuTypesRead_PopulatesState (0.00s)
=== RUN   TestGpuTypesRead_MultipleGpus
--- PASS: TestGpuTypesRead_MultipleGpus (0.00s)
=== RUN   TestGpuTypesRead_ApiError
--- PASS: TestGpuTypesRead_ApiError (0.00s)
=== RUN   TestGpuTypesRead_MissingGpusField
--- PASS: TestGpuTypesRead_MissingGpusField (0.00s)
PASS
ok      github.com/runpod/terraform-provider-runpod/internal/provider/datasource_gpu_types    0.312s
```

### Verification

- All tests pass ✓
- No breaking changes to schema ✓
- Backward compatible with existing configurations ✓
- REST endpoint properly configured via `RUNPOD_BASE_URL` env var ✓
