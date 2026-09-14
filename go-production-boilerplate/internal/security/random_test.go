package security

import "testing"

func TestRandomTokenAndHash(t *testing.T) {
	a, err := RandomToken(32)
	if err != nil {
		t.Fatal(err)
	}
	b, err := RandomToken(32)
	if err != nil {
		t.Fatal(err)
	}
	if a == b {
		t.Fatal("random tokens should differ")
	}
	if HashToken(a) == a || len(HashToken(a)) != 64 {
		t.Fatal("token hash should be a 64-character SHA-256 hex string")
	}
}
