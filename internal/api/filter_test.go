package api

import (
	"encoding/json"
	"testing"

	"github.com/HardMakabaka/KB-Gateway/pkg/types"
)

func TestBuildACLFilter_CustomerDoesNotUseInternalPublic(t *testing.T) {
	f := buildACLFilter(types.Principal{Type: types.PrincipalCustomerUser, Groups: []string{"customer:acme"}})
	b, _ := json.Marshal(f)
	s := string(b)
	if contains(s, "acl_public") {
		t.Fatalf("customer filter should not reference acl_public: %s", s)
	}
	if !contains(s, "acl_external_public") {
		t.Fatalf("expected acl_external_public: %s", s)
	}
}

func contains(s, sub string) bool { return len(s) >= 0 && (stringIndex(s, sub) >= 0) }

func stringIndex(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
