package service

import "myapp/internal/entity"

func (s *AppService) GetUserById(id int) (entity.User, error) {
	u, err := s.db.GetUserById(id)
	if err != nil {
		return u, err
	}
	return u, nil
}

func (s *AppService) AddNewUser(id int, username, firstname string) error {
	err := s.db.AddNewUser(id, username, firstname)
	if err != nil {
		return err
	}
	return nil
}

func (s *AppService) EditAdmin(username string, is_admin int) error {
	err := s.db.EditAdmin(username, is_admin)
	if err != nil {
		return err
	}
	return nil
}
