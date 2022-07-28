package statement

import "fmt"

func CreateDatabase(name string) string {
	return fmt.Sprintf("CREATE DATABASE `%s`;", name)
}
