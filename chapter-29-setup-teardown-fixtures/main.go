package setup_teardown_fixtures

import (
	"fmt"
	"os"
	"sync"
)

type Database struct {
	mu     sync.Mutex
	tables map[string][]map[string]interface{}
}

func NewDatabase() *Database {
	return &Database{
		tables: make(map[string][]map[string]interface{}),
	}
}

func (db *Database) Insert(table string, row map[string]interface{}) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	if _, ok := db.tables[table]; !ok {
		db.tables[table] = make([]map[string]interface{}, 0)
	}
	db.tables[table] = append(db.tables[table], row)
	return nil
}

func (db *Database) Query(table string) ([]map[string]interface{}, error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	rows, ok := db.tables[table]
	if !ok {
		return nil, fmt.Errorf("table %q not found", table)
	}
	out := make([]map[string]interface{}, len(rows))
	copy(out, rows)
	return out, nil
}

func (db *Database) Count(table string) int {
	db.mu.Lock()
	defer db.mu.Unlock()
	return len(db.tables[table])
}

func (db *Database) Truncate(table string) {
	db.mu.Lock()
	defer db.mu.Unlock()
	delete(db.tables, table)
}

func (db *Database) Close() {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.tables = make(map[string][]map[string]interface{})
}

type Fixture struct {
	DB *Database
}

func NewFixture() *Fixture {
	return &Fixture{
		DB: NewDatabase(),
	}
}

func (f *Fixture) Setup() {
	f.DB.Insert("config", map[string]interface{}{"key": "version", "value": "1.0"})
}

func (f *Fixture) Teardown() {
	f.DB.Close()
}

func (f *Fixture) SeedUser(name, email string) {
	f.DB.Insert("users", map[string]interface{}{
		"name":  name,
		"email": email,
	})
}

type TestSuite struct {
	DB *Database
}

func NewTestSuite() *TestSuite {
	return &TestSuite{}
}

func (s *TestSuite) Setup() {
	s.DB = NewDatabase()
	s.DB.Insert("users", map[string]interface{}{"name": "Admin", "role": "admin"})
}

func (s *TestSuite) Teardown() {
	s.DB.Close()
}

func (s *TestSuite) Run(t func()) {
	s.Setup()
	defer s.Teardown()
	t()
}

type ResourceManager struct {
	files []string
}

func NewResourceManager() *ResourceManager {
	return &ResourceManager{}
}

func (rm *ResourceManager) CreateTempFile(content string) (string, error) {
	f, err := os.CreateTemp("", "fixture-*")
	if err != nil {
		return "", fmt.Errorf("creating temp file: %w", err)
	}
	defer f.Close()
	if _, err := f.WriteString(content); err != nil {
		return "", fmt.Errorf("writing temp file: %w", err)
	}
	rm.files = append(rm.files, f.Name())
	return f.Name(), nil
}

func (rm *ResourceManager) Cleanup() {
	for _, f := range rm.files {
		os.Remove(f)
	}
	rm.files = nil
}

func (rm *ResourceManager) HasFiles() bool {
	return len(rm.files) > 0
}

type DBConfig struct {
	Driver string
	DSN    string
}

func MustGetenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

type EnvFixture struct {
	original map[string]string
}

func NewEnvFixture() *EnvFixture {
	return &EnvFixture{
		original: make(map[string]string),
	}
}

func (ef *EnvFixture) Set(key, value string) {
	ef.original[key] = os.Getenv(key)
	os.Setenv(key, value)
}

func (ef *EnvFixture) Restore() {
	for key, val := range ef.original {
		if val == "" {
			os.Unsetenv(key)
		} else {
			os.Setenv(key, val)
		}
	}
}
