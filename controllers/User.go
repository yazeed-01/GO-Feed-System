package controllers

import (
	"net/http"
	"time"

	"feed/initializers"
	"feed/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var usersCollection *mongo.Collection = initializers.OpenCollection(initializers.Client, "user")

func CreateUser(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Initialize followers and following as empty arrays if not set
	if user.Followers == nil {
		user.Followers = []primitive.ObjectID{}
	}
	if user.Following == nil {
		user.Following = []primitive.ObjectID{}
	}

	user.CreatedAt = time.Now()

	// Insert the new user into the database
	result, err := usersCollection.InsertOne(c.Request.Context(), user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// Set the user's ID after insertion
	user.ID = result.InsertedID.(primitive.ObjectID)

	c.JSON(http.StatusCreated, user)
}

// GetUser retrieves a user by ID
func GetUser(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var user models.User
	err = usersCollection.FindOne(c.Request.Context(), bson.M{"_id": id}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	c.JSON(http.StatusOK, user)
}

// UpdateUser updates a user's information
func UpdateUser(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var updateData bson.M
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err = usersCollection.UpdateOne(
		c.Request.Context(),
		bson.M{"_id": id},
		bson.M{"$set": updateData},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User updated successfully"})
}

// DeleteUser deletes a user
func DeleteUser(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	_, err = usersCollection.DeleteOne(c.Request.Context(), bson.M{"_id": id})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

// ListUsers retrieves a list of all users
func ListUsers(c *gin.Context) {
	var users []models.User
	cursor, err := usersCollection.Find(c.Request.Context(), bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve users"})
		return
	}
	defer cursor.Close(c.Request.Context())

	if err = cursor.All(c.Request.Context(), &users); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode users"})
		return
	}
	c.JSON(http.StatusOK, users)
}
func FollowUser(c *gin.Context) {
	followerID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid follower ID"})
		return
	}
	followeeID, err := primitive.ObjectIDFromHex(c.Param("followeeID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid followee ID"})
		return
	}

	// Update followee's followers
	_, err = usersCollection.UpdateOne(
		c.Request.Context(),
		bson.M{"_id": followeeID},
		bson.M{
			"$addToSet": bson.M{
				"followers": followerID,
			},
		},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update followee", "details": err.Error()})
		return
	}

	// Update follower's following
	_, err = usersCollection.UpdateOne(
		c.Request.Context(),
		bson.M{"_id": followerID},
		bson.M{
			"$addToSet": bson.M{
				"following": followeeID,
			},
		},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update follower", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully followed user"})
}

// UnfollowUser handles the unfollow action
func UnfollowUser(c *gin.Context) {
	followerID, err := primitive.ObjectIDFromHex(c.Param("followerID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid follower ID"})
		return
	}
	followeeID, err := primitive.ObjectIDFromHex(c.Param("followeeID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid followee ID"})
		return
	}

	// Attempt to update the follower and followee in a single operation
	_, err = usersCollection.UpdateOne(
		c.Request.Context(),
		bson.M{"_id": followerID},
		bson.M{"$pull": bson.M{"following": followeeID}},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update follower"})
		return
	}

	_, err = usersCollection.UpdateOne(
		c.Request.Context(),
		bson.M{"_id": followeeID},
		bson.M{"$pull": bson.M{"followers": followerID}},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update followee"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully unfollowed user"})
}

// SetCelebrityStatus sets the celebrity status of a user
func SetCelebrityStatus(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var status struct {
		IsCelebrity bool `json:"is_celebrity"`
	}
	if err := c.ShouldBindJSON(&status); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err = usersCollection.UpdateOne(
		c.Request.Context(),
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"is_celebrity": status.IsCelebrity}},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update celebrity status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Celebrity status updated"})
}
