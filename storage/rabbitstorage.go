package storage

import (
	"fmt"
	"program/tools"
	"time"

	"github.com/streadway/amqp"
)

type RabbitFs struct {
	rabbitConn    *amqp.Connection
	rabbitChannel *amqp.Channel
	rabbitQueue   amqp.Queue
}

func NewRabbitStorage(url string) (*RabbitFs, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}
	// defer conn.Close()
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	// defer ch.Close()

	q, err := ch.QueueDeclare(
		"jokes-message",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}

	c := &RabbitFs{
		rabbitConn:    conn,
		rabbitChannel: ch,
		rabbitQueue:   q,
	}
	return c, nil
}
func (r *RabbitFs) GetMsg() {
	for {
		msgs, err := r.rabbitChannel.Consume(
			r.rabbitQueue.Name,
			"",
			false,
			false,
			false,
			false,
			nil,
		)

		if err != nil {
			fmt.Printf("failed to fetch  message %v", err)
			return
		}
		// fmt.Println("messages len", len(msgs))

		// if len(msgs) > 0 {
		// for _, msg := range msgs {
		// 	msgCh <- msg

		// }
		// }

		for i := 1; i <= 5; i++ {
			go r.Worker(msgs, i)
		}
		time.Sleep(time.Second)
		// fmt.Printf("no messages in queue\n")

	}

}

func (r *RabbitFs) Worker(msgs <-chan amqp.Delivery, id int) {
	if len(msgs) <= 0 {
		fmt.Printf("no messages in queue\n")
		return
	}
	for msg := range msgs {

		// var j model.Joke

		fmt.Printf("worker %v started a job\n", id)

		_, err := tools.CreateAndSaveMessages(string(msg.Body))
		if err != nil {
			fmt.Println(err)
			return
		}

		// err := json.Unmarshal(msg.Body, &j)
		// if err != nil {
		// 	fmt.Println(err)
		// }

		// err = r.UploadMessageTos3(res)
		// if err != nil {
		// 	return
		// }

		err = msg.Ack(false)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("worker %v finished a job\n", id)

	}
	r.rabbitChannel.Close()

}
