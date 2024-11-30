package routes

import (
	"feed/controllers"

	"github.com/gin-gonic/gin"
)

func SetupRoutes() *gin.Engine {
	r := gin.Default()

	// User routes
	r.POST("/users", controllers.CreateUser)                             // create a new user
	r.GET("/users/:id", controllers.GetUser)                             // get a user by ID
	r.PUT("/users/:id", controllers.UpdateUser)                          // update a user
	r.DELETE("/users/:id", controllers.DeleteUser)                       // delete a user
	r.GET("/users", controllers.ListUsers)                               // list all users
	r.POST("/users/:id/follow/:followeeID", controllers.FollowUser)      // follow a user
	r.POST("/users/:id/unfollow/:followeeID", controllers.UnfollowUser)  // unfollow a user
	r.PUT("/users/:id/celebrity-status", controllers.SetCelebrityStatus) // set the celebrity status of a user

	// Post routes
	r.POST("/users/:id/posts", controllers.CreatePost)    // create a new post
	r.GET("/posts/:id", controllers.GetPost)              // get a post by ID
	r.PUT("/posts/:id", controllers.UpdatePost)           // update a post
	r.DELETE("/posts/:id", controllers.DeletePost)        // delete a post
	r.GET("/posts", controllers.ListPosts)                // list all posts
	r.POST("/posts/:id/like", controllers.LikePost)       // like a post
	r.POST("/posts/:id/unlike", controllers.UnlikePost)   // unlike a post
	r.POST("/posts/:id/tags", controllers.AddTag)         // add a tag to a post
	r.DELETE("/posts/:id/tags", controllers.RemoveTag)    // remove a tag from a post
	r.GET("/users/:id/posts", controllers.GetPostsByUser) // get all posts by a user

	// Feed routes
	r.GET("/feeds/:id", controllers.GetFeed)
	return r
}
