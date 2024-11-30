package controllers

import (
	"context"
	"net/http"
	"time"

	"feed/initializers"
	"feed/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var postsCollection *mongo.Collection = initializers.OpenCollection(initializers.Client, "post")

// CreatePost handles the creation of a new post
func CreatePost(c *gin.Context) {
	id, _ := primitive.ObjectIDFromHex(c.Param("id"))
	var post models.Post
	if err := c.ShouldBindJSON(&post); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	post.CreatedAt = time.Now()
	post.LikeCount = 0
	post.UserID = id
	result, err := postsCollection.InsertOne(context.Background(), post)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create post"})
		return
	}

	post.ID = result.InsertedID.(primitive.ObjectID)
	c.JSON(http.StatusCreated, post)
}

// GetPost retrieves a post by ID
func GetPost(c *gin.Context) {
	id, _ := primitive.ObjectIDFromHex(c.Param("id"))
	var post models.Post
	err := postsCollection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&post)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}
	c.JSON(http.StatusOK, post)
}

// UpdatePost updates a post's information
func UpdatePost(c *gin.Context) {
	id, _ := primitive.ObjectIDFromHex(c.Param("id"))
	var updateData bson.M
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := postsCollection.UpdateOne(
		context.Background(),
		bson.M{"_id": id},
		bson.M{"$set": updateData},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update post"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Post updated successfully"})
}

// DeletePost deletes a post
func DeletePost(c *gin.Context) {
	id, _ := primitive.ObjectIDFromHex(c.Param("id"))
	_, err := postsCollection.DeleteOne(context.Background(), bson.M{"_id": id})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete post"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Post deleted successfully"})
}

// ListPosts retrieves a list of all posts
func ListPosts(c *gin.Context) {
	var posts []models.Post
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := postsCollection.Find(context.Background(), bson.M{}, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve posts"})
		return
	}
	defer cursor.Close(context.Background())

	if err = cursor.All(context.Background(), &posts); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode posts"})
		return
	}
	c.JSON(http.StatusOK, posts)
}

// LikePost increments the like count of a post
func LikePost(c *gin.Context) {
	id, _ := primitive.ObjectIDFromHex(c.Param("id"))
	_, err := postsCollection.UpdateOne(
		context.Background(),
		bson.M{"_id": id},
		bson.M{"$inc": bson.M{"like_count": 1}},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to like post"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Post liked successfully"})
}

// UnlikePost decrements the like count of a post
func UnlikePost(c *gin.Context) {
	id, _ := primitive.ObjectIDFromHex(c.Param("id"))
	_, err := postsCollection.UpdateOne(
		context.Background(),
		bson.M{"_id": id},
		bson.M{"$inc": bson.M{"like_count": -1}},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unlike post"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Post unliked successfully"})
}

// AddTag adds a tag to a post
func AddTag(c *gin.Context) {
	id, _ := primitive.ObjectIDFromHex(c.Param("id"))
	var tag struct {
		Tag string `json:"tag"`
	}
	if err := c.ShouldBindJSON(&tag); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := postsCollection.UpdateOne(
		context.Background(),
		bson.M{"_id": id},
		bson.M{"$addToSet": bson.M{"tags": tag.Tag}},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add tag"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Tag added successfully"})
}

// RemoveTag removes a tag from a post
func RemoveTag(c *gin.Context) {
	id, _ := primitive.ObjectIDFromHex(c.Param("id"))
	var tag struct {
		Tag string `json:"tag"`
	}
	if err := c.ShouldBindJSON(&tag); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := postsCollection.UpdateOne(
		context.Background(),
		bson.M{"_id": id},
		bson.M{"$pull": bson.M{"tags": tag.Tag}},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove tag"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Tag removed successfully"})
}

// GetPostsByUser retrieves all posts by a specific user
func GetPostsByUser(c *gin.Context) {
	userID, _ := primitive.ObjectIDFromHex(c.Param("userID"))
	var posts []models.Post
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := postsCollection.Find(context.Background(), bson.M{"user_id": userID}, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve posts"})
		return
	}
	defer cursor.Close(context.Background())

	if err = cursor.All(context.Background(), &posts); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode posts"})
		return
	}
	c.JSON(http.StatusOK, posts)
}
