# Chapter 32: Extracting Interfaces from Third-Party Dependencies

## Description

Extract a thin interface around a concrete third-party library to make your code testable. When a struct embeds a concrete type from a vendored library (e.g. `*i2c.I2C`, `*sql.DB`, `*redis.Client`), every test requires real infrastructure. Wrap the library behind a small interface, inject it through the constructor, and swap it with a `testify/mock` implementation in tests.

Real-world example: the [SPS30](https://github.com/padiazg/go-aqi) particulate matter sensor driver. The original version embedded `*i2c.I2C` from `github.com/d2r2/go-i2c` directly — impossible to unit test without a physical I2C bus. The refactored version extracts a `TransportProvider` interface with `Write`/`Read` methods and injects it via `NewSPS30(transport TransportProvider)`.

## Code

**Before — concrete dependency, untestable:**

```go
import i2c "github.com/d2r2/go-i2c"

type SPS30 struct {
	i2c *i2c.I2C  // concrete type, cannot mock
}

func NewSPS30(i2c *i2c.I2C) *SPS30 {
	return &SPS30{i2c: i2c}
}
```

**After — interface extraction, testable:**

```go
type TransportProvider interface {
	Write(out []byte) error
	Read(in []byte, full bool) (int, error)
}

type SPS30 struct {
	transport TransportProvider
}

func NewSPS30(transport TransportProvider) (*SPS30, error) {
	if transport == nil {
		return nil, errors.New("transport not provided")
	}
	return &SPS30{transport: transport}, nil
}

func (s *SPS30) ReadMeasurement() (*AirQualityReading, error) {
	buf := make([]byte, 60)
	if err := s.sendCommand([]byte{0x03, 0x00}, buf); err != nil {
		return nil, fmt.Errorf("ReadMeasurement: %w", err)
	}
	return &AirQualityReading{
		MassPM1:  bytesToFloat32([]byte{buf[0], buf[1], buf[3], buf[4]}),
		MassPM25: bytesToFloat32([]byte{buf[6], buf[7], buf[9], buf[10]}),
		// ... remaining fields
	}, nil
}
```

## Test

```go
type mockTransportProvider struct {
	mock.Mock
}

func (m *mockTransportProvider) Write(out []byte) error {
	args := m.Called(out)
	return args.Error(0)
}

func (m *mockTransportProvider) Read(in []byte, full bool) (int, error) {
	args := m.Called(in, full)
	r0 := args.Get(0).(int)
	return r0, args.Error(1)
}

func TestSPS30_ReadMeasurement(t *testing.T) {
	// 60-byte payload with CRC8 checksum bytes interleaved
	payload := []byte{
		0x40, 0xc2, 0x2d, 0x5f, 0xce, 0xa7, // PM1.0
		0x41, 0x86, 0x20, 0x4f, 0xaf, 0x43, // PM2.5
		// ... full 60 bytes
	}

	m := new(mockTransportProvider)
	m.On("Write", []byte{0x03, 0x00}).Return(nil)
	m.On("Read", mock.MatchedBy(func(b []byte) bool {
		return len(b) == 60
	}), false).
		Run(func(args mock.Arguments) {
			copy(args.Get(0).([]byte), payload)
		}).
		Return(60, nil)

	s, _ := NewSPS30(m)
	r, err := s.ReadMeasurement()

	assert.NoError(t, err)
	assert.InDelta(t, 6.07, r.MassPM1, 0.01)
	m.AssertExpectations(t)
}
```

## Testing Approach

Third-party dependency interface extraction:

1. **Identify the seam** — look for concrete third-party types in your struct fields. `*sql.DB`, `*redis.Client`, `*i2c.I2C`, `*amqp.Connection` are all candidates. Replace with a small interface that exposes only the methods your code calls.

2. **Extract the minimal interface** — define only the methods the driver actually uses. The SPS30 only needs `Write` and `Read` — not the full I2C register API. Smaller interfaces mean simpler mocks.

3. **Wire the real adapters in production** — in `main.go`, pass the real library instance (e.g. `i2c.NewI2C(...)`) into the constructor. The `go-i2c` library implements `TransportProvider` structurally — no adapter wrapper needed if it already has matching methods.

4. **Mock the transport in tests** — the `mockTransportProvider` uses testify's `mock.Mock` to simulate sensor responses. `mock.MatchedBy` verifies buffer sizes match protocol expectations. `Run` callbacks populate the read buffer with realistic sensor data.

5. **Test helpers in isolation** — `bytesToFloat32` is a pure function with no I/O. Test it with a separate table-driven test covering edge cases (zero, negative, boundary values).

6. **Error path coverage** — each transport method has a failure variant: write error during StartMeasurement, write error during sendCommand, read error during IsDataReady or ReadMeasurement. The mock returns errors for each, and the test verifies the error is wrapped with context (the `%w` in `fmt.Errorf`).

7. **Constructor validation** — `NewSPS30` rejects `nil` transport. This catches misconfiguration at construction time rather than at first sensor read — a common mistake in driver code that's hard to debug on hardware.
