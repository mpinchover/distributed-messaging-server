package records

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Message struct {
	ID            primitive.ObjectID `bson:"_id,omitempty"`
	CreatedAt     time.Time          `bson:"created_at"`
	UpdatedAt     time.Time          `bson:"updated_at"`
	DeletedAt     *time.Time         `bson:"deleted_at,omitempty"`
	UserUUID      string             `bson:"user_uuid"`
	RoomUUID      string             `bson:"room_uuid"`
	RoomID        int                `bson:"room_id,omitempty"`
	MessageText   string             `bson:"message_text"`
	UUID          string             `bson:"uuid"`
	MessageStatus string             `bson:"message_status"`
	CreatedAtNano float64 `bson:"created_at_nano"`
}

type Room struct {
	ID            primitive.ObjectID `bson:"_id,omitempty"`
	CreatedAt     time.Time          `bson:"created_at"`
	UpdatedAt     time.Time          `bson:"updated_at"`
	DeletedAt     *time.Time         `bson:"deleted_at,omitempty"`
	UUID          string             `bson:"uuid"`
	CreatedAtNano float64            `bson:"created_at_nano"`

	Members  []*Member  `bson:"-"`
	Messages []*Message `bson:"-"`
}

type Member struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	CreatedAt time.Time          `bson:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at"`
	DeletedAt *time.Time         `bson:"deleted_at,omitempty"`
	RoomUUID  string             `bson:"room_uuid"`
	RoomID    int                `bson:"room_id,omitempty"`
	UserUUID  string             `bson:"user_uuid"`
}

// /* AUTH   */
// // for ext service, not chat user
// type AuthProfile struct {
// 	gorm.Model
// 	UUID           string
// 	Email          string
// 	HashedPassword string
// 	Mobile         string
// }

// /* MATCHING   */

// // after user has answered
// type TrackedQuestion struct {
// 	gorm.Model
// 	UUID         string
// 	QuestionText string
// 	Category     string
// 	UserUUID     string
// 	QuestionUUID string
// 	Liked        bool
// }

// type DiscoverProfile struct {
// 	gorm.Model
// 	Gender           string
// 	GenderPreference string
// 	Age              int64
// 	MinAgePref       int64
// 	MaxAgePref       int64
// 	UserUUID         string
// 	CurrentLat       float64
// 	CurrentLng       float64
// }

// type TrackedLike struct {
// 	gorm.Model
// 	UUID       string
// 	UserUUID   string
// 	TargetUUID string
// 	Liked      bool
// }
