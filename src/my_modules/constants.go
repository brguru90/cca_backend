package my_modules

type AccessLevelType string

var AccessLevel = struct {
	SUPER_ADMIN AccessLevelType
	ADMIN       AccessLevelType
	CUSTOMER    AccessLevelType
}{
	SUPER_ADMIN: "super_admin",
	ADMIN:       "admin",
	CUSTOMER:    "customer",
}
