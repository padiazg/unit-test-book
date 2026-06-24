# Chapter 19: Temp Files and Parsing

## Description

Use `t.TempDir()` to create temporary directories that are automatically cleaned up when the test completes. Combined with `os.WriteFile` / `os.ReadFile`, you can test file I/O, JSON parsing, and config loading without polluting the filesystem or relying on pre-existing fixture files.

## Code

```go
type Config struct {
	Host    string `json:"host"`
	Port    int    `json:"port"`
	Debug   bool   `json:"debug"`
	Timeout int    `json:"timeout"`
}

func LoadConfig(path string) (*Config, error) {
	if path == "" {
		return nil, fmt.Errorf("path is empty")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("config file is empty")
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	return &cfg, nil
}

func SaveConfig(path string, cfg *Config) error {
	// creates parent directories, marshals JSON, writes file
}
```

## Test

```go
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
		assert.Equal(t, 0, cfg.Port)  // zero value
	})
}

func TestLoadConfigWithDefaults(t *testing.T) {
	t.Run("empty path returns defaults", func(t *testing.T) {
		cfg, err := LoadConfigWithDefaults("")
		require.NoError(t, err)
		assert.Equal(t, "localhost", cfg.Host)
		assert.Equal(t, 8080, cfg.Port)
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
		assert.Equal(t, 3000, cfg.Port)         // override
	})
}

func TestSaveConfig(t *testing.T) {
	t.Run("save and reload", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "subdir", "config.json")

		cfg := &Config{Host: "saved.host", Port: 3000}
		err := SaveConfig(path, cfg)
		require.NoError(t, err)

		loaded, err := LoadConfig(path)
		require.NoError(t, err)
		assert.Equal(t, "saved.host", loaded.Host)
		assert.Equal(t, 3000, loaded.Port)
	})
}
```

## Testing Approach

Temp files and parsing:

1. **`t.TempDir()` auto-cleanup** — the directory and all its contents are removed when the test finishes. No `os.RemoveAll`, no defer, no leftover fixtures. Each subtest gets its own directory — no file collisions.
2. **Error path coverage** — empty path, missing file, empty file, and malformed JSON are all tested as distinct cases. The error messages are wrapped at each layer (`reading config file`, `parsing config`) making errors easy to debug.
3. **`filepath.Join` for cross-platform paths** — always join paths with `filepath.Join`, never string concatenation. The tests work on Windows, Linux, and macOS.
4. **Save-and-reload round trip** — `SaveConfig` writes to `t.TempDir()/subdir/config.json` (including parent directory creation), then `LoadConfig` reads it back. One test validates both write and read paths with a single fixture.
