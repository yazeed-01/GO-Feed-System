package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"feed/initializers"
	"feed/models"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var feedsCollection *mongo.Collection = initializers.OpenCollection(initializers.Client, "feed")
var redisClient *redis.Client = initializers.OpenRedis()

func GetFeed(c *gin.Context) {
	// Get userID from params and handle invalid ObjectID
	userID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Parse pagination parameters
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page number"})
		return
	}
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit number"})
		return
	}
	skip := (page - 1) * limit

	// Check Redis cache first
	cacheKey := fmt.Sprintf("feed:%s:%d:%d", userID.Hex(), page, limit)
	cachedFeed, err := redisClient.Get(context.Background(), cacheKey).Result()

	if err == nil {
		// Cache hit: deserialize cached feed
		var posts []models.Post
		err := json.Unmarshal([]byte(cachedFeed), &posts)
		if err != nil {
			fmt.Println("Error deserializing cached feed:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to deserialize feed"})
			return
		}

		// Return cached posts
		c.JSON(http.StatusOK, gin.H{
			"posts": posts,
			"page":  page,
			"limit": limit,
		})
		return
	}

	// Fetch user data from DB
	var user models.User
	err = usersCollection.FindOne(context.Background(), bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user data"})
		}
		return
	}

	// Fetch feed data or create a new feed if not found
	var feed models.Feed
	err = feedsCollection.FindOne(context.Background(), bson.M{"user_id": userID}).Decode(&feed)
	if err == mongo.ErrNoDocuments {
		feed = models.Feed{
			UserID:    userID,
			Posts:     []primitive.ObjectID{},
			UpdatedAt: time.Now(),
		}
		_, err = feedsCollection.InsertOne(context.Background(), feed)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create feed"})
			return
		}
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve feed"})
		return
	}

	// Check if feed needs updating (for non-celebrity follows)
	if user.LastFeedUpdate.Before(feed.UpdatedAt) {
		err = updateFeed(userID, &feed)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update feed"})
			return
		}
	}

	// Fetch posts from the feed
	var posts []models.Post
	cursor, err := postsCollection.Find(
		context.Background(),
		bson.M{
			"_id": bson.M{"$in": feed.Posts},
		},
		options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}).SetSkip(int64(skip)).SetLimit(int64(limit)),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve posts"})
		return
	}
	defer cursor.Close(context.Background())

	if err = cursor.All(context.Background(), &posts); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode posts"})
		return
	}

	// Cache the response for future requests
	serializedPosts, err := json.Marshal(posts)
	if err != nil {
		fmt.Println("Error serializing feed:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to serialize feed"})
		return
	}

	err = redisClient.Set(context.Background(), cacheKey, serializedPosts, 10*time.Minute).Err() // 10-minute cache expiry
	if err != nil {
		fmt.Println("Error caching feed:", err)
	}

	// Return the posts
	c.JSON(http.StatusOK, gin.H{
		"posts": posts,
		"page":  page,
		"limit": limit,
	})
}

// updateFeed updates the user's feed with new posts based on their following list and optimizes fanout for celebrities.
func updateFeed(userID primitive.ObjectID, feed *models.Feed) error {
	// Fetch user data again
	var user models.User
	err := usersCollection.FindOne(context.Background(), bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		return err
	}

	// Get posts from non-celebrity users that the current user is following
	followingIDs := make([]primitive.ObjectID, 0)
	for _, id := range user.Following {
		var followedUser models.User
		err := usersCollection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&followedUser)
		if err == nil && !followedUser.IsCelebrity {
			followingIDs = append(followingIDs, id)
		}
	}

	// Query new posts based on the following users
	cursor, err := postsCollection.Find(
		context.Background(),
		bson.M{
			"user_id":    bson.M{"$in": followingIDs},
			"created_at": bson.M{"$gt": user.LastFeedUpdate},
		},
		options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}),
	)
	if err != nil {
		return err
	}
	defer cursor.Close(context.Background())

	var newPosts []models.Post
	if err = cursor.All(context.Background(), &newPosts); err != nil {
		return err
	}

	// Update feed with new posts
	newPostIDs := make([]primitive.ObjectID, len(newPosts))
	for i, post := range newPosts {
		newPostIDs[i] = post.ID
	}
	feed.Posts = append(newPostIDs, feed.Posts...)
	feed.UpdatedAt = time.Now()

	_, err = feedsCollection.UpdateOne(
		context.Background(),
		bson.M{"user_id": userID},
		bson.M{
			"$set": bson.M{
				"posts":      feed.Posts,
				"updated_at": feed.UpdatedAt,
			},
		},
	)
	if err != nil {
		return err
	}

	// Update user's last feed update time
	_, err = usersCollection.UpdateOne(
		context.Background(),
		bson.M{"_id": userID},
		bson.M{"$set": bson.M{"last_feed_update": feed.UpdatedAt}},
	)
	return err
}
