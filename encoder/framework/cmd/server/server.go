package main

import (
	"encoder/application/services"
	"encoder/framework/database"
	"encoder/framework/queue"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/streadway/amqp"
)

var db database.Database

func init() {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
	autoMigrateDb, err := strconv.ParseBool(os.Getenv("AUTO_MIGRATE_DB"))
	if err != nil {
		log.Fatalf("Error parsing AUTO_MIGRATE_DB")
	}
	debug, err := strconv.ParseBool(os.Getenv("DEBUG"))
	if err != nil {
		log.Fatalf("Error parsing DEBUG")
	}
	db.AutoMigrateDb = autoMigrateDb
	db.Debug = debug
	db.DsnTest = os.Getenv("DSN_TEST")
	db.Dsn = os.Getenv("DSN")
	db.DbTypeTest = os.Getenv("DB_TYPE_TEST")
	db.DbType = os.Getenv("DB_TYPE")
	db.Env = os.Getenv("ENV")
}

func main() {
	messageChannel := make(chan amqp.Delivery)
	jobReturnChannel := make(chan services.JobWorkerResult)
	dbConnection, err := db.Connect()
	if err != nil {
		log.Fatalf("error connecting to the database", err)
	}
	defer dbConnection.Close()
	rabbitMQ := queue.NewRabbitMQ()
	ch := rabbitMQ.Connect()
	defer ch.Close()
	rabbitMQ.Consume(messageChannel)
	jobManager := services.NewJobManager(dbConnection, messageChannel, jobReturnChannel, rabbitMQ)
	jobManager.Start(ch)

}
