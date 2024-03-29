package repo

import (
	"fmt"
	"messaging-service/src/types/records"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

const (
	PAGINATION_MESSAGES = 20
	PAGINATION_ROOMS    = 10
)

type RepoInterface interface {
	// messaging
	LeaveRoom(userUUID string, roomUUID string) error
	UpdateMessage(message *records.Message) error
	GetMembersByRoomUUID(roomUUID string) ([]*records.Member, error)
	GetMessageByUUID(uuid string) (*records.Message, error)
	SaveSeenBy(seenBy *records.SeenBy) error
	GetRoomByRoomUUID(roomUUID string) (*records.Room, error)
	SaveMessage(msg *records.Message) error
	GetMessagesByRoomUUID(roomUUID string, offset int) ([]*records.Message, error)
	GetMessagesByRoomUUIDs(roomUUIDs string, offset int) ([]*records.Message, error)
	GetRoomsByUserUUID(uuid string, offset int) ([]*records.Room, error)
	DeleteRoom(roomUUID string) error
	SaveRoom(room *records.Room) error
	GetRoomsByUserUUIDForSubscribing(userUUID string) ([]*records.Room, error)
}

type Repo struct {
	DB *gorm.DB
}

func connect() (*gorm.DB, error) {
	dsn := fmt.Sprintf("root:root@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", os.Getenv("MYSQL_HOST"), os.Getenv("MYSQL_PORT"), os.Getenv("MYSQL_DB_NAME"))
	return gorm.Open(mysql.Open(dsn), &gorm.Config{
		// Logger: logger.Default.LogMode(logger.Info),
	})
}

func New() *Repo {
	var db *gorm.DB
	db, err := connect()
	if err != nil {
		panic(err)
	}

	return &Repo{
		DB: db,
	}
}
