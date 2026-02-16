package api

import (
	"github.com/HardMakabaka/KB-Gateway/internal/qdrant"
	"github.com/HardMakabaka/KB-Gateway/pkg/types"
)

func buildBaseFilter(projectScope []string) qdrant.Filter {
	must := []any{
		matchAny("project_id", projectScope),
		matchBool("is_active", true),
		matchBool("deleted", false),
	}
	return qdrant.Filter{"must": must}
}

func buildACLFilter(pr types.Principal) qdrant.Filter {
	// Internal: acl_public OR group intersects
	// Customer: acl_external_public OR group intersects
	should := []any{}
	if pr.Type == types.PrincipalCustomerUser {
		should = append(should, matchBool("acl_external_public", true))
	} else {
		should = append(should, matchBool("acl_public", true))
	}
	if len(pr.Groups) > 0 {
		should = append(should, matchAny("acl_allow", pr.Groups))
	}
	return qdrant.Filter{"should": should}
}

func andFilters(a, b qdrant.Filter) qdrant.Filter {
	mustA, _ := a["must"].([]any)
	mustB, _ := b["must"].([]any)
	shouldA, _ := a["should"].([]any)
	shouldB, _ := b["should"].([]any)
	out := qdrant.Filter{}
	if len(mustA)+len(mustB) > 0 {
		out["must"] = append(mustA, mustB...)
	}
	if len(shouldA)+len(shouldB) > 0 {
		out["should"] = append(shouldA, shouldB...)
	}
	return out
}

func matchAny(field string, values []string) map[string]any {
	return map[string]any{
		"key": field,
		"match": map[string]any{
			"any": values,
		},
	}
}

func matchBool(field string, value bool) map[string]any {
	return map[string]any{
		"key": field,
		"match": map[string]any{
			"value": value,
		},
	}
}
