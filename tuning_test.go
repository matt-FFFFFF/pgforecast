package pgforecast

import "testing"

func TestDefaultDisplayConfig(t *testing.T) {
	tc := DefaultTuningConfig()

	// Wind strength defaults
	tests := []struct {
		name string
		tier WindStrengthTier
		rgb  string
		icon string
	}{
		{"light", tc.Display.WindStrength.Light, "#4fd1c5", "💤"},
		{"moderate", tc.Display.WindStrength.Moderate, "#48bb78", "✅"},
		{"fresh", tc.Display.WindStrength.Fresh, "#ecc94b", "⚠️"},
		{"strong", tc.Display.WindStrength.Strong, "#ed8936", "🟠"},
		{"very_strong", tc.Display.WindStrength.VeryStrong, "#f56565", "🔴"},
	}
	for _, tt := range tests {
		if tt.tier.RGB != tt.rgb {
			t.Errorf("%s: RGB = %q, want %q", tt.name, tt.tier.RGB, tt.rgb)
		}
		if tt.tier.Icon != tt.icon {
			t.Errorf("%s: Icon = %q, want %q", tt.name, tt.tier.Icon, tt.icon)
		}
	}

	// Gradient defaults
	if tc.Display.Gradient.Low.RGB != "#48bb78" {
		t.Errorf("gradient low RGB = %q, want #48bb78", tc.Display.Gradient.Low.RGB)
	}
	if tc.Display.Gradient.Medium.Icon != "⚠️" {
		t.Errorf("gradient medium icon = %q, want ⚠️", tc.Display.Gradient.Medium.Icon)
	}
	if tc.Display.Gradient.High.RGB != "#f56565" {
		t.Errorf("gradient high RGB = %q, want #f56565", tc.Display.Gradient.High.RGB)
	}
}

func TestWindStrengthTierFor(t *testing.T) {
	tc := DefaultTuningConfig()

	tests := []struct {
		speed float64
		label string
	}{
		{3, "Light"},       // below ideal_min (8)
		{7.9, "Light"},     // just below ideal_min
		{8, "Moderate"},    // at ideal_min
		{18, "Moderate"},   // at ideal_max
		{19, "Fresh"},      // above ideal_max, below acceptable_max
		{22, "Fresh"},      // at acceptable_max
		{23, "Strong"},     // above acceptable_max, below dangerous_max
		{25, "Strong"},     // at dangerous_max
		{26, "Very Strong"}, // above dangerous_max
		{50, "Very Strong"},
	}
	for _, tt := range tests {
		tier := tc.WindStrengthTierFor(tt.speed)
		if tier.Label != tt.label {
			t.Errorf("WindStrengthTierFor(%.1f) = %q, want %q", tt.speed, tier.Label, tt.label)
		}
	}
}

func TestGradientIcon(t *testing.T) {
	tc := DefaultTuningConfig()

	if tc.GradientIcon(GradientLow) != "✅" {
		t.Errorf("GradientIcon(Low) = %q, want ✅", tc.GradientIcon(GradientLow))
	}
	if tc.GradientIcon(GradientMedium) != "⚠️" {
		t.Errorf("GradientIcon(Medium) = %q, want ⚠️", tc.GradientIcon(GradientMedium))
	}
	if tc.GradientIcon(GradientHigh) != "🔴" {
		t.Errorf("GradientIcon(High) = %q, want 🔴", tc.GradientIcon(GradientHigh))
	}
}
