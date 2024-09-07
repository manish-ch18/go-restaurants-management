package controllers

import (
	"context"
	"go-restaurants-management/database"
	"go-restaurants-management/models"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var menuCollection *mongo.Collection = database.OpenColletion(database.Client, "menu")

func GetMenus() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		result, err := menuCollection.Find(context.TODO(), bson.M{})
		defer cancel()
		if err != nil {
			errorRes := models.ErrorResponse(http.StatusInternalServerError, "error", "error occured while listin menu items.")
			c.JSON(errorRes.GetCode(), gin.H{"error": errorRes})
			return
		}
		var allMenus []bson.M
		if err = result.All(ctx, &allMenus); err != nil {
			errorRes := models.ErrorResponse(http.StatusInternalServerError, "error", "error occured while listing menu items.")
			c.JSON(errorRes.GetCode(), gin.H{"error": errorRes})
			log.Fatal(err)
			return
		}
		successRes := models.SuccessResponse(true, "Success", allMenus)
		c.JSON(http.StatusOK, successRes)
	}
}

func GetMenu() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		menuId := c.Param("menu_id")
		var menu models.Menu
		err := menuCollection.FindOne(ctx, bson.M{"menu_id": menuId}).Decode(&menu)
		defer cancel()
		if err != nil {
			errorRes := models.ErrorResponse(http.StatusInternalServerError, "error", "error occured while fetching the menu item")
			c.JSON(errorRes.GetCode(), gin.H{"error": errorRes})
			return
		}
		successRes := models.SuccessResponse(true, "Success", menu)
		c.JSON(http.StatusOK, successRes)
	}
}

func CreateMenu() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var menu models.Menu

		if err := c.BindJSON(&menu); err != nil {
			errorRes := models.ErrorResponse(http.StatusBadRequest, "error", err.Error())
			c.JSON(errorRes.GetCode(), gin.H{"error": errorRes})
			return
		}

		validationErro := validate.Struct(menu)
		if validationErro != nil {
			errorRes := models.ErrorResponse(http.StatusBadRequest, "error", validationErro.Error())
			c.JSON(errorRes.GetCode(), gin.H{"error": errorRes})
			return
		}
		defer cancel()

		menu.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		menu.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		menu.ID = primitive.NewObjectID()
		menu.Menu_id = menu.ID.Hex()

		_, insetErr := menuCollection.InsertOne(ctx, menu)
		if insetErr != nil {
			errorRes := models.ErrorResponse(http.StatusInternalServerError, "error", "error occured while inserting the menu item")
			c.JSON(errorRes.GetCode(), gin.H{"error": errorRes})
			return
		}

		defer cancel()
		successRes := models.SuccessResponse(true, "Success", menu)
		c.JSON(http.StatusOK, successRes)
	}
}

func UpdateMenu() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var menu models.Menu

		if err := c.BindJSON(&menu); err != nil {
			errorRes := models.ErrorResponse(http.StatusBadRequest, "error", err.Error())
			c.JSON(errorRes.GetCode(), gin.H{"error": errorRes})
			return
		}

		menuId := c.Param("menu_id")
		filter := bson.M{"menu_id": menuId}

		var updateObj primitive.D

		if menu.Start_date != nil && menu.End_date != nil {
			if !inTimeSpan(*menu.Start_date, *menu.End_date, time.Now()) {
				errorRes := models.ErrorResponse(http.StatusBadRequest, "error", "Invalid date range")
				c.JSON(errorRes.GetCode(), gin.H{"error": errorRes})
				return
			}
		}

		updateObj = append(updateObj, bson.E{Key: "start_date", Value: menu.Start_date})
		updateObj = append(updateObj, bson.E{Key: "end_date", Value: menu.End_date})

		if menu.Name != "" {
			updateObj = append(updateObj, bson.E{Key: "name", Value: menu.Name})
		}

		if menu.Category != "" {
			updateObj = append(updateObj, bson.E{Key: "category", Value: menu.Category})
		}

		menu.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

		updateObj = append(updateObj, bson.E{Key: "updated_at", Value: menu.Updated_at})

		upsert := true
		opt := options.UpdateOptions{
			Upsert: &upsert,
		}
		_, err := menuCollection.UpdateOne(
			ctx,
			filter,
			bson.D{
				{Key: "$set", Value: updateObj},
			},
			&opt,
		)
		if err != nil {
			errorRes := models.ErrorResponse(http.StatusInternalServerError, "error", "error occured while updating the menu item")
			c.JSON(errorRes.GetCode(), gin.H{"error": errorRes})
			return
		}

		defer cancel()

		successRes := models.SuccessResponse(true, "Success", updateObj)

		c.JSON(http.StatusOK, successRes)
	}
}

func inTimeSpan(start, end, check time.Time) bool {
	return start.After(time.Now()) && end.After(start)
}

func DeleteMenu() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		// Extract menu_id from the URL path parameters
		menuId := c.Param("menu_id")

		// Define the filter to find the menu by menu_id
		filter := bson.M{"menu_id": menuId}

		// Retrieve the menu item before deleting it
		var menu models.Menu
		err := menuCollection.FindOne(ctx, filter).Decode(&menu)
		if err != nil {
			errorRes := models.ErrorResponse(http.StatusNotFound, "error", "menu not found")
			c.JSON(errorRes.GetCode(), gin.H{"error": errorRes})
			return
		}

		// Attempt to delete the menu item from the collection
		result, err := menuCollection.DeleteOne(ctx, filter)
		if err != nil {
			errorRes := models.ErrorResponse(http.StatusInternalServerError, "error", "error occurred while deleting the menu item")
			c.JSON(errorRes.GetCode(), gin.H{"error": errorRes})
			return
		}

		// Check if the menu item was found and deleted
		if result.DeletedCount == 0 {
			errorRes := models.ErrorResponse(http.StatusNotFound, "error", "menu item not found")
			c.JSON(errorRes.GetCode(), gin.H{"error": errorRes})
			return
		}

		// Delete all food items associated with this menu_id
		foodFilter := bson.M{"menu_id": menuId}
		_, err = foodCollection.DeleteMany(ctx, foodFilter)
		if err != nil {
			errorRes := models.ErrorResponse(http.StatusInternalServerError, "error", "error occurred while deleting associated food items")
			c.JSON(errorRes.GetCode(), gin.H{"error": errorRes})
			return
		}

		// Return a success response with the deleted menu details
		successRes := models.SuccessResponse(true, "Menu and associated food items deleted successfully", menu)
		c.JSON(http.StatusOK, successRes)
	}
}
