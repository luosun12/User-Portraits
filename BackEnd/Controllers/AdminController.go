package Controllers

import (
	"UserPortrait/etc"
)

func (s *SqlController) FindAdminByName(name string) (etc.Admininfo, error) {
	var admin etc.Admininfo
	err := s.DB.Table("admin_info").Where("adminname = ?", name).Take(&admin).Error
	return admin, err
}

func (s *SqlController) InsertAdmin(admin etc.Admininfo) error {
	err := s.DB.Table("admin_info").Create(&admin).Error
	return err
}

func (s *SqlController) UpdateAdminByID(id uint, name string, pswd string) error {

	err := s.DB.Table("admin_info").Where("id = ?", id).Updates(map[string]interface{}{
		"adminname": name, "password": pswd}).Error
	return err
}
