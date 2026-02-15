package acl

import (
	"testing"

	"github.com/HardMakabaka/KB-Gateway/pkg/types"
)

func TestAllowed_InternalPublic(t *testing.T) {
	pr := types.Principal{Type: types.PrincipalInternalUser, Groups: nil}
	if !Allowed(pr, DocACL{Public: true}) {
		t.Fatalf("expected allowed")
	}
}

func TestAllowed_CustomerDoesNotInheritInternalPublic(t *testing.T) {
	pr := types.Principal{Type: types.PrincipalCustomerUser, Groups: nil}
	if Allowed(pr, DocACL{Public: true}) {
		t.Fatalf("expected denied")
	}
}

func TestAllowed_CustomerExternalPublic(t *testing.T) {
	pr := types.Principal{Type: types.PrincipalCustomerUser, Groups: nil}
	if !Allowed(pr, DocACL{ExternalPublic: true}) {
		t.Fatalf("expected allowed")
	}
}

func TestAllowed_GroupIntersection(t *testing.T) {
	pr := types.Principal{Type: types.PrincipalCustomerUser, Groups: []string{"customer:acme"}}
	if !Allowed(pr, DocACL{Allow: []string{"customer:acme"}}) {
		t.Fatalf("expected allowed")
	}
}
