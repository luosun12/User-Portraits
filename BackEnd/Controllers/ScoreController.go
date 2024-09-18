package Controllers

import (
	"UserPortrait/etc"
	"database/sql"
	"strconv"
	"time"
)

func (s *SqlController) InsertScore(userID uint, score float32) error {
	var scoreData = etc.Score{UserID: userID, Score: score, Date: time.Now().Format(time.DateOnly)}
	result := s.DB.Table("network_score").Create(&scoreData).Error
	return result
}

func (s *SqlController) UpdateScore(userID uint, score float32) error {
	var scoreData = etc.Score{UserID: userID, Score: score, Date: time.Now().Format(time.DateOnly)}
	result := s.DB.Table("network_score").Where("user_id = ?", userID).Updates(&scoreData).Error
	return result
}

func (s *SqlController) FindScoreRecord(userID uint, date string) error {
	result := s.DB.Table("network_score").Where("user_id = ? AND date = ?", userID, date).First(&etc.Score{}).Error
	return result
}

func (s *SqlController) AverageScoreByDate() ([]etc.AverageScoreInterface, error) {
	var aves []etc.AverageScoreInterface
	rows, err := s.DB.Table("network_score").Select("date as date,AVG(score) as average_score").Group("date").Rows()
	if err != nil {
		return []etc.AverageScoreInterface{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var date string
		var average sql.RawBytes
		if err = rows.Scan(&date, &average); err != nil {
			return []etc.AverageScoreInterface{}, err
		}
		averageFloat, err := strconv.ParseFloat(string(average), 64)
		if err != nil {
			return []etc.AverageScoreInterface{}, err
		}
		aves = append(aves, etc.AverageScoreInterface{Date: date, Average: float32(averageFloat)})
	}
	return aves, nil
}
