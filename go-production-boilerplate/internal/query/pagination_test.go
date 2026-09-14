package query

import (
	"net/url"
	"testing"
)

func TestParseClampsPagination(t *testing.T) {
	values := url.Values{}
	values.Set("page", "0")
	values.Set("limit", "500")
	values.Set("sortOrder", "wrong")
	p := Parse(values)
	if p.Page != 1 {
		t.Fatalf("expected page 1, got %d", p.Page)
	}
	if p.Limit != MaxLimit {
		t.Fatalf("expected max limit %d, got %d", MaxLimit, p.Limit)
	}
	if p.SortOrder != "desc" {
		t.Fatalf("expected desc sort order, got %s", p.SortOrder)
	}
}

func TestMeta(t *testing.T) {
	m := NewMeta(2, 20, 45)
	if m.TotalPages != 3 || !m.HasNext || !m.HasPrevious {
		t.Fatalf("unexpected meta: %+v", m)
	}
}
