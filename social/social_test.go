package social_test

import (
	"social-2-telego/social"
	"testing"
)

func TestValidateTwitter(t *testing.T) {
	instance := social.Twitter{}

	if err := instance.SetURL("https://twitter.com/loremipsum/status/1234567890"); err != nil {
		t.Errorf("Error: %v", err)
	}

	if err := instance.SetURL("https://x.com/loremipsum/status/1234567890"); err != nil {
		t.Errorf("Error: %v", err)
	}

	if err := instance.SetURL("https://example.com/loremipsum/status/1234567890"); err == nil {
		t.Errorf("Expected error, got nil")
	}
}
