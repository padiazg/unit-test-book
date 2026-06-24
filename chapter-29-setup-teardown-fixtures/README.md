# Chapter 29: Setup/Teardown Fixtures

## Description

Use `TestMain` for package-level setup/teardown, struct-based fixtures for per-test state, `t.TempDir()` for temporary files, and `os.Setenv`/`os.Unsetenv` with defer for environment variable mocking. Fixtures isolate test state and guarantee cleanup regardless of test outcome.

Real-world example: `hexago/internal/adapters/secondary/database/category_repository_test.go` — `NewFixture()` struct with Setup/Teardown methods for database transactions.

## Code

```go
type Fixture struct {
	DB *Database
}

func NewFixture() *Fixture {
	return &Fixture{DB: NewDatabase()}
}

func (f *Fixture) Setup() {
	f.DB.Insert("config", map[string]interface{}{"key": "version", "value": "1.0"})
}

func (f *Fixture) Teardown() {
	f.DB.Close()
}

type EnvFixture struct {
	original map[string]string
}

func NewEnvFixture() *EnvFixture {
	return &EnvFixture{original: make(map[string]string)}
}

func (ef *EnvFixture) Set(key, value string) {
	ef.original[key] = os.Getenv(key)
	os.Setenv(key, value)
}

func (ef *EnvFixture) Restore() {
	for key, val := range ef.original {
		if val == "" { os.Unsetenv(key) } else { os.Setenv(key, val) }
	}
}
```

## Test

```go
func TestFixture_SetupTeardown(t *testing.T) {
	f := NewFixture()
	f.Setup()
	rows, err := f.DB.Query("config")
	require.NoError(t, err)
	assert.Equal(t, "1.0", rows[0]["value"])
	f.Teardown()
	assert.Equal(t, 0, f.DB.Count("config"))
}

func TestResourceManager(t *testing.T) {
	t.Run("create and cleanup", func(t *testing.T) {
		rm := NewResourceManager()
		path, err := rm.CreateTempFile("hello world")
		require.NoError(t, err)
		data, _ := os.ReadFile(path)
		assert.Equal(t, "hello world", string(data))
		rm.Cleanup()
		_, err = os.Stat(path)
		assert.True(t, os.IsNotExist(err))
	})
}

func TestEnvFixture(t *testing.T) {
	ef := NewEnvFixture()
	ef.Set("MY_APP_MODE", "test")
	assert.Equal(t, "test", os.Getenv("MY_APP_MODE"))
	ef.Restore()
	assert.Equal(t, "", os.Getenv("MY_APP_MODE"))
}

func TestMustGetenv(t *testing.T) {
	os.Setenv("TEST_VAR", "custom")
	defer os.Unsetenv("TEST_VAR")
	assert.Equal(t, "custom", MustGetenv("TEST_VAR", "default"))
	assert.Equal(t, "default", MustGetenv("NONEXISTENT", "default"))
}

func TestDatabase_InsertAndQuery(t *testing.T) {
	db := NewDatabase()
	db.Insert("users", map[string]interface{}{"name": "Alice"})
	rows, _ := db.Query("users")
	assert.Len(t, rows, 1)
}

func TestDatabase_Truncate(t *testing.T) {
	db := NewDatabase()
	db.Insert("temp", map[string]interface{}{"id": 1})
	db.Truncate("temp")
	assert.Equal(t, 0, db.Count("temp"))
}
```

## Testing Approach

Setup/teardown fixtures:

1. **Fixture struct per test** — `NewFixture()` returns a fresh `Fixture` with its own `Database`. Each test sets up, runs, and tears down. No shared state means no pollution between tests.
2. **`defer` for teardown** — `f.Teardown()` in `defer` guarantees cleanup even if the test panics. `EnvFixture.Restore()` restores environment variables regardless of test outcome.
3. **`TestMain` for expensive resources** — `TestMain` creates a `TestSuite`, runs `Setup()` once, runs all tests via `m.Run()`, then runs `Teardown()`. Use sparingly — one-time setup is for resources that are expensive to create (database connections, test servers).
4. **`os.Setenv` + `defer os.Unsetenv`** — env var changes are global. Always pair `Setenv` with `defer Unsetenv` in the same test. `EnvFixture` provides a reusable wrapper that saves originals and restores them.
