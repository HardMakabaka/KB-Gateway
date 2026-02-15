package acl

import "github.com/HardMakabaka/KB-Gateway/pkg/types"

type DocACL struct {
	Public         bool     `json:"acl_public"`
	ExternalPublic bool     `json:"acl_external_public"`
	Allow          []string `json:"acl_allow"`
}

func Allowed(pr types.Principal, a DocACL) bool {
	// Customer visibility never uses internal-public.
	switch pr.Type {
	case types.PrincipalCustomerUser:
		if a.ExternalPublic {
			return true
		}
		return intersects(a.Allow, pr.Groups)
	case types.PrincipalInternalUser, types.PrincipalService:
		if a.Public {
			return true
		}
		return intersects(a.Allow, pr.Groups)
	default:
		return false
	}
}

func intersects(a, b []string) bool {
	if len(a) == 0 || len(b) == 0 {
		return false
	}
	set := make(map[string]struct{}, len(a))
	for _, x := range a {
		set[x] = struct{}{}
	}
	for _, y := range b {
		if _, ok := set[y]; ok {
			return true
		}
	}
	return false
}
