package models


type Todo struct{
	Id string `gorel:"id"`
	Title string `gorel:"title"`
	Userid int64 `gorel:"userId"`
}