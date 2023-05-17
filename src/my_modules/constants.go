package my_modules

type AccessLevelType struct {
	Label  string
	Weight int
}

var AccessLevel = struct {
	SUPER_ADMIN AccessLevelType
	ADMIN       AccessLevelType
	CUSTOMER    AccessLevelType
}{
	SUPER_ADMIN: AccessLevelType{Label: "super_admin", Weight: 100},
	ADMIN:       AccessLevelType{Label: "admin", Weight: 99},
	CUSTOMER:    AccessLevelType{Label: "customer", Weight: 1},
	// ADMIN:       "admin",
	// CUSTOMER:    "customer",
}

var AllAccessLevel = map[string]bool{
	AccessLevel.SUPER_ADMIN.Label: true,
	AccessLevel.ADMIN.Label:       true,
	AccessLevel.CUSTOMER.Label:    true,
}

var AllAccessLevelReverseMap = map[string]AccessLevelType{
	AccessLevel.SUPER_ADMIN.Label: AccessLevel.SUPER_ADMIN,
	AccessLevel.ADMIN.Label:       AccessLevel.ADMIN,
	AccessLevel.CUSTOMER.Label:    AccessLevel.CUSTOMER,
}
