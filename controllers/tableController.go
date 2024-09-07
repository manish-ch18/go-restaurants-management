package controllers

import (
	"go-restaurants-management/database"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

var tableCollection *mongo.Collection = database.OpenColletion(database.Client, "table")

func GetTables() gin.HandlerFunc {
	return func(ctx *gin.Context) {

	}
}

func GetTable() gin.HandlerFunc {
	return func(ctx *gin.Context) {

	}
}

func CreateTable() gin.HandlerFunc {
	return func(ctx *gin.Context) {

	}
}

func UpdateTable() gin.HandlerFunc {
	return func(ctx *gin.Context) {

	}
}
