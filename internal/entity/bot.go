package entity

type Bot struct {
	Id          int
	Token       string
	Username    string
	Firstname   string
	IsDonor     int
	ChId        int
	ChLink      string
	GroupLinkId int
}

func NewBot(id int, username, firstname, token string, isDonor int) Bot {
	b := Bot{
		Id:        id,
		Username:  username,
		Firstname: firstname,
		Token:     token,
		IsDonor:   isDonor,
	}
	return b
}
