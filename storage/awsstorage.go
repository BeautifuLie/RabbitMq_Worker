package storage

import (
	"fmt"
	"os"
	"program/tools"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/sqs"
)

type AWSFs struct {
	awss3       *s3.S3
	sqsSess     *sqs.SQS
	bucketRead  string
	bucketWrite string
	bucketSQS   string
	urlQueue    string
}

func NewAwsStorage(region, id, secret, token string) (*AWSFs, error) {

	session, err := session.NewSession(
		&aws.Config{
			Region: &region,
			Credentials: credentials.NewStaticCredentials(
				id,
				secret,
				token,
			),
		})

	if err != nil {
		return nil, err
	}
	awss3 := s3.New(session)
	sqsSess := sqs.New(session)
	conn := &AWSFs{
		awss3:       awss3,
		sqsSess:     sqsSess,
		bucketRead:  os.Getenv("BUCKET_READ_NAME"),
		bucketWrite: os.Getenv("BUCKET_WRITE_NAME"),
		bucketSQS:   *aws.String("jokes-sqs-messages"),
		urlQueue:    *aws.String("https://sqs.eu-central-1.amazonaws.com/333746971525/JokesQueueSend"),
	}
	return conn, nil
}
func (a *AWSFs) GetMsg(msgCh chan *sqs.Message) {
	for {
		msgResult, err := a.sqsSess.ReceiveMessage(&sqs.ReceiveMessageInput{
			QueueUrl:            &a.urlQueue,
			MaxNumberOfMessages: aws.Int64(10),
			WaitTimeSeconds:     aws.Int64(10),
		})

		if err != nil {
			fmt.Printf("failed to fetch sqs message %v", err)
			continue
		}
		fmt.Println("messages len", len(msgResult.Messages))

		if len(msgResult.Messages) > 0 {
			for _, msg := range msgResult.Messages {
				msgCh <- msg

			}
		}

		fmt.Printf("no messages in queue\n")

	}

}

func (a *AWSFs) Worker(msgCh chan *sqs.Message, id int) {

	for msg := range msgCh {

		fmt.Printf("worker %v started a job\n", id)

		res, err := tools.CreateAndSaveMessages(*msg.Body)
		if err != nil {
			fmt.Println(err)
			return
		}

		err = a.UploadMessageTos3(res)
		if err != nil {
			return
		}
		err = a.DeleteMsg(*msg.ReceiptHandle)
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Printf("worker %v finished a job\n", id)

	}

}
func (a *AWSFs) DeleteMsg(messageHandle string) error {

	_, err := a.sqsSess.DeleteMessage(&sqs.DeleteMessageInput{
		QueueUrl:      &a.urlQueue,
		ReceiptHandle: &messageHandle,
	})

	if err != nil {
		return err
	}
	return nil
}
func (a *AWSFs) UploadMessageTos3(filename string) error {

	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	uploader := s3manager.NewUploaderWithClient(a.awss3)
	upParams := &s3manager.UploadInput{
		Bucket: &a.bucketSQS,
		Key:    &filename,
		Body:   f,
	}
	_, err = uploader.Upload(upParams, func(u *s3manager.Uploader) {
		u.LeavePartsOnError = true
	})
	if err != nil {
		return err
	}

	return nil
}
func (a *AWSFs) GetQueueUrl(queueName string) (string, error) {

	result, err := a.sqsSess.GetQueueUrl(&sqs.GetQueueUrlInput{
		QueueName: &queueName,
	})

	if err != nil {
		return "", err
	}

	return *result.QueueUrl, nil
}
