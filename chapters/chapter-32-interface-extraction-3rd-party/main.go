package interface_extraction_3rd_party

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"time"
)

// TransportProvider defines bidirectional byte transport for sensor communication.
// This interface is extracted from a concrete I2C library (github.com/d2r2/go-i2c)
// to make the driver testable without physical hardware.
type TransportProvider interface {
	Write(out []byte) error
	Read(in []byte, full bool) (int, error)
}

// Before: the old SPS30 struct embedded *i2c.I2C directly from
// github.com/d2r2/go-i2c — no interface, impossible to unit test.
//
//	type SPS30 struct {
//		i2c *i2c.I2C
//	}
//
// After: TransportProvider interface extracted, dependency injected via constructor.

// AirQualityReading holds particulate matter data from the sensor.
type AirQualityReading struct {
	Timestamp           time.Time
	MassPM1             float32
	MassPM25            float32
	MassPM4             float32
	MassPM10            float32
	NumberPM05          float32
	NumberPM1           float32
	NumberPM25          float32
	NumberPM4           float32
	NumberPM10          float32
	TypicalParticleSize float32
}

// SPS30 drives a Sensirion SPS30 particulate matter sensor over I2C.
type SPS30 struct {
	transport TransportProvider
	id        string
}

// NewSPS30 creates a sensor instance with an injectable transport.
func NewSPS30(transport TransportProvider) (*SPS30, error) {
	if transport == nil {
		return nil, errors.New("transport not provided")
	}
	return &SPS30{transport: transport, id: "sps30"}, nil
}

// StartMeasurement begins the sensor's measurement cycle.
func (s *SPS30) StartMeasurement() error {
	cmd := []byte{0x00, 0x10, 0x03, 0x00, 0xac}
	if err := s.transport.Write(cmd); err != nil {
		return fmt.Errorf("StartMeasurement: %w", err)
	}
	return nil
}

// StopMeasurement stops the measurement cycle.
func (s *SPS30) StopMeasurement() error {
	cmd := []byte{0x01, 0x04}
	if err := s.transport.Write(cmd); err != nil {
		return fmt.Errorf("StopMeasurement: %w", err)
	}
	return nil
}

// IsDataReady checks if a new measurement is available.
func (s *SPS30) IsDataReady() (bool, error) {
	buf := make([]byte, 3)
	if err := s.sendCommand([]byte{0x02, 0x02}, buf); err != nil {
		return false, err
	}
	return buf[1] == 0x01, nil
}

// ReadMeasurement reads a full measurement payload from the sensor.
func (s *SPS30) ReadMeasurement() (*AirQualityReading, error) {
	buf := make([]byte, 60)
	if err := s.sendCommand([]byte{0x03, 0x00}, buf); err != nil {
		return nil, fmt.Errorf("ReadMeasurement: %w", err)
	}

	return &AirQualityReading{
		Timestamp:           time.Now(),
		MassPM1:             bytesToFloat32([]byte{buf[0], buf[1], buf[3], buf[4]}),
		MassPM25:            bytesToFloat32([]byte{buf[6], buf[7], buf[9], buf[10]}),
		MassPM4:             bytesToFloat32([]byte{buf[12], buf[13], buf[15], buf[16]}),
		MassPM10:            bytesToFloat32([]byte{buf[18], buf[19], buf[21], buf[22]}),
		NumberPM05:          bytesToFloat32([]byte{buf[24], buf[25], buf[27], buf[28]}),
		NumberPM1:           bytesToFloat32([]byte{buf[30], buf[31], buf[33], buf[34]}),
		NumberPM25:          bytesToFloat32([]byte{buf[36], buf[37], buf[39], buf[40]}),
		NumberPM4:           bytesToFloat32([]byte{buf[42], buf[43], buf[45], buf[46]}),
		NumberPM10:          bytesToFloat32([]byte{buf[48], buf[49], buf[51], buf[52]}),
		TypicalParticleSize: bytesToFloat32([]byte{buf[54], buf[55], buf[57], buf[58]}),
	}, nil
}

// sendCommand writes a command and reads the response.
func (s *SPS30) sendCommand(addr []byte, in []byte) error {
	if err := s.transport.Write(addr); err != nil {
		return fmt.Errorf("sendCommand write %X: %w", addr, err)
	}
	if _, err := s.transport.Read(in, false); err != nil {
		return fmt.Errorf("sendCommand read %X: %w", addr, err)
	}
	return nil
}

// bytesToFloat32 converts 4 big-endian bytes to a float32 (IEEE 754).
func bytesToFloat32(data []byte) float32 {
	return math.Float32frombits(binary.BigEndian.Uint32(data))
}
