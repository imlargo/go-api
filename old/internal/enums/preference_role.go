package enums

type UserType string

const (
	UserTypeIndividual UserType = "individual"
	UserTypeAgency     UserType = "agency"
)

func (r UserType) IsValid() bool {
	switch r {
	case UserTypeIndividual, UserTypeAgency:
		return true
	default:
		return false
	}
}
