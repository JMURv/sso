package model

import (
	"github.com/goccy/go-json"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"os"
	"time"
)

type User struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name      string    `json:"name" gorm:"type:varchar(50);not null"`
	Password  string    `json:"password" gorm:"type:varchar(255);not null"`
	Email     string    `json:"email" gorm:"type:varchar(50);not null;unique"`
	Avatar    string    `json:"avatar" gorm:"type:varchar(255)"`
	Address   string    `json:"address" gorm:"type:varchar(255)"`
	Phone     string    `json:"phone" gorm:"type:varchar(20)"`
	IsOpt     bool      `json:"is_opt" gorm:"default:false"`
	IsAdmin   bool      `json:"is_admin" gorm:"default:false"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

func MustPrecreateUsers(conn *gorm.DB) {
	var count int64
	if err := conn.Model(&User{}).Count(&count).Error; err != nil {
		panic(err)
	}

	if count == 0 {
		bytes, err := os.ReadFile("db/precreate/users.json")
		if err != nil {
			panic(err)
		}

		p := make([]*User, 0, 2)
		if err = json.Unmarshal(bytes, &p); err != nil {
			panic(err)
		}

		for _, v := range p {
			password, err := bcrypt.GenerateFromPassword([]byte(v.Password), bcrypt.DefaultCost)
			if err != nil {
				panic(err)
			}
			v.Password = string(password)

			if err = conn.Create(v).Error; err != nil {
				panic(err)
			}
		}

		zap.L().Debug("Users have been created")
	} else {
		zap.L().Debug("Users already exist")
	}
}
