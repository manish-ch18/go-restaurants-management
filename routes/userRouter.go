package routes

import (
	"go-restaurants-management/controllers"

	"github.com/gin-gonic/gin"
)

func UserRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.GET("/users", controllers.GetUsers())
	incomingRoutes.GET("/users/:user_id", controllers.GetUser())
	incomingRoutes.POST("/users/signup", controllers.SignUp())
	incomingRoutes.POST("/users/login", controllers.Login())
	incomingRoutes.POST("/users", controllers.CreateUser())
	incomingRoutes.PATCH("/users/:user_id", controllers.UpdateUser())
	incomingRoutes.DELETE("/users/:user_id", controllers.DeleteUser())
}
