package models

import "GoRelCli/gorel/enums"

type User struct{
	id int64
	email string
	username string
	isVerified bool
	userType enums.UserRole
	todos []Todo
}