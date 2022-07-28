package db

import (
	"testing"
)

func TestGetCharsetFromCollate(t *testing.T) {
	if GetCharsetFromCollate("utf8mb4_unicode_ci") != "utf8mb4" {
		t.Error("wrong collate to charset")
	}
}
