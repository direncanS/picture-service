package main

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"io"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/lambda-lama/picture-service/utils"
)

type S3Bucket struct {
	S3Client *s3.Client
}

func NewS3Bucket(client *s3.Client) *S3Bucket {
	return &S3Bucket{S3Client: client}
}

func (client S3Bucket) uploadImage(bucketName string, bucketKey string, imageBytes []byte) error {
	_, err := client.S3Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String(bucketName),
		Key:         aws.String(bucketKey),
		Body:        bytes.NewReader(imageBytes),
		ContentType: aws.String("image/jpeg"),
	})

	if err != nil {
		fmt.Println("Error uploading file: ", err)
		return err
	}

	return nil
}

func (client S3Bucket) getImage(bucketName string, objectKey string) (image.Image, error) {
	result, err := client.S3Client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		fmt.Printf("Couldn't get object %v:%v: %v\n", bucketName, objectKey, err)
		return nil, err
	}
	defer result.Body.Close()

	imageData, err := io.ReadAll(result.Body)
	if err != nil {
		fmt.Printf("Couldn't read object body from %v: %v\n", objectKey, err)
		return nil, err
	}

	img, _, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		fmt.Println("Error decoding image:", err)
		return nil, err
	}

	return img, nil
}

func handler(ctx context.Context, sqsEvent events.SQSEvent) error {
	config, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-1"))
	if err != nil {
		fmt.Println("Error while loading the aws config ", err)
		return err
	}

	client := s3.NewFromConfig(config)
	s3Bucket := NewS3Bucket(client)

	for _, message := range sqsEvent.Records {
		bucketName, found := message.MessageAttributes["bucket_name"]
		if !found {
			fmt.Println("no bucket name found")
			continue
		}

		imageObjectKey, found := message.MessageAttributes["image_object_key"]
		if !found {
			fmt.Println("no object key found")
			continue
		}

		img, err := s3Bucket.getImage(*bucketName.StringValue, *imageObjectKey.StringValue)
		if err != nil {
			fmt.Println("error getting image: ", err.Error())
			continue
		}

		bImg, err := utils.CompressImage(img)
		if err != nil {
			fmt.Println("erorr compressing image: ", err.Error())
		}

		s3Bucket.uploadImage(*bucketName.StringValue, *imageObjectKey.StringValue, bImg)

		fmt.Printf("The message %s for event source %s = %s \n", message.MessageId, message.EventSource, message.Body)
	}

	return nil
}

func main() {
	lambda.Start(handler)
}
