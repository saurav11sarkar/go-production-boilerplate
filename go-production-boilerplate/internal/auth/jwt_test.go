package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/saurav11sarkar/go-production-boilerplate/internal/domain"
)

func TestJWTCreateAndParse(t *testing.T) {
	manager := NewJWTManager("test-api", "12345678901234567890123456789012", time.Minute)
	id := uuid.New()
	token, _, err := manager.CreateAccessToken(id, domain.RoleAdmin, 3)
	if err != nil {
		t.Fatal(err)
	}
	gotID, gotRole, gotVersion, err := manager.ParseAccessToken(token)
	if err != nil {
		t.Fatal(err)
	}
	if gotID != id || gotRole != domain.RoleAdmin || gotVersion != 3 {
		t.Fatalf("unexpected claims: id=%s role=%s tokenVersion=%d", gotID, gotRole, gotVersion)
	}
}
