package model

import "time"

// UserSession stocke le token OAuth Google associé à un cookie de session
type UserSession struct {
	ID           string    `json:"id" gorm:"primaryKey"`
	AccessToken  string    `json:"-" gorm:"type:text;not null"`
	RefreshToken string    `json:"-" gorm:"type:text"`
	TokenType    string    `json:"-"`
	Expiry       time.Time `json:"expiry"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
