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

type InvoiceViewFormat struct {
	Invoice_id       string
	Payment_method   string
	Order_id         string
	Payment_status   *string
	Payment_due      interface{}
	Table_number     interface{}
	Payment_due_date time.Time
	Order_details    interface{}
}

var invoiceCollection *mongo.Collection = database.OpenColletion(database.Client, "invoice")

func GetInvoices() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		result, err := invoiceCollection.Find(context.TODO(), bson.M{})

		defer cancel()
		if err != nil {
			errorRes := models.ErrorResponse(http.StatusInternalServerError, "error", "error occured while fetching the invoices.")
			c.JSON(errorRes.GetCode(), gin.H{"error": errorRes})
			return
		}
		var allInvoices []bson.M
		if err = result.All(ctx, &allInvoices); err != nil {
			errorRes := models.ErrorResponse(http.StatusInternalServerError, "error", "error occured while fetching the invoices.")
			c.JSON(errorRes.GetCode(), gin.H{"error": errorRes})
			log.Fatal(err)
			return
		}

		successRes := models.SuccessResponse(true, "Success", allInvoices)

		c.JSON(http.StatusOK, successRes)
	}
}

func GetInvoice() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		invoiceId := c.Param("invoice_id")
		var invoice models.Invoice
		err := invoiceCollection.FindOne(ctx, bson.M{"invoice_id": invoiceId}).Decode(&invoice)
		defer cancel()
		if err != nil {
			errorRes := models.ErrorResponse(http.StatusInternalServerError, "error", "error occured while fetching the invoice.")
			c.JSON(errorRes.GetCode(), gin.H{"error": errorRes})
			return
		}
		var invoiceView InvoiceViewFormat

		allOrderItems, err := ItemsByOrder(invoice.Order_id)
		invoiceView.Order_id = invoice.Order_id
		invoiceView.Payment_due_date = invoice.Payment_due_date
		invoiceView.Payment_method = "null"
		if invoice.Payment_method != nil {
			invoiceView.Payment_method = *invoice.Payment_method
		}

		invoiceView.Invoice_id = invoice.Invoice_id
		invoiceView.Payment_status = *&invoice.Payment_status
		invoiceView.Payment_due = allOrderItems[0]["payment_due	"]
		invoiceView.Table_number = allOrderItems[0]["table_number"]
		invoiceView.Order_details = allOrderItems[0]["order_items"]

		successRes := models.SuccessResponse(true, "Success", invoiceView)
		c.JSON(http.StatusOK, successRes)

	}
}

func CreateInvoice() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var invoice models.Invoice

		if err := c.BindJSON(&invoice); err != nil {
			errorRes := models.ErrorResponse(http.StatusBadRequest, "error", err.Error())
			c.JSON(errorRes.GetCode(), gin.H{"error": errorRes})
			return
		}

		var order models.Order

		err := orderCollection.FindOne(ctx, bson.M{"order_id": invoice.Order_id}).Decode(&order)
		defer cancel()

		if err != nil {
			errorRes := models.ErrorResponse(http.StatusInternalServerError, "error", "order was not found")
			c.JSON(errorRes.GetCode(), gin.H{"error": errorRes})
			return
		}

		status := "PENDING"
		if invoice.Payment_status == nil {
			invoice.Payment_status = &status
		}
		invoice.Payment_due_date, _ = time.Parse(time.RFC3339, time.Now().AddDate(0, 0, 1).Format(time.RFC3339))
		invoice.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		invoice.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		invoice.ID = primitive.NewObjectID()
		invoice.Invoice_id = invoice.ID.Hex()

		validationErro := validate.Struct(invoice)
		if validationErro != nil {
			errorRes := models.ErrorResponse(http.StatusBadRequest, "error", validationErro.Error())
			c.JSON(errorRes.GetCode(), gin.H{"error": errorRes})
			return
		}

		_, err = invoiceCollection.InsertOne(ctx, invoice)
	}
}

func UpdateInvoice() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var invoice models.Invoice
		invoiceId := c.Param("invoice_id")

		if err := c.BindJSON(&invoice); err != nil {
			errorRes := models.ErrorResponse(http.StatusBadRequest, "error", err.Error())
			c.JSON(errorRes.GetCode(), gin.H{"error": errorRes})
			defer cancel()
			return
		}
		filter := bson.M{"invoice_id": invoiceId}

		var updateObj primitive.D

		if invoice.Payment_method != nil {
			updateObj = append(updateObj, bson.E{Key: "payment_method", Value: invoice.Payment_method})
		}

		if invoice.Payment_status != nil {
			updateObj = append(updateObj, bson.E{Key: "payment_due_date", Value: invoice.Payment_due_date})
		}

		invoice.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

		updateObj = append(updateObj, bson.E{Key: "updated_at", Value: invoice.Updated_at})

		upsert := true
		opt := options.UpdateOptions{
			Upsert: &upsert,
		}

		status := "PENDING"
		if invoice.Payment_status == nil {
			invoice.Payment_status = &status
		}
		_, err := invoiceCollection.UpdateOne(
			ctx,
			filter,
			bson.D{
				{Key: "$set", Value: updateObj},
			},
			&opt,
		)

		if err != nil {
			errorRes := models.ErrorResponse(http.StatusInternalServerError, "error", "error occured while updating the invoice.")
			c.JSON(errorRes.GetCode(), gin.H{"error": errorRes})
			return
		}
		defer cancel()
		successRes := models.SuccessResponse(true, "Success", updateObj)
		c.JSON(http.StatusOK, successRes)
	}
}

func DeleteInvoice() gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}
