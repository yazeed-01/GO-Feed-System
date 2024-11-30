package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID             primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	Username       string               `bson:"username" json:"username"`
	Bio            string               `bson:"bio,omitempty" json:"bio"`
	Following      []primitive.ObjectID `bson:"following" json:"following"`
	Followers      []primitive.ObjectID `bson:"followers" json:"followers"`
	IsCelebrity    bool                 `bson:"is_celebrity" json:"is_celebrity"`
	CreatedAt      time.Time            `bson:"created_at" json:"created_at"`
	LastFeedUpdate time.Time            `bson:"last_feed_update" json:"last_feed_update"`
}

type Post struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
	Content   string             `bson:"content" json:"content"`
	LikeCount int                `bson:"like_count" json:"like_count"`
	Tags      []string           `bson:"tags,omitempty" json:"tags"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}

type Feed struct {
	ID        primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	UserID    primitive.ObjectID   `bson:"user_id" json:"user_id"`
	Posts     []primitive.ObjectID `bson:"posts" json:"posts"`
	UpdatedAt time.Time            `bson:"updated_at" json:"updated_at"`
}
