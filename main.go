package main

import (
	"log"

	"gin-billing/handler"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func main() {
	r := gin.Default()
	r.Use(handler.ErrorHandle())
	midtrans := r.Group("/midtrans")
	{
		midtransController := handler.NewMidtransControllerImpl(validator.New())
		midtrans.POST("/create", midtransController.Create)
	}

	r.Run()
}
