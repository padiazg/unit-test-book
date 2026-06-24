package setup_teardown_fixtures

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDatabase_InsertAndQuery(t *testing.T) {
	db := NewDatabase()
	err := db.Insert("users", map[string]interface{}{"name": "Alice", "age": 30})
	require.NoError(t, err)

	rows, err := db.Query("users")
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.Equal(t, "Alice", rows[0]["name"])
}

func TestDatabase_QueryNonExistent(t *testing.T) {
	db := NewDatabase()
	_, err := db.Query("nonexistent")
	assert.Error(t, err)
}

func TestDatabase_Count(t *testing.T) {
	db := NewDatabase()
	db.Insert("items", map[string]interface{}{"id": 1})
	db.Insert("items", map[string]interface{}{"id": 2})
	assert.Equal(t, 2, db.Count("items"))
}

func TestDatabase_Truncate(t *testing.T) {
	db := NewDatabase()
	db.Insert("temp", map[string]interface{}{"id": 1})
	assert.Equal(t, 1, db.Count("temp"))
	db.Truncate("temp")
	assert.Equal(t, 0, db.Count("temp"))
}

func TestFixture_SetupTeardown(t *testing.T) {
	f := NewFixture()
	f.Setup()

	rows, err := f.DB.Query("config")
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.Equal(t, "1.0", rows[0]["value"])

	f.Teardown()
	assert.Equal(t, 0, f.DB.Count("config"))
}

func TestFixture_SeedUser(t *testing.T) {
	f := NewFixture()
	f.SeedUser("Bob", "bob@test.com")
	f.SeedUser("Charlie", "charlie@test.com")

	rows, err := f.DB.Query("users")
	require.NoError(t, err)
	assert.Len(t, rows, 2)
}

func TestSuite_Run(t *testing.T) {
	s := NewTestSuite()
	called := false
	s.Run(func() {
		called = true
	})
	assert.True(t, called)
}

func TestSuite_Setup(t *testing.T) {
	s := NewTestSuite()
	s.Setup()
	defer s.Teardown()

	rows, err := s.DB.Query("users")
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.Equal(t, "Admin", rows[0]["name"])
}

func TestResourceManager(t *testing.T) {
	t.Run("create and cleanup", func(t *testing.T) {
		rm := NewResourceManager()

		path, err := rm.CreateTempFile("hello world")
		require.NoError(t, err)
		assert.True(t, rm.HasFiles())

		data, err := os.ReadFile(path)
		require.NoError(t, err)
		assert.Equal(t, "hello world", string(data))

		rm.Cleanup()
		assert.False(t, rm.HasFiles())
		_, err = os.Stat(path)
		assert.True(t, os.IsNotExist(err))
	})
}

func TestEnvFixture(t *testing.T) {
	ef := NewEnvFixture()
	ef.Set("MY_APP_MODE", "test")
	assert.Equal(t, "test", os.Getenv("MY_APP_MODE"))
	assert.Equal(t, "test", MustGetenv("MY_APP_MODE", "prod"))

	ef.Restore()
	assert.Equal(t, "", os.Getenv("MY_APP_MODE"))
}

func TestMustGetenv(t *testing.T) {
	os.Setenv("TEST_VAR", "custom")
	defer os.Unsetenv("TEST_VAR")

	assert.Equal(t, "custom", MustGetenv("TEST_VAR", "default"))
	assert.Equal(t, "default", MustGetenv("NONEXISTENT_VAR", "default"))
}

func TestDatabase_ConcurrentAccess(t *testing.T) {
	db := NewDatabase()

	t.Run("parallel inserts", func(t *testing.T) {
		t.Parallel()
		for i := 0; i < 10; i++ {
			db.Insert("parallel_test", map[string]interface{}{"n": i})
		}
	})

	t.Run("parallel reads", func(t *testing.T) {
		t.Parallel()
		rows, err := db.Query("parallel_test")
		require.NoError(t, err)
		_ = rows
	})
}
