package service

import "myapp/internal/entity"

func (s *AppService) GetUserById(id int) (entity.User, error) {
	return s.db.GetUserById(id)
}

func (s *AppService) GetUserByUsername(username string) (entity.User, error) {
	return s.db.GetUserByUsername(username)
}

func (s *AppService) AddNewUser(id int, username, firstname string) error {
	return s.db.AddNewUser(id, username, firstname)
}

func (s *AppService) EditAdmin(username string, is_admin int) error {
	return s.db.EditAdmin(username, is_admin)
}
