package enums

type UserRole string

const (
	UserRoleUser   UserRole = "user"
	UserRolePoster UserRole = "poster"
	UserRoleAgency UserRole = "admin/agency"
	UserRoleClient UserRole = "client"
	UserRoleLeader UserRole = "team_leader"
	UserRoleAdmin  UserRole = "admin"
)

func (r UserRole) IsValid() bool {
	switch r {
	case UserRoleUser, UserRolePoster, UserRoleAgency, UserRoleClient, UserRoleLeader, UserRoleAdmin:
		return true
	default:
		return false
	}
}
