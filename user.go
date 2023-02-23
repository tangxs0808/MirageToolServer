package main

import (
	"encoding/base64"
	"time"

	"github.com/rs/zerolog/log"
)

type User struct {
	ID     string `gorm:"primary_key;unique;not null"`
	Name   string `gorm:"not null"`
	Avatar []byte `gorm:"type:blob"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (t *MirageTool) UpdateOrCreateUser(ID string, name string, avatar string) *User {
	user := User{}
	avatarData, err := base64.StdEncoding.DecodeString(avatar)
	if err != nil {
		avatarData = nil
	}
	if err := t.db.Where("id = ?", ID).First(&user).Error; err == nil {
		user.ID = ID
		user.Name = name
		user.Avatar = avatarData
		user.UpdatedAt = time.Now()
		t.db.Save(user)
		return &user
	}
	user.ID = ID
	user.Name = name
	user.Avatar = avatarData
	if err := t.db.Create(&user).Error; err != nil {
		log.Error().
			Str("func", "CreateUser").
			Err(err).
			Msg("Could not create row")
	}

	return &user
}

func (t *MirageTool) GetUserByID(ID string) *User {
	user := User{}
	if err := t.db.Where("id = ?", ID).First(&user).Error; err == nil {
		return &user
	}
	return nil
}
