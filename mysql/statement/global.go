package statement

func DisableForeignKeyChecks() string {
	return "SET FOREIGN_KEY_CHECKS = 0;"
}

func EnableForeignKeyChecks() string {
	return "SET FOREIGN_KEY_CHECKS = 1;"
}
