package controllers

import (
	"context"
	"go-restaurants-management/database"
	"go-restaurants-management/models"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var foodCollection *mongo.Collection = database.OpenColletion(database.Client, "food")
var validate = validator.New()

func GetFoods() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)

		recordPerPage, err := strconv.Atoi(c.Query("recordPerPage"))
		if err != nil || recordPerPage < 1 {
			recordPerPage = 10
		}
		page, err := strconv.Atoi(c.Query("page"))

		if err != nil || page < 1 {
			page = 1
		}

		startIndex := (page - 1) * recordPerPage
		startIndex, err = strconv.Atoi(c.Query("startIndex"))

		matchStage := bson.D{
			{Key: "$match", Value: bson.D{{}}},
		}
		groupStage := bson.D{
			{Key: "$group", Value: bson.D{{Key: "_id", Value: bson.D{{Key: "_id", Value: "null"}}}}},
			{Key: "total_count", Value: bson.D{{Key: "$sum", Value: "1"}}},
			{Key: "data", Value: bson.D{{Key: "$push", Value: "$$ROOT"}}},
		}
		projectStage := bson.D{
			{
				Key: "$project",
				Value: bson.D{
					{Key: "_id", Value: 0},
					{Key: "total_count", Value: 1},
					{Key: "food_items", Value: bson.D{{Key: "$slice", Value: []interface{}{"$data", startIndex, recordPerPage}}}},
				},
			},
		}

		result, err := foodCollection.Aggregate(ctx, mongo.Pipeline{
			matchStage,
			groupStage,
			projectStage,
		})
		defer cancel()
		if err != nil {
			errorRes := models.ErrorResponse(http.StatusInternalServerError, "error", "error occured while fetching the food items")
			c.JSON(errorRes.GetCode(), gin.H{"error": errorRes})
			return
		}

		var allFoods []bson.M

		if err = result.All(ctx, &allFoods); err != nil {
			errorRes := models.ErrorResponse(http.StatusInternalServerError, "error", "error occured while fetching the food items")
			c.JSON(errorRes.GetCode(), gin.H{"error": errorRes})
			log.Fatal(err)
			return
		}

		successRes := models.SuccessResponse(true, "Success", allFoods)

		c.JSON(http.StatusOK, successRes)
	}
}

func GetFood() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		foodId := c.Param("food_id")
		var food models.Food
		err := foodCollection.FindOne(ctx, bson.M{"food_id": foodId}).Decode(&food)
		defer cancel()
		if err != nil {
			errorRes := models.ErrorResponse(http.StatusInternalServerError, "error", "error occured while fetching the food item")
			c.JSON(errorRes.GetCode(), gin.H{"error": errorRes})
			return
		}
		successRes := models.SuccessResponse(true, "Success", food)
		c.JSON(http.StatusOK, successRes)
	}
}

func CreateFood() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var menu models.Menu
		var food models.Food

		if err := c.BindJSON(&food); err != nil {
			errorRes := models.ErrorResponse(http.StatusBadRequest, "error", err.Error())
			c.JSON(errorRes.GetCode(), gin.H{"error": errorRes})
			return
		}

		validationErr := validate.Struct(food)
		if validationErr != nil {
			errorRes := models.ErrorResponse(http.StatusBadRequest, "error", validationErr.Error())
			c.JSON(errorRes.GetCode(), gin.H{"error": errorRes})
			return
		}

		err := menuCollection.FindOne(ctx, bson.M{"menu_id": food.Menu_id}).Decode(&menu)
		defer cancel()
		if err != nil {
			errorRes := models.ErrorResponse(http.StatusInternalServerError, "error", "menu not found")
			c.JSON(errorRes.GetCode(), gin.H{"error": errorRes})
			return
		}
		food.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		food.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		food.ID = primitive.NewObjectID()
		food.Food_id = food.ID.Hex()
		var num = toFixed(*food.Price, 2)
		food.Price = &num

		_, insetErr := foodCollection.InsertOne(ctx, food)
		if insetErr != nil {
			errorRes := models.ErrorResponse(http.StatusInternalServerError, "error", "error occured while inserting the food item")
			c.JSON(errorRes.GetCode(), gin.H{"error": errorRes})
			return
		}

		defer cancel()
		successRes := models.SuccessResponse(true, "Success", food)
		c.JSON(http.StatusOK, successRes)
	}
}

func round(num float64) int {
	return int(num + math.Copysign(0.5, num))
}

func toFixed(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(round(num*output)) / output
}

func UpdateFood() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var menu models.Menu
		var food models.Food
		foodId := c.Param("food_id")

		if err := c.BindJSON(&food); err != nil {
			errorRes := models.ErrorResponse(http.StatusBadRequest, "error", err.Error())
			c.JSON(errorRes.GetCode(), gin.H{"error": errorRes})
			return
		}

		var updateObj primitive.D

		if food.Name != nil {
			updateObj = append(updateObj, bson.E{Key: "name", Value: food.Name})
		}

		if food.Price != nil {
			updateObj = append(updateObj, bson.E{Key: "price", Value: food.Price})
		}

		if food.Food_image != nil {
			updateObj = append(updateObj, bson.E{Key: "food_image", Value: food.Food_image})
		}

		if food.Menu_id != nil {
			err := menuCollection.FindOne(ctx, bson.M{"menu_id": food.Menu_id}).Decode(&menu)
			defer cancel()
			if err != nil {
				errorRes := models.ErrorResponse(http.StatusInternalServerError, "error", "menu not found")
				c.JSON(errorRes.GetCode(), gin.H{"error": errorRes})
				return
			}
			updateObj = append(updateObj, bson.E{Key: "menu_id", Value: food.Menu_id})
		}

		food.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		updateObj = append(updateObj, bson.E{Key: "updated_at", Value: food.Updated_at})
		upsert := true
		filter := bson.M{"food_id": foodId}
		opt := options.UpdateOptions{
			Upsert: &upsert,
		}

		_, err := foodCollection.UpdateOne(
			ctx,
			filter,
			bson.D{{Key: "$set", Value: updateObj}},
			&opt,
		)

		if err != nil {
			errorRes := models.ErrorResponse(http.StatusInternalServerError, "error", "error occured while updating the food item")
			c.JSON(errorRes.GetCode(), gin.H{"error": errorRes})
			return
		}

		defer cancel()

		successRes := models.SuccessResponse(true, "Success", updateObj)

		c.JSON(http.StatusOK, successRes)
	}
}

func DeleteFood() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		// Extract food_id from the URL path parameters
		foodId := c.Param("food_id")

		// Define the filter to find the food by food_id
		filter := bson.M{"food_id": foodId}

		// Retrieve the food item before deleting it
		var food models.Food
		err := foodCollection.FindOne(ctx, filter).Decode(&food)
		if err != nil {
			errorRes := models.ErrorResponse(http.StatusNotFound, "error", "food not found")
			c.JSON(errorRes.GetCode(), gin.H{"error": errorRes})
			return
		}

		// Attempt to delete the food item from the collection
		result, err := foodCollection.DeleteOne(ctx, filter)
		if err != nil {
			errorRes := models.ErrorResponse(http.StatusInternalServerError, "error", "error occurred while deleting the food item")
			c.JSON(errorRes.GetCode(), gin.H{"error": errorRes})
			return
		}

		// Check if the food item was found and deleted
		if result.DeletedCount == 0 {
			errorRes := models.ErrorResponse(http.StatusNotFound, "error", "food item not found")
			c.JSON(errorRes.GetCode(), gin.H{"error": errorRes})
			return
		}

		// Return a success response with the deleted food details
		successRes := models.SuccessResponse(true, "Food item deleted successfully", food)
		c.JSON(http.StatusOK, successRes)
	}
}

func GetMostOrderedFood() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		// Aggregation pipeline to get top 5 most ordered food items
		pipeline := mongo.Pipeline{
			// Unwind the items array in each order (assuming orders contain an array of items)
			{{Key: "$unwind", Value: bson.D{{Key: "path", Value: "$items"}}}},

			// Group by food_id and count the number of orders for each food item
			{{Key: "$group", Value: bson.D{
				{Key: "_id", Value: "$items.food_id"},
				{Key: "totalOrders", Value: bson.D{{Key: "$sum", Value: 1}}},
			}}},

			// Sort by totalOrders in descending order to get the most ordered food items
			{{Key: "$sort", Value: bson.D{{Key: "totalOrders", Value: -1}}}},

			// Limit to top 5 most ordered food items
			{{Key: "$limit", Value: 5}},
		}

		// Execute the aggregation pipeline
		cursor, err := orderCollection.Aggregate(ctx, pipeline)
		if err != nil {
			errorRes := models.ErrorResponse(http.StatusInternalServerError, "error", "error occurred while fetching the most ordered food items")
			c.JSON(errorRes.GetCode(), gin.H{"error": errorRes})
			return
		}

		// Parse the result into a slice
		var mostOrderedFood []bson.M
		if err := cursor.All(ctx, &mostOrderedFood); err != nil {
			errorRes := models.ErrorResponse(http.StatusInternalServerError, "error", "error occurred while decoding the result")
			c.JSON(errorRes.GetCode(), gin.H{"error": errorRes})
			return
		}

		// Check if any food items are found
		if len(mostOrderedFood) == 0 {
			errorRes := models.ErrorResponse(http.StatusNotFound, "error", "no orders found")
			c.JSON(errorRes.GetCode(), gin.H{"error": errorRes})
			return
		}

		// Return the top 5 most ordered food items
		successRes := models.SuccessResponse(true, "Top 5 most ordered food items retrieved successfully", mostOrderedFood)
		c.JSON(http.StatusOK, successRes)
	}
}
