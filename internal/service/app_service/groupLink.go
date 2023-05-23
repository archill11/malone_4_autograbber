package service

import "myapp/internal/entity"

func (s *AppService) AddNewGroupLink(title, link string) error {
	return s.db.AddNewGroupLink(title, link)
}

func (s *AppService) DeleteGroupLink(id int) error {
	return s.db.DeleteGroupLink(id)
}

func (s *AppService) UpdateGroupLink(id int, link string) error {
	return s.db.UpdateGroupLink(id, link)
}

func (s *AppService) GetAllGroupLinks() ([]entity.GroupLink, error) {
	return s.db.GetAllGroupLinks()
}

func (s *AppService) GetGroupLinkById(id int) (entity.GroupLink, error) {
	return s.db.GetGroupLinkById(id)
}
