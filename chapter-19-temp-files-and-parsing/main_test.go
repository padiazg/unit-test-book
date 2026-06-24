package temp_files_and_parsing

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	t.Run("valid config file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "config.json")
		err := os.WriteFile(path, []byte(`{"host":"example.com","port":9090,"debug":true,"timeout":60}`), 0644)
		require.NoError(t, err)

		cfg, err := LoadConfig(path)
		require.NoError(t, err)
		assert.Equal(t, "example.com", cfg.Host)
		assert.Equal(t, 9090, cfg.Port)
		assert.True(t, cfg.Debug)
		assert.Equal(t, 60, cfg.Timeout)
	})

	t.Run("file not found", func(t *testing.T) {
		_, err := LoadConfig("/nonexistent/path/config.json")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "reading config file")
	})

	t.Run("empty path", func(t *testing.T) {
		_, err := LoadConfig("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "path is empty")
	})

	t.Run("empty file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "empty.json")
		err := os.WriteFile(path, []byte{}, 0644)
		require.NoError(t, err)

		_, err = LoadConfig(path)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "empty")
	})

	t.Run("invalid JSON", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "bad.json")
		err := os.WriteFile(path, []byte(`{invalid}`), 0644)
		require.NoError(t, err)

		_, err = LoadConfig(path)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "parsing config")
	})

	t.Run("partial config uses zero values", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "partial.json")
		err := os.WriteFile(path, []byte(`{"host":"test.local"}`), 0644)
		require.NoError(t, err)

		cfg, err := LoadConfig(path)
		require.NoError(t, err)
		assert.Equal(t, "test.local", cfg.Host)
		assert.Equal(t, 0, cfg.Port) // zero value
		assert.False(t, cfg.Debug)   // zero value
	})
}

func TestLoadConfigWithDefaults(t *testing.T) {
	t.Run("empty path returns defaults", func(t *testing.T) {
		cfg, err := LoadConfigWithDefaults("")
		require.NoError(t, err)
		assert.Equal(t, "localhost", cfg.Host)
		assert.Equal(t, 8080, cfg.Port)
		assert.Equal(t, 30, cfg.Timeout)
	})

	t.Run("missing file returns defaults", func(t *testing.T) {
		cfg, err := LoadConfigWithDefaults("/nonexistent/config.json")
		require.NoError(t, err)
		assert.Equal(t, "localhost", cfg.Host)
	})

	t.Run("file overrides defaults", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "config.json")
		os.WriteFile(path, []byte(`{"port":3000}`), 0644)

		cfg, err := LoadConfigWithDefaults(path)
		require.NoError(t, err)
		assert.Equal(t, "localhost", cfg.Host) // default
		assert.Equal(t, 3000, cfg.Port)        // override
		assert.Equal(t, 30, cfg.Timeout)       // default
	})
}

func TestSaveConfig(t *testing.T) {
	t.Run("save and reload", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "subdir", "config.json")

		cfg := &Config{Host: "saved.host", Port: 3000, Debug: true, Timeout: 45}
		err := SaveConfig(path, cfg)
		require.NoError(t, err)

		loaded, err := LoadConfig(path)
		require.NoError(t, err)
		assert.Equal(t, "saved.host", loaded.Host)
		assert.Equal(t, 3000, loaded.Port)
		assert.True(t, loaded.Debug)
		assert.Equal(t, 45, loaded.Timeout)
	})

	t.Run("nil config", func(t *testing.T) {
		err := SaveConfig("/tmp/test.json", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "nil")
	})
}
