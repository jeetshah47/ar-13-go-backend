# Testing Guide for Go Backend

This document provides guidance on writing and running tests for the AR-13 Go backend API.

## Test Structure

Tests are located alongside the code they test, using the `_test.go` suffix. For example:
- `internal/handlers/auth.go` → `internal/handlers/auth_test.go`
- `internal/handlers/project.go` → `internal/handlers/project_test.go`

## Test Utilities

Test utilities are located in `internal/test/testutils.go` and provide:
- `SetupTestRouter()` - Creates a test Gin router
- `SetupTestConfig()` - Creates test configuration
- `SetupTestHandler()` - Creates handlers with test config
- `CreateTestRequest()` - Creates HTTP requests for testing
- `CreateAuthenticatedRequest()` - Creates authenticated requests
- `ExecuteRequest()` - Executes requests and returns responses
- `AssertJSONResponse()` - Asserts JSON responses
- `AssertErrorResponse()` - Asserts error responses
- `GenerateTestToken()` - Generates JWT tokens for testing

## Running Tests

### Run All Tests
```bash
go test ./...
```

### Run Tests for a Specific Package
```bash
go test ./internal/handlers
```

### Run Tests with Verbose Output
```bash
go test -v ./...
```

### Run Tests with Coverage
```bash
go test -cover ./...
```

### Generate Coverage Report
```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Run Specific Test
```bash
go test -v -run TestAuthHandler_Login ./internal/handlers
```

## Writing Tests

### Basic Test Structure

```go
package handlers

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestHandler_Method(t *testing.T) {
    // Setup
    gin.SetMode(gin.TestMode)
    router := gin.New()
    handler := NewHandler()
    router.POST("/api/endpoint", handler.Method)

    // Test cases
    t.Run("test case name", func(t *testing.T) {
        req := test.CreateTestRequest("POST", "/api/endpoint", requestBody)
        recorder := test.ExecuteRequest(router, req)

        assert.Equal(t, http.StatusOK, recorder.Code)
        // Additional assertions
    })
}
```

### Testing with Authentication

```go
func TestProtectedEndpoint(t *testing.T) {
    token, err := test.GenerateTestToken("user123", "test@example.com", "Admin")
    require.NoError(t, err)

    req := test.CreateAuthenticatedRequest("GET", "/api/protected", nil, token)
    recorder := test.ExecuteRequest(router, req)

    assert.Equal(t, http.StatusOK, recorder.Code)
}
```

### Testing with Context Values

For endpoints that require context values (like userID from middleware):

```go
handler := func(c *gin.Context) {
    c.Set("userID", "user123")
    actualHandler.Method(c)
}
router.POST("/api/endpoint", handler)
```

## Test Coverage Goals

- **Unit Tests**: Test individual handler methods in isolation
- **Integration Tests**: Test handlers with mocked services
- **API Tests**: Test full request/response cycles

## Mocking Services

For unit tests, you should mock service dependencies. Example:

```go
type MockAuthService struct {
    // Implement service interface
}

func (m *MockAuthService) Login(ctx context.Context, req models.LoginRequest) (*models.LoginResponse, error) {
    // Mock implementation
}
```

## Best Practices

1. **Test One Thing**: Each test should verify one specific behavior
2. **Use Table-Driven Tests**: For multiple similar test cases
3. **Test Edge Cases**: Empty inputs, invalid data, missing fields
4. **Test Error Cases**: Invalid requests, unauthorized access, not found
5. **Clean Setup**: Use `t.Run()` to organize related tests
6. **Assertions**: Use `assert` for non-critical checks, `require` for critical ones
7. **Test Names**: Use descriptive names that explain what is being tested

## Example Test File

See `internal/handlers/auth_test.go` for a complete example of:
- Table-driven tests
- Multiple test cases
- Error handling tests
- Authentication tests

## Continuous Integration

Tests should be run in CI/CD pipelines:

```yaml
# Example GitHub Actions
- name: Run tests
  run: go test -v -coverprofile=coverage.out ./...

- name: Upload coverage
  uses: codecov/codecov-action@v3
  with:
    file: ./coverage.out
```

## Troubleshooting

### Tests Fail with Database Connection Errors
- Use mocks for service layers that require database
- Or use a test database (separate from development)

### Tests Fail with JWT Errors
- Ensure `jwt.InitializeJWT()` is called in test setup
- Use `test.SetupTestConfig()` which includes JWT secret

### Tests Fail with Missing Dependencies
- Run `go mod tidy` to ensure all dependencies are available
- Check that `github.com/stretchr/testify` is installed

## Next Steps

1. Add more comprehensive tests for all handlers
2. Add integration tests with test database
3. Add performance/benchmark tests
4. Set up CI/CD test automation
5. Add test coverage reporting

