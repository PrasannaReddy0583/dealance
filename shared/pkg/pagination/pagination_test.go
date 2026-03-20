package pagination_test

import (
	"testing"
	"time"

	"github.com/dealance/shared/pkg/pagination"
)

func TestEncodeDecode(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Nanosecond)
	id := "550e8400-e29b-41d4-a716-446655440000"

	cursor := pagination.EncodeCursor(now, id)
	if cursor == "" {
		t.Fatal("cursor must not be empty")
	}

	decodedTime, decodedID, err := pagination.DecodeCursor(cursor)
	if err != nil {
		t.Fatalf("DecodeCursor: %v", err)
	}
	if decodedID != id {
		t.Errorf("ID = %q, want %q", decodedID, id)
	}
	if decodedTime.UnixNano() != now.UnixNano() {
		t.Errorf("Time = %v, want %v", decodedTime, now)
	}
}

func TestDecodeInvalidCursor(t *testing.T) {
	_, _, err := pagination.DecodeCursor("not-valid-base64!!!")
	if err == nil {
		t.Error("expected error for invalid cursor")
	}
}

func TestParseParams_Defaults(t *testing.T) {
	p := pagination.ParseParams("", 0)
	if p.Limit != pagination.DefaultLimit {
		t.Errorf("Limit = %d, want %d", p.Limit, pagination.DefaultLimit)
	}
	if p.CursorTime != nil {
		t.Error("CursorTime should be nil")
	}
}

func TestParseParams_MaxLimit(t *testing.T) {
	p := pagination.ParseParams("", 100)
	if p.Limit != pagination.MaxLimit {
		t.Errorf("Limit = %d, want %d (maxlimit)", p.Limit, pagination.MaxLimit)
	}
}

func TestParseParams_WithCursor(t *testing.T) {
	now := time.Now().UTC()
	id := "test-id"
	cursor := pagination.EncodeCursor(now, id)

	p := pagination.ParseParams(cursor, 10)
	if p.CursorTime == nil {
		t.Fatal("CursorTime should not be nil")
	}
	if p.CursorID == nil {
		t.Fatal("CursorID should not be nil")
	}
	if *p.CursorID != id {
		t.Errorf("CursorID = %q, want %q", *p.CursorID, id)
	}
}

func TestHasMore(t *testing.T) {
	if !pagination.HasMore(21, 20) {
		t.Error("HasMore(21, 20) should be true")
	}
	if pagination.HasMore(20, 20) {
		t.Error("HasMore(20, 20) should be false")
	}
	if pagination.HasMore(10, 20) {
		t.Error("HasMore(10, 20) should be false")
	}
}

func TestTrimResults(t *testing.T) {
	items := []int{1, 2, 3, 4, 5}
	trimmed := pagination.TrimResults(items, 3)
	if len(trimmed) != 3 {
		t.Errorf("len = %d, want 3", len(trimmed))
	}

	noTrim := pagination.TrimResults(items, 10)
	if len(noTrim) != 5 {
		t.Errorf("len = %d, want 5", len(noTrim))
	}
}

func TestSQLCondition(t *testing.T) {
	p := pagination.ParseParams("", 20)
	sql, args := pagination.SQLCondition(p, 1)
	if sql != "" {
		t.Errorf("sql should be empty for no cursor, got %q", sql)
	}
	if args != nil {
		t.Error("args should be nil")
	}

	now := time.Now().UTC()
	cursor := pagination.EncodeCursor(now, "id-1")
	p2 := pagination.ParseParams(cursor, 20)
	sql2, args2 := pagination.SQLCondition(p2, 3)
	if sql2 == "" {
		t.Error("sql should not be empty with cursor")
	}
	if len(args2) != 2 {
		t.Errorf("expected 2 args, got %d", len(args2))
	}
}
