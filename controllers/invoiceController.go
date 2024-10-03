package controller

import (
	"context"
	"fmt"
	"golang-restaurant-backend-app/database"
	"golang-restaurant-backend-app/models"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type InvoiceViewFormat struct {
	Invoice_id       string      `json:"invoice_id"`
	Payment_method   string      `json:"payment_method"`
	Order_id         string      `json:"order_id"`
	Payment_status   *string     `json:"payment_status"`
	Payment_due      interface{} `json:"payment_due"`
	Table_number     interface{} `json:"table_number"`
	Payment_due_date time.Time   `json:"payment_due_date"`
	Order_details    interface{} `json:"order_details"`
}

var invoiceCollection *mongo.Collection = database.OpenCollection(database.Client, "invoice")

func GetInvoices() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		result, err := invoiceCollection.Find(context.TODO(), bson.M{})

		defer cancel()

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while listing the invoice items"})
		}

		var allInvoices []bson.M
		if err = result.All(ctx, &allInvoices); err != nil {
			log.Fatal(err)
		}
		c.JSON(http.StatusOK, allInvoices)
	}
}

func GetInvoice() gin.HandlerFunc {
	return func(c *gin.Context) {

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		invoiceId := c.Param("invoice_id")

		var invoice models.Invoice

		err := invoiceCollection.FindOne(ctx, bson.M{"order_id": invoiceId}).Decode(&invoice)

		defer cancel()

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while listing the invoice items"})
			return
		}

		var invoiceView InvoiceViewFormat

		allOrderItems, err := ItemsByOrder(invoice.Order_id)
		invoiceView.Order_id = invoice.Order_id
		invoiceView.Payment_due_date = invoice.Payment_due_date

		invoiceView.Payment_method = "null"
		if invoice.Payment_method != nil {
			invoiceView.Payment_method = *invoice.Payment_method
		} else {
			invoiceView.Payment_method = "null"
		}
		invoiceView.Invoice_id = invoice.Invoice_id
		invoiceView.Payment_status = invoice.Payment_status
		if len(allOrderItems) > 0 {
			invoiceView.Payment_due = allOrderItems[0]["payment_due"]
			invoiceView.Table_number = allOrderItems[0]["table_number"]
			invoiceView.Order_details = allOrderItems[0]["order_items"]
		} else {
			invoiceView.Payment_due = nil
			invoiceView.Table_number = nil
			invoiceView.Order_details = nil
		}

		c.JSON(http.StatusOK, invoiceView)
	}
}

func CreateInvoice() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		var invoice models.Invoice

		invoiceId := c.Param("invoice_id")

		if err := c.BindJSON(&invoice); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		filter := bson.M{"invoice_id": invoiceId}

		var updateObj Primitive.D

		if invoice.Payment_method != nil {

		}
		if invoice.Payment_status != nil {

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

		result, err := invoiceCollection.UpdateOne(
			ctx,
			filter,
			bson.D{
				{Key: "$set", Value: updateObj},
			},
			&opt,
		)

		if err != nil {
			msg := fmt.Sprint("Invoice item failed to update")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
		}
		c.JSON(http.StatusOK, result)

	}
}

func UpdateInvoice() gin.HandlerFunc {
	return func(c *gin.Context) {
	}
}
