package interface_extraction_3rd_party

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// mockTransportProvider implements TransportProvider for testing.
type mockTransportProvider struct {
	mock.Mock
}

func (m *mockTransportProvider) Write(out []byte) error {
	args := m.Called(out)
	return args.Error(0)
}

func (m *mockTransportProvider) Read(in []byte, full bool) (int, error) {
	args := m.Called(in, full)
	return args.Get(0).(int), args.Error(1)
}

// sampleReadPayload is a 60-byte SPS30 measurement response with checksum bytes
// interleaved every 2 bytes. Each pair of data bytes is followed by a CRC8
// checksum byte, then the next pair.
var sampleReadPayload = []byte{
	// Mass Concentration PM1.0 [µg/m³]
	0x40, 0xc2, 0x2d, 0x5f, 0xce, 0xa7,
	// Mass Concentration PM2.5
	0x41, 0x86, 0x20, 0x4f, 0xaf, 0x43,
	// Mass Concentration PM4.0
	0x41, 0xc9, 0x33, 0x6c, 0xb0, 0xdf,
	// Mass Concentration PM10
	0x41, 0xd6, 0x5e, 0xd8, 0xe1, 0x82,
	// Number Concentration PM0.5
	0x41, 0x8c, 0xfb, 0x6a, 0x52, 0x26,
	// Number Concentration PM1.0
	0x42, 0x11, 0xa3, 0x47, 0x09, 0x2e,
	// Number Concentration PM2.5
	0x42, 0x3f, 0x3a, 0x37, 0xa1, 0x50,
	// Number Concentration PM4.0
	0x42, 0x48, 0x55, 0x8c, 0xb8, 0x10,
	// Number Concentration PM10
	0x42, 0x49, 0x64, 0xe4, 0x68, 0x76,
	// Typical Particle Size [µm]
	0x3f, 0xa4, 0x92, 0x05, 0xf2, 0x16,
	// Padding
	0x00,
}

// ---- Constructor tests ----

type checkNewFn func(*testing.T, *SPS30, error)

var checkNew = func(fns ...checkNewFn) []checkNewFn { return fns }

func checkNewError(want string) checkNewFn {
	return func(t *testing.T, _ *SPS30, err error) {
		t.Helper()
		if want == "" {
			assert.NoError(t, err)
			return
		}
		require.Error(t, err)
		assert.Contains(t, err.Error(), want)
	}
}

func TestNewSPS30(t *testing.T) {
	tests := []struct {
		name      string
		transport TransportProvider
		checks    []checkNewFn
	}{
		{
			name:      "nil transport",
			transport: nil,
			checks: checkNew(
				checkNewError("transport not provided"),
			),
		},
		{
			name:      "valid transport",
			transport: new(mockTransportProvider),
			checks: checkNew(
				checkNewError(""),
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := NewSPS30(tt.transport)
			for _, c := range tt.checks {
				c(t, s, err)
			}
		})
	}
}

// ---- ReadMeasurement tests ----

type checkReadFn func(*testing.T, *AirQualityReading, error)

var checkRead = func(fns ...checkReadFn) []checkReadFn { return fns }

func checkReadSuccess(pm1, pm25, pm4, pm10 float32) checkReadFn {
	return func(t *testing.T, r *AirQualityReading, err error) {
		t.Helper()
		require.NoError(t, err)
		require.NotNil(t, r)
		assert.InDelta(t, pm1, r.MassPM1, 0.01, "MassPM1")
		assert.InDelta(t, pm25, r.MassPM25, 0.01, "MassPM25")
		assert.InDelta(t, pm4, r.MassPM4, 0.01, "MassPM4")
		assert.InDelta(t, pm10, r.MassPM10, 0.01, "MassPM10")
	}
}

func checkReadError(want string) checkReadFn {
	return func(t *testing.T, r *AirQualityReading, err error) {
		t.Helper()
		require.Error(t, err)
		assert.Contains(t, err.Error(), want)
		assert.Nil(t, r)
	}
}

func TestSPS30_ReadMeasurement(t *testing.T) {
	tests := []struct {
		name   string
		before func(*mockTransportProvider)
		checks []checkReadFn
	}{
		{
			name: "success",
			before: func(m *mockTransportProvider) {
				m.On("Write", []byte{0x03, 0x00}).Return(nil)
				m.On("Read", mock.MatchedBy(func(b []byte) bool { return len(b) == 60 }), false).
					Run(func(args mock.Arguments) {
						copy(args.Get(0).([]byte), sampleReadPayload)
					}).
					Return(60, nil)
			},
			checks: checkRead(
				checkReadSuccess(6.07, 16.79, 25.18, 26.86),
			),
		},
		{
			name: "write error",
			before: func(m *mockTransportProvider) {
				m.On("Write", []byte{0x03, 0x00}).Return(errors.New("i2c write error"))
			},
			checks: checkRead(
				checkReadError("sendCommand write"),
			),
		},
		{
			name: "read error",
			before: func(m *mockTransportProvider) {
				m.On("Write", []byte{0x03, 0x00}).Return(nil)
				m.On("Read", mock.Anything, false).Return(0, errors.New("i2c read error"))
			},
			checks: checkRead(
				checkReadError("sendCommand read"),
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := new(mockTransportProvider)
			if tt.before != nil {
				tt.before(m)
			}

			s, err := NewSPS30(m)
			require.NoError(t, err)

			r, err := s.ReadMeasurement()
			for _, c := range tt.checks {
				c(t, r, err)
			}
			m.AssertExpectations(t)
		})
	}
}

// ---- IsDataReady tests ----

type checkReadyFn func(*testing.T, bool, error)

var checkReady = func(fns ...checkReadyFn) []checkReadyFn { return fns }

func checkReadyResult(want bool) checkReadyFn {
	return func(t *testing.T, ready bool, err error) {
		t.Helper()
		require.NoError(t, err)
		assert.Equal(t, want, ready)
	}
}

func checkReadyError(want string) checkReadyFn {
	return func(t *testing.T, _ bool, err error) {
		t.Helper()
		require.Error(t, err)
		assert.Contains(t, err.Error(), want)
	}
}

func TestSPS30_IsDataReady(t *testing.T) {
	tests := []struct {
		name   string
		before func(*mockTransportProvider)
		checks []checkReadyFn
	}{
		{
			name: "data ready",
			before: func(m *mockTransportProvider) {
				m.On("Write", []byte{0x02, 0x02}).Return(nil)
				m.On("Read", mock.MatchedBy(func(b []byte) bool { return len(b) == 3 }), false).
					Run(func(args mock.Arguments) {
						buf := args.Get(0).([]byte)
						buf[0] = 0x00
						buf[1] = 0x01
						buf[2] = 0xb0
					}).
					Return(3, nil)
			},
			checks: checkReady(
				checkReadyResult(true),
			),
		},
		{
			name: "not ready",
			before: func(m *mockTransportProvider) {
				m.On("Write", []byte{0x02, 0x02}).Return(nil)
				m.On("Read", mock.MatchedBy(func(b []byte) bool { return len(b) == 3 }), false).
					Run(func(args mock.Arguments) {
						buf := args.Get(0).([]byte)
						buf[1] = 0x00
					}).
					Return(3, nil)
			},
			checks: checkReady(
				checkReadyResult(false),
			),
		},
		{
			name: "read error",
			before: func(m *mockTransportProvider) {
				m.On("Write", []byte{0x02, 0x02}).Return(nil)
				m.On("Read", mock.Anything, false).Return(0, errors.New("i2c error"))
			},
			checks: checkReady(
				checkReadyError("sendCommand read"),
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := new(mockTransportProvider)
			if tt.before != nil {
				tt.before(m)
			}

			s, err := NewSPS30(m)
			require.NoError(t, err)

			ready, err := s.IsDataReady()
			for _, c := range tt.checks {
				c(t, ready, err)
			}
			m.AssertExpectations(t)
		})
	}
}

// ---- StartMeasurement tests ----

func TestSPS30_StartMeasurement(t *testing.T) {
	tests := []struct {
		name   string
		before func(*mockTransportProvider)
		want   string
	}{
		{
			name: "success",
			before: func(m *mockTransportProvider) {
				m.On("Write", []byte{0x00, 0x10, 0x03, 0x00, 0xac}).Return(nil)
			},
			want: "",
		},
		{
			name: "transport error",
			before: func(m *mockTransportProvider) {
				m.On("Write", []byte{0x00, 0x10, 0x03, 0x00, 0xac}).Return(errors.New("bus error"))
			},
			want: "StartMeasurement",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := new(mockTransportProvider)
			if tt.before != nil {
				tt.before(m)
			}

			s, err := NewSPS30(m)
			require.NoError(t, err)

			err = s.StartMeasurement()
			if tt.want == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.want)
			}
			m.AssertExpectations(t)
		})
	}
}

// ---- Helper: bytesToFloat32 tests ----

func Test_bytesToFloat32(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want float32
	}{
		{name: "zero", data: []byte{0x00, 0x00, 0x00, 0x00}, want: 0},
		{name: "one", data: []byte{0x3F, 0x80, 0x00, 0x00}, want: 1},
		{name: "two", data: []byte{0x40, 0x00, 0x00, 0x00}, want: 2},
		{name: "negative one", data: []byte{0xBF, 0x80, 0x00, 0x00}, want: -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := bytesToFloat32(tt.data)
			assert.Equal(t, tt.want, got)
		})
	}
}
