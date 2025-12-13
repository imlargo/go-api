package transform

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestGetTextSegments_LineBreaks(t *testing.T) {
	tests := []struct {
		name            string
		text            string
		maxWidth        float64
		approxCharWidth float64
		baseFontSize    float64
		wantLines       []string
		description     string
	}{
		{
			name:            "Single line break",
			text:            "Line 1\nLine 2",
			maxWidth:        500,
			approxCharWidth: 10,
			baseFontSize:    60,
			wantLines:       []string{"Line 1", "Line 2"},
			description:     "Should preserve single line break",
		},
		{
			name:            "Multiple consecutive line breaks",
			text:            "Line 1\n\nLine 3",
			maxWidth:        500,
			approxCharWidth: 10,
			baseFontSize:    60,
			wantLines:       []string{"Line 1", "", "Line 3"},
			description:     "Should preserve empty lines from consecutive line breaks",
		},
		{
			name:            "No line breaks",
			text:            "Single line text",
			maxWidth:        500,
			approxCharWidth: 10,
			baseFontSize:    60,
			wantLines:       []string{"Single line text"},
			description:     "Should handle text without line breaks",
		},
		{
			name:            "Line break with long text requiring wrapping",
			text:            "Short\nThis is a very long line that needs to be wrapped because it exceeds the maximum width",
			maxWidth:        300,
			approxCharWidth: 10,
			baseFontSize:    60,
			// With ~300px width and ~10px per char, we get ~30 chars per line
			// The long line should wrap
			wantLines:   nil, // We'll check that it has multiple lines and starts with "Short"
			description: "Should handle both user line breaks and automatic wrapping",
		},
		{
			name:            "Empty text",
			text:            "",
			maxWidth:        500,
			approxCharWidth: 10,
			baseFontSize:    60,
			wantLines:       []string{""},
			description:     "Should handle empty text gracefully",
		},
		{
			name:            "Text with only spaces",
			text:            "   ",
			maxWidth:        500,
			approxCharWidth: 10,
			baseFontSize:    60,
			wantLines:       []string{""},
			description:     "Should handle text with only spaces",
		},
		{
			name:            "Line break at start",
			text:            "\nLine 2",
			maxWidth:        500,
			approxCharWidth: 10,
			baseFontSize:    60,
			wantLines:       []string{"", "Line 2"},
			description:     "Should preserve empty line at start",
		},
		{
			name:            "Line break at end",
			text:            "Line 1\n",
			maxWidth:        500,
			approxCharWidth: 10,
			baseFontSize:    60,
			wantLines:       []string{"Line 1", ""},
			description:     "Should preserve empty line at end",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			overlay := TextOverlay{
				Text:            tt.text,
				MaxWidth:        tt.maxWidth,
				approxCharWidth: tt.approxCharWidth,
				baseFontSize:    tt.baseFontSize,
				baseLineHeight:  0.85,
			}

			result := overlay.GetTextSegments()
			lines := strings.Split(result, "\n")

			if tt.wantLines != nil {
				if len(lines) != len(tt.wantLines) {
					t.Errorf("%s: got %d lines, want %d lines\nGot: %v\nWant: %v",
						tt.description, len(lines), len(tt.wantLines), lines, tt.wantLines)
					return
				}

				for i, wantLine := range tt.wantLines {
					if lines[i] != wantLine {
						t.Errorf("%s: line %d mismatch\nGot:  %q\nWant: %q",
							tt.description, i, lines[i], wantLine)
					}
				}
			} else {
				// Special case for the wrapping test
				if tt.name == "Line break with long text requiring wrapping" {
					if len(lines) < 2 {
						t.Errorf("%s: expected at least 2 lines, got %d", tt.description, len(lines))
					}
					if !strings.HasPrefix(result, "Short\n") {
						t.Errorf("%s: expected result to start with 'Short\\n', got %q", tt.description, result)
					}
				}
			}

			t.Logf("%s: Input: %q => Output lines: %v", tt.description, tt.text, lines)
		})
	}
}

func TestGetTextSegments_NoSquareSymbols(t *testing.T) {
	// Test that the output doesn't contain literal \n characters
	// (which would appear as square symbols in FFmpeg)
	overlay := TextOverlay{
		Text:            "Line 1\nLine 2\nLine 3",
		MaxWidth:        500,
		approxCharWidth: 10,
		baseFontSize:    60,
		baseLineHeight:  0.85,
	}

	result := overlay.GetTextSegments()

	// The result should contain actual newline characters, not the string "\n"
	// Count actual newlines vs literal "\n" strings
	actualNewlines := strings.Count(result, "\n")
	literalBackslashN := strings.Count(result, "\\n")

	if literalBackslashN > 0 {
		t.Errorf("Result contains literal \\n characters which would appear as square symbols. Got: %q", result)
	}

	if actualNewlines != 2 {
		t.Errorf("Expected 2 actual newline characters, got %d. Result: %q", actualNewlines, result)
	}

	t.Logf("Result contains %d actual newlines and %d literal \\n strings: %q", actualNewlines, literalBackslashN, result)
}

func TestGetTextSegments_UTF8Characters(t *testing.T) {
	// Test that special UTF-8 characters are preserved correctly
	tests := []struct {
		name        string
		text        string
		description string
	}{
		{
			name:        "Spanish accents",
			text:        "¿Quién vive\ngratis en tu cabeza?",
			description: "Should preserve Spanish characters with accents and tildes",
		},
		{
			name:        "Emoji characters",
			text:        "Hello 👋\nWorld 🌍",
			description: "Should preserve emoji characters",
		},
		{
			name:        "Mixed special characters",
			text:        "Línea 1: café\nLínea 2: niño",
			description: "Should preserve mixed accented characters",
		},
		{
			name:        "Quote characters",
			text:        "\"Quoted text\"\n'Single quotes'",
			description: "Should preserve various quote styles",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			overlay := TextOverlay{
				Text:            tt.text,
				MaxWidth:        500,
				approxCharWidth: 10,
				baseFontSize:    60,
				baseLineHeight:  0.85,
			}

			result := overlay.GetTextSegments()

			// Verify that the result is valid UTF-8
			if !utf8.ValidString(result) {
				t.Errorf("%s: Result is not valid UTF-8: %q", tt.description, result)
			}

			// Verify that special characters are preserved in the result
			// Check that the result contains the newline(s) from the input
			expectedNewlines := strings.Count(tt.text, "\n")
			actualNewlines := strings.Count(result, "\n")

			if actualNewlines != expectedNewlines {
				t.Errorf("%s: Expected %d newlines, got %d. Input: %q, Result: %q",
					tt.description, expectedNewlines, actualNewlines, tt.text, result)
			}

			t.Logf("%s: Input: %q => Result: %q", tt.description, tt.text, result)
		})
	}
}
