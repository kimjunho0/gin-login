package main

import "gin-login/migrate"

func main() {
	migrate.ConnectDB()
}
