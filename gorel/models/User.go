package models

import "GoRelCli/gorel/enums"

type User struct{
	Id int64 `gorel:"id"`
	Email string `gorel:"email"`
	Username string `gorel:"username"`
	Isverified bool `gorel:"isVerified"`
	Usertype enums.UserRole
 `gorel:"userType"`	Todos []Todo `gorel:"todos"`
}