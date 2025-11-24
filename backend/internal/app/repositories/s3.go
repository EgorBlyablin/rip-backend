package repositories

import (
	"context"
	"fmt"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	log "github.com/sirupsen/logrus"
)

type S3Repository struct {
	s3     *minio.Client
	bucket string
}

func NewS3Repository(host string, port int, bucket string) (*S3Repository, error) {
	client, err := minio.New(fmt.Sprintf("%s:%d", host, port), &minio.Options{
		Creds: credentials.NewStaticV4("minioadmin", "minioadmin", ""),
	})
	if err != nil {
		return nil, err
	}

	bucketExists, err := client.BucketExists(context.Background(), bucket)
	if err != nil {
		return nil, err
	}
	if !bucketExists {
		log.Errorf("Bucket %s does not exist, creating it", bucket)
		if err := client.MakeBucket(context.Background(), bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, err
		}
	}

	return &S3Repository{s3: client, bucket: bucket}, nil
}

func (r *S3Repository) UploadFile(key string, body io.Reader, imageSize int64, contentType string) (string, error) {
	uploadInfo, err := r.s3.PutObject(context.Background(), r.bucket, key, body, imageSize, minio.PutObjectOptions{
		ContentType: contentType,
	})

	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"Bucket": r.bucket,
			"Key":    key,
		}).Error("Error occurred while uploading file to S3")
		return "", err
	}
	return uploadInfo.Key, nil
}

func (r *S3Repository) DeleteFile(key string) error {
	if err := r.s3.RemoveObject(context.Background(), r.bucket, key, minio.RemoveObjectOptions{}); err != nil {
		log.WithError(err).WithFields(log.Fields{
			"Bucket": r.bucket,
			"Key":    key,
		}).Error("Error occurred while deleting file from S3")
		return err
	}
	return nil
}
