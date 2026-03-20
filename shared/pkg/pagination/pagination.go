package pagination

import (
	"encoding/base64"
	"fmt"
	"strings"
	"time"
)

const (
	DefaultLimit = 20
	MaxLimit     = 50
	ShortsLimit  = 10
)

// Params holds the parsed pagination parameters from query string.
type Params struct {
	Cursor    string
	Limit     int
	CursorTime *time.Time
	CursorID   *string
}

// ParseParams parses cursor and limit from query parameters.
// If limit is 0 or exceeds max, defaults are applied.
func ParseParams(cursor string, limit int) Params {
	if limit <= 0 {
		limit = DefaultLimit
	}
	if limit > MaxLimit {
		limit = MaxLimit
	}

	p := Params{
		Cursor: cursor,
		Limit:  limit,
	}

	if cursor != "" {
		ct, cid, err := DecodeCursor(cursor)
		if err == nil {
			p.CursorTime = &ct
			p.CursorID = &cid
		}
	}

	return p
}

// EncodeCursor encodes a (created_at, id) pair into an opaque cursor string.
func EncodeCursor(createdAt time.Time, id string) string {
	raw := fmt.Sprintf("%d|%s", createdAt.UnixNano(), id)
	return base64.URLEncoding.EncodeToString([]byte(raw))
}

// DecodeCursor decodes an opaque cursor string back into (created_at, id).
func DecodeCursor(cursor string) (time.Time, string, error) {
	decoded, err := base64.URLEncoding.DecodeString(cursor)
	if err != nil {
		return time.Time{}, "", fmt.Errorf("invalid cursor encoding: %w", err)
	}

	parts := strings.SplitN(string(decoded), "|", 2)
	if len(parts) != 2 {
		return time.Time{}, "", fmt.Errorf("invalid cursor format")
	}

	var nanos int64
	_, err = fmt.Sscanf(parts[0], "%d", &nanos)
	if err != nil {
		return time.Time{}, "", fmt.Errorf("invalid cursor timestamp: %w", err)
	}

	t := time.Unix(0, nanos).UTC()
	id := parts[1]

	return t, id, nil
}

// SQLCondition returns the WHERE clause fragment for cursor-based pagination.
// Parameters should be bound positionally after any existing parameters.
// Returns the SQL fragment and the two values to bind.
func SQLCondition(params Params, paramStart int) (string, []interface{}) {
	if params.CursorTime == nil || params.CursorID == nil {
		return "", nil
	}
	sql := fmt.Sprintf("AND (created_at, id) < ($%d, $%d)", paramStart, paramStart+1)
	return sql, []interface{}{*params.CursorTime, *params.CursorID}
}

// HasMore checks if there are more results beyond the current page.
// Call with len(results) which should be limit+1 if there are more.
func HasMore(resultCount, limit int) bool {
	return resultCount > limit
}

// TrimResults trims results to the requested limit if there were more.
func TrimResults[T any](results []T, limit int) []T {
	if len(results) > limit {
		return results[:limit]
	}
	return results
}

// BuildMeta creates pagination Meta for the response.
// lastItem should be the last item in the trimmed result set.
func BuildMeta(hasMore bool, count int, nextCursor string) map[string]interface{} {
	meta := map[string]interface{}{
		"has_more": hasMore,
		"count":    count,
	}
	if hasMore && nextCursor != "" {
		meta["next_cursor"] = nextCursor
	}
	return meta
}
