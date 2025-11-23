package domain

type User struct {
	ID       string `gorm:"column:id;primaryKey"`
	Username string `json:"username"`
	TeamName string `json:"team_name"`
	IsActive bool   `json:"isActive"`
}
