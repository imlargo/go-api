package ffmpeg

import (
	"math"
	"testing"
)

func TestScaleToFillNoBorders(t *testing.T) {
	tests := []struct {
		name        string
		width       int
		height      int
		angleDeg    float64
		maxExpected float64 // Maximum acceptable zoom factor
		description string
	}{
		{
			name:        "Vertical video (9:16) with 5 degree rotation",
			width:       1080,
			height:      1920,
			angleDeg:    5.0,
			maxExpected: 1.20, // Should be reasonable, not excessive
			description: "Vertical videos should have reasonable zoom",
		},
		{
			name:        "Horizontal video (16:9) with 5 degree rotation",
			width:       1920,
			height:      1080,
			angleDeg:    5.0,
			maxExpected: 1.20, // Should be similar to vertical, not excessive
			description: "Horizontal videos should not have excessive zoom",
		},
		{
			name:        "Square video (1:1) with 5 degree rotation",
			width:       1080,
			height:      1080,
			angleDeg:    5.0,
			maxExpected: 1.20,
			description: "Square videos should have symmetric zoom",
		},
		{
			name:        "Vertical video (9:16) with 10 degree rotation",
			width:       1080,
			height:      1920,
			angleDeg:    10.0,
			maxExpected: 1.35,
			description: "Larger rotation angles need more zoom",
		},
		{
			name:        "Horizontal video (16:9) with 10 degree rotation",
			width:       1920,
			height:      1080,
			angleDeg:    10.0,
			maxExpected: 1.35,
			description: "Horizontal videos should scale similarly to vertical for same angle",
		},
		{
			name:        "No rotation",
			width:       1920,
			height:      1080,
			angleDeg:    0.0,
			maxExpected: 1.01, // Should be essentially 1.0
			description: "No rotation should need minimal or no zoom",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scale := scaleToFillNoBorders(tt.width, tt.height, tt.angleDeg)

			// Check that scale is positive
			if scale <= 0 {
				t.Errorf("%s: scale factor must be positive, got %.6f", tt.description, scale)
			}

			// Check that scale is not excessive
			if scale > tt.maxExpected {
				t.Errorf("%s: scale factor %.6f exceeds maximum expected %.6f (angle=%.1f°)",
					tt.description, scale, tt.maxExpected, tt.angleDeg)
			}

			// Check that scale is at least 1.0 for non-zero rotation
			if tt.angleDeg != 0 && scale < 1.0 {
				t.Errorf("%s: scale factor %.6f is less than 1.0 for angle %.1f°",
					tt.description, scale, tt.angleDeg)
			}

			t.Logf("%s: w=%d, h=%d, angle=%.1f° => scale=%.6f (max expected: %.2f)",
				tt.description, tt.width, tt.height, tt.angleDeg, scale, tt.maxExpected)
		})
	}
}

func TestScaleToFillNoBordersSymmetry(t *testing.T) {
	// Test that horizontal and vertical videos with the same angle produce similar zoom factors
	verticalScale := scaleToFillNoBorders(1080, 1920, 5.0)   // 9:16 vertical
	horizontalScale := scaleToFillNoBorders(1920, 1080, 5.0) // 16:9 horizontal

	// The scales should be similar (within 5% of each other)
	diff := math.Abs(verticalScale - horizontalScale)
	maxDiff := 0.05 * math.Max(verticalScale, horizontalScale)

	if diff > maxDiff {
		t.Errorf("Vertical and horizontal videos should have similar zoom factors for same angle. "+
			"Vertical: %.6f, Horizontal: %.6f, Difference: %.6f (max allowed: %.6f)",
			verticalScale, horizontalScale, diff, maxDiff)
	}

	t.Logf("Symmetry test: Vertical scale=%.6f, Horizontal scale=%.6f, Difference=%.6f",
		verticalScale, horizontalScale, diff)
}
