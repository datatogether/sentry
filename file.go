package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"io"
	"net/http"
)

type File struct {
	Url  string
	Data *bytes.Buffer
	Hash string
}

func NewFileFromRes(url string, res *http.Response) (*File, error) {
	f := &File{
		Url:  url,
		Data: &bytes.Buffer{},
	}

	if _, err := io.Copy(f.Data, res.Body); err != nil {
		return nil, err
	}

	f.CalcHash()
	return f, res.Body.Close()
}

func (f *File) CalcHash() {
	h := sha256.New()
	h.Write(f.Data.Bytes())
	f.Hash = base64.URLEncoding.EncodeToString(h.Sum(nil))
}

func (f *File) PutS3() error {
	if f.Data == nil {
		return fmt.Errorf("no data for saving url to s3: '%s'", f.Url)
	}

	if f.Hash == "" {
		f.CalcHash()
	}

	svc := s3.New(session.New(&aws.Config{
		Region:      aws.String(cfg.AwsRegion),
		Credentials: credentials.NewStaticCredentials(cfg.AwsAccessKeyId, cfg.AwsSecretAccessKey, ""),
	}))

	_, err := svc.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(cfg.AwsS3BucketName),
		Key:    aws.String(f.s3Path(f.Hash)),
		Body:   bytes.NewReader(f.Data.Bytes()),
	})

	return err
}

func (f *File) s3Path(hash string) string {
	return cfg.AwsS3BucketPath + "/" + hash
}
