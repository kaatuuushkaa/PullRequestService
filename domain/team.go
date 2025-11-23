package domain

type Team struct {
	Name    string `gorm:"primaryKey" json:"team_name"`
	Members []User `gorm:"foreignKey:TeamName;references:Name"`
}
