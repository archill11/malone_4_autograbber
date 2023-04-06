package entity

type User struct {
	Id        int
	Username  string
	Firstname string
	IsAdmin   int
}

func NewUser(id int, username, firstname string, botId int) User {
	u := User{
		Id:        id,
		Username:  username,
		Firstname: firstname,
		IsAdmin:   0,
	}
	return u
}
