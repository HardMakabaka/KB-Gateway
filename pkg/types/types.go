package types

type PrincipalType string

const (
	PrincipalInternalUser PrincipalType = "internal_user"
	PrincipalCustomerUser PrincipalType = "customer_user"
	PrincipalService      PrincipalType = "service"
)

type Principal struct {
	Type   PrincipalType `json:"type"`
	ID     string        `json:"id"`
	Groups []string      `json:"groups"`
}
