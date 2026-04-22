package analysis_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/better/backend/internal/analysis"
)

func TestCalculateEV(t *testing.T) {
	tests := []struct {
		name      string
		modelProb float64
		odd       float64
		expected  float64
	}{
		{
			name:      "positive EV",
			modelProb: 0.60,
			odd:       2.00,
			expected:  0.20, // (0.60 * 2.00) - 1 = 0.20
		},
		{
			name:      "negative EV",
			modelProb: 0.40,
			odd:       2.00,
			expected:  -0.20, // (0.40 * 2.00) - 1 = -0.20
		},
		{
			name:      "zero EV",
			modelProb: 0.50,
			odd:       2.00,
			expected:  0.00, // (0.50 * 2.00) - 1 = 0.00
		},
		{
			name:      "zero odd",
			modelProb: 0.50,
			odd:       0.00,
			expected:  0.00,
		},
		{
			name:      "high confidence positive EV",
			modelProb: 0.70,
			odd:       1.80,
			expected:  0.26, // (0.70 * 1.80) - 1 = 0.26
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analysis.CalculateEV(tt.modelProb, tt.odd)
			assert.InDelta(t, tt.expected, result, 0.001)
		})
	}
}

func TestCalculateImpliedProbability(t *testing.T) {
	tests := []struct {
		name     string
		odd      float64
		expected float64
	}{
		{
			name:     "even odds",
			odd:      2.00,
			expected: 0.50,
		},
		{
			name:     "strong favorite",
			odd:      1.25,
			expected: 0.80,
		},
		{
			name:     "underdog",
			odd:      4.00,
			expected: 0.25,
		},
		{
			name:     "zero odd",
			odd:      0.00,
			expected: 0.00,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analysis.CalculateImpliedProbability(tt.odd)
			assert.InDelta(t, tt.expected, result, 0.001)
		})
	}
}

func TestAnalyzeValue(t *testing.T) {
	t.Run("value opportunity exists", func(t *testing.T) {
		result := analysis.AnalyzeValue(0.60, 2.00)

		assert.True(t, result.IsValue)
		assert.InDelta(t, 0.60, result.ModelProb, 0.001)
		assert.InDelta(t, 0.50, result.ImpliedProb, 0.001)
		assert.InDelta(t, 0.20, result.EV, 0.001)
		assert.True(t, result.Divergence > 0)
	})

	t.Run("no value opportunity", func(t *testing.T) {
		result := analysis.AnalyzeValue(0.40, 2.00)

		assert.False(t, result.IsValue)
		assert.InDelta(t, -0.20, result.EV, 0.001)
	})

	t.Run("marginal value", func(t *testing.T) {
		result := analysis.AnalyzeValue(0.55, 1.90)

		// EV = (0.55 * 1.90) - 1 = 0.045
		assert.True(t, result.EV > 0)
		assert.True(t, result.IsValue)
	})
}
