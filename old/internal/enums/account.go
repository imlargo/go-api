package enums

type AccountRole string

const (
	AccountRoleMain   AccountRole = "main"
	AccountRoleBackup AccountRole = "backup"
)

func (ar AccountRole) IsValid() bool {
	switch ar {
	case AccountRoleMain, AccountRoleBackup:
		return true
	default:
		return false
	}
}

type AccountStatus string

const (
	AccountStatusActive   AccountStatus = "active"
	AccountStatusInactive AccountStatus = "inactive"
)

func (as AccountStatus) IsValid() bool {
	switch as {
	case AccountStatusActive, AccountStatusInactive:
		return true
	default:
		return false
	}
}
