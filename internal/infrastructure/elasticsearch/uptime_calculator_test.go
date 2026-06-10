package elasticsearch

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"testing"
	"time"

	esv8 "github.com/elastic/go-elasticsearch/v8"
	"github.com/stretchr/testify/assert"
)

type mockTransport struct {
	roundTripFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.roundTripFunc(req)
}

func TestESUptimeCalculator_CalculateUptime_Success(t *testing.T) {
	callCount := 0
	mockTrans := &mockTransport{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			callCount++
			body := `{"count": 0}`
			if callCount == 1 {
				body = `{"count": 200}`
			} else if callCount == 2 {
				body = `{"count": 190}`
			}
			header := make(http.Header)
			header.Set("X-Elastic-Product", "Elasticsearch")
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewBufferString(body)),
				Header:     header,
			}, nil
		},
	}

	cfg := esv8.Config{
		Transport: mockTrans,
	}
	es, _ := esv8.NewTypedClient(cfg)

	calc := NewESUptimeCalculator(es, "test-index")
	uptime, err := calc.CalculateUptime(context.Background(), time.Now(), time.Now())

	assert.NoError(t, err)
	assert.Equal(t, 95.0, uptime) // 190 / 200 * 100
}

func TestESUptimeCalculator_CalculateUptime_ZeroTotal(t *testing.T) {
	callCount := 0
	mockTrans := &mockTransport{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			callCount++
			// Total count is 0
			header := make(http.Header)
			header.Set("X-Elastic-Product", "Elasticsearch")
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewBufferString(`{"count": 0}`)),
				Header:     header,
			}, nil
		},
	}

	cfg := esv8.Config{
		Transport: mockTrans,
	}
	es, _ := esv8.NewTypedClient(cfg)

	calc := NewESUptimeCalculator(es, "test-index")
	uptime, err := calc.CalculateUptime(context.Background(), time.Now(), time.Now())

	assert.NoError(t, err)
	assert.Equal(t, 0.0, uptime)
}

func TestESUptimeCalculator_CalculateUptime_TotalCountError(t *testing.T) {
	mockTrans := &mockTransport{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			return nil, errors.New("es connection refused")
		},
	}

	cfg := esv8.Config{
		Transport: mockTrans,
	}
	es, _ := esv8.NewTypedClient(cfg)

	calc := NewESUptimeCalculator(es, "test-index")
	uptime, err := calc.CalculateUptime(context.Background(), time.Now(), time.Now())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to count total observations")
	assert.Equal(t, 0.0, uptime)
}

func TestESUptimeCalculator_CalculateUptime_SuccessCountError(t *testing.T) {
	callCount := 0
	mockTrans := &mockTransport{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			callCount++
			if callCount == 1 {
				header := make(http.Header)
				header.Set("X-Elastic-Product", "Elasticsearch")
				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(bytes.NewBufferString(`{"count": 100}`)),
					Header:     header,
				}, nil
			}
			return nil, errors.New("es read timeout")
		},
	}

	cfg := esv8.Config{
		Transport: mockTrans,
	}
	es, _ := esv8.NewTypedClient(cfg)

	calc := NewESUptimeCalculator(es, "test-index")
	uptime, err := calc.CalculateUptime(context.Background(), time.Now(), time.Now())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to count success observations")
	assert.Equal(t, 0.0, uptime)
}

func TestESUptimeCalculator_CalculateUptime_NilClient(t *testing.T) {
	calc := NewESUptimeCalculator(nil, "test-index")
	uptime, err := calc.CalculateUptime(context.Background(), time.Now(), time.Now())

	assert.NoError(t, err)
	assert.Equal(t, 0.0, uptime)
}
