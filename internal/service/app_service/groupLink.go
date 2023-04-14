package service

import "myapp/internal/entity"

func (s *AppService) AddNewGroupLink(title, link string) error {
	err := s.db.AddNewGroupLink(title, link)
	return err
}

func (s *AppService) DeleteGroupLink(id int) error {
	err := s.db.DeleteGroupLink(id)
	return err
}

func (s *AppService) GetAllGroupLinks() ([]entity.GroupLink, error) {
	grs, err := s.db.GetAllGroupLinks()
	return grs, err
}

func (s *AppService) GetGroupLinkById(id int) (entity.GroupLink, error) {
	gr, err := s.db.GetGroupLinkById(id)
	return gr, err
}
