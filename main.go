package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"program/storage"

	"github.com/joho/godotenv"
	_ "gocloud.dev/blob/s3blob"
)

func main() {
	sch := make(chan os.Signal, 1)
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Println("Error during load environments", err)
	}
	// awsstor, err := storage.NewAwsStorage(
	// 	os.Getenv("AWS_REGION"),
	// 	os.Getenv("AWS_ACCESS_KEY_ID"),
	// 	os.Getenv("AWS_SECRET_ACCESS_KEY"),
	// 	"")
	// if err != nil {
	// 	log.Fatal("Error during connect to AWS services", "error", err)
	// }
	rabbitstor, err := storage.NewRabbitStorage(os.Getenv("RABBIT_MQ_URI"))
	if err != nil {
		log.Fatal("Error during connect to RabbitMQ broker", "error", err)
	}
	// msgCh := make(chan *sqs.Message, 100)
	// msgCh := make(chan *amqp.Delivery, 100)

	// go awsstor.GetMsg(msgCh)
	// go rabbitstor.GetMsg()

	go rabbitstor.GetMsg()

	// listen for interrupt and kill signals
	signal.Notify(sch, os.Interrupt, os.Kill)

	// after you've done everything you needed to do:
	<-sch // this is blocking, so after th
}
