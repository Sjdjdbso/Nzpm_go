package upload_utils

import (
	"testing"
)

func TestGetIdFromUrl(t *testing.T) {
	helper := &GoogleDriveHelper{}

	cases := []struct {
		url      string
		expected string
	}{
		{
			url:      "https://drive.google.com/drive/folders/12-mKJDyelDF7ElvmSJanEefBRbqyNHuG",
			expected: "12-mKJDyelDF7ElvmSJanEefBRbqyNHuG",
		},
		{
			url:      "https://drive.google.com/file/d/1J-rhw82m_4h9MV71h0D86erd8BxjGOK7/view",
			expected: "1J-rhw82m_4h9MV71h0D86erd8BxjGOK7",
		},
		{
			url:      "https://drive.google.com/open?id=1KUGbfbwvFJeZSEbv1oY24XLbRh-j8kCH_JHYK_5yRYg",
			expected: "1KUGbfbwvFJeZSEbv1oY24XLbRh-j8kCH_JHYK_5yRYg",
		},
		{
			url:      "https://drive.google.com/uc?id=1VndvVu_qCWTNzKMAJRq7OEIGXBUsNwKN7axYhIo0nxKbZGo_xZRIM_cz&export=download",
			expected: "1VndvVu_qCWTNzKMAJRq7OEIGXBUsNwKN7axYhIo0nxKbZGo_xZRIM_cz",
		},
		{
			url:      "12-mKJDyelDF7ElvmSJanEefBRbqyNHuG",
			expected: "12-mKJDyelDF7ElvmSJanEefBRbqyNHuG",
		},
	}

	for _, c := range cases {
		id, err := helper.GetIdFromUrl(c.url)
		if err != nil {
			t.Errorf("Unexpected error for %s: %v", c.url, err)
		}
		if id != c.expected {
			t.Errorf("For %s, expected %s, got %s", c.url, c.expected, id)
		}
	}
}
