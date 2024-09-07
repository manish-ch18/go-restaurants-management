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

var orderCollection *mongo.Collection = database.OpenColletion(database.Client, "order")

func GetOrders() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)

		result, err := orderCollection.Find(context.TODO(), bson.M{})

		defer cancel()
		if err != nil {
			errorRes := models.ErrorResponse(http.StatusInternalServerError, "error", "error occured while fetching the orders.")
			c.JSON(errorRes.GetCode(), gin.H{"error": errorRes})
			return
		}
		var allOrders []bson.M

		if err = result.All(ctx, &allOrders); err != nil {
			errorRes := models.ErrorResponse(http.StatusInternalServerError, "error", "error occured while fetching the orders.")
			c.JSON(errorRes.GetCode(), gin.H{"error": errorRes})
			log.Fatal(err)
			return
		}

		successRes := models.SuccessResponse(true, "Success", allOrders)

		c.JSON(http.StatusOK, successRes)
	}
}

func GetOrder() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		orderId := c.Param("order_id")
		var order models.Order
		err := orderCollection.FindOne(ctx, bson.M{"order_id": orderId}).Decode(&order)
		defer cancel()
		if err != nil {
			errorRes := models.ErrorResponse(http.StatusInternalServerError, "error", "error occured while fetching the order item")
			c.JSON(errorRes.GetCode(), gin.H{"error": errorRes})
			return
		}
		successRes := models.SuccessResponse(true, "Success", order)
		c.JSON(http.StatusOK, successRes)
	}
}

func CreateOrder() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		var table models.Table
		var order models.Order

		if err := c.BindJSON(&order); err != nil {
			errorRes := models.ErrorResponse(http.StatusBadRequest, "error", err.Error())
			c.JSON(errorRes.GetCode(), gin.H{"error": errorRes})
			return
		}

		validationErr := validate.Struct(order)

		if validationErr != nil {
			errorRes := models.ErrorResponse(http.StatusBadRequest, "error", validationErr.Error())
			c.JSON(errorRes.GetCode(), gin.H{"error": errorRes})
			return
		}

		if order.Table_id != nil {
			err := tableCollection.FindOne(ctx, bson.M{"table_id": order.Table_id}).Decode(&table)
			defer cancel()
			if err != nil {
				errorRes := models.ErrorResponse(http.StatusInternalServerError, "error", "table was not found")
				c.JSON(errorRes.GetCode(), gin.H{"error": errorRes})
				return
			}
		}

		order.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		order.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		order.ID = primitive.NewObjectID()
		order.Order_id = order.ID.Hex()
		_, insertErr := orderCollection.InsertOne(ctx, order)

		if insertErr != nil {
			errorRes := models.ErrorResponse(http.StatusInternalServerError, "error", "error occured while inserting the order item")
			c.JSON(errorRes.GetCode(), gin.H{"error": errorRes})
			return
		}

		defer cancel()
		successRes := models.SuccessResponse(true, "Success", order)
		c.JSON(http.StatusOK, successRes)
	}
}

func UpdateOrder() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var table models.Table
		var order models.Order
		var updateObj primitive.D
		orderId := c.Param("order_id")

		if err := c.BindJSON(&order); err != nil {
			errorRes := models.ErrorResponse(http.StatusBadRequest, "error", err.Error())
			c.JSON(errorRes.GetCode(), gin.H{"error": errorRes})
			return
		}

		if order.Table_id != nil {
			err := menuCollection.FindOne(ctx, bson.M{"table_id": order.Table_id}).Decode(&table)
			defer cancel()
			if err != nil {
				errorRes := models.ErrorResponse(http.StatusInternalServerError, "error", "table was not found")
				c.JSON(errorRes.GetCode(), gin.H{"error": errorRes})
				return
			}
			updateObj = append(updateObj, bson.E{Key: "table_id", Value: order.Table_id})
		}

		order.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		updateObj = append(updateObj, bson.E{Key: "updated_at", Value: order.Updated_at})
		upsert := true
		filter := bson.M{"order_id": orderId}
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
			errorRes := models.ErrorResponse(http.StatusInternalServerError, "error", "error occured while updating the order item")
			c.JSON(errorRes.GetCode(), gin.H{"error": errorRes})
			return
		}

		defer cancel()

		successRes := models.SuccessResponse(true, "Success", updateObj)

		c.JSON(http.StatusOK, successRes)
	}
}

func DeleteOrder() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		// Extract order_id from the URL path parameters
		orderId := c.Param("order_id")

		// Define the filter to find the order by order_id
		filter := bson.M{"order_id": orderId}

		// Retrieve the order before deleting it
		var order models.Order
		err := orderCollection.FindOne(ctx, filter).Decode(&order)
		if err != nil {
			errorRes := models.ErrorResponse(http.StatusNotFound, "error", "order not found")
			c.JSON(errorRes.GetCode(), gin.H{"error": errorRes})
			return
		}

		// Attempt to delete the order from the collection
		result, err := orderCollection.DeleteOne(ctx, filter)
		if err != nil {
			errorRes := models.ErrorResponse(http.StatusInternalServerError, "error", "error occurred while deleting the order")
			c.JSON(errorRes.GetCode(), gin.H{"error": errorRes})
			return
		}

		// Check if the order was found and deleted
		if result.DeletedCount == 0 {
			errorRes := models.ErrorResponse(http.StatusNotFound, "error", "order not found")
			c.JSON(errorRes.GetCode(), gin.H{"error": errorRes})
			return
		}

		// Return a success response with the deleted order details
		successRes := models.SuccessResponse(true, "Order deleted successfully", order)
		c.JSON(http.StatusOK, successRes)
	}
}

func OrderItemOrderCreator(order models.Order) string {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	order.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	order.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	order.ID = primitive.NewObjectID()
	order.Order_id = order.ID.Hex()
	orderCollection.InsertOne(ctx, order)
	defer cancel()
	return order.Order_id
}
