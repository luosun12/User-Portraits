package SQLController

import (
	"UserPortrait/etc"
)

func (s *SqlController) FindUserByMAC(mac string) (etc.Userinfo, error) {
	var user etc.Userinfo
	err := s.DB.Table("user_info").Where("mac_info = ?", mac).Take(&user).Error
	return user, err
}

func (s *SqlController) FindUserByName(name string) (etc.Userinfo, error) {
	var user etc.Userinfo
	err := s.DB.Table("user_info").Where("username = ?", name).Take(&user).Error
	return user, err
}

func (s *SqlController) InsertUser(user etc.Userinfo) {
	err := s.DB.Table("user_info").Create(&user).Error
	if err != nil {
		panic(err)
	}
}

func (s *SqlController) UpdateUserByID(id uint, name string, pswd string) {

	err := s.DB.Table("user_info").Where("id = ?", id).Updates(map[string]interface{}{
		"username": name, "password": pswd}).Error
	if err != nil {
		panic(err)
	}
}
