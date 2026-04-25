package roles

type Role string

const (
	RoleAdmin   Role = "admin"
	RoleManager Role = "manager"
	RoleClient  Role = "client"
)

func Valid(role string) bool {
	switch Role(role) {
	case RoleAdmin, RoleManager, RoleClient:
		return true
	default:
		return false
	}
}
