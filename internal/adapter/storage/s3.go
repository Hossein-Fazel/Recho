package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/Hossein-Fazel/Recho/internal/apperr"
	"github.com/Hossein-Fazel/Recho/internal/application"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager"
	tmtypes "github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
)

const defaultPresignExpiry = 15 * time.Minute

// S3Config configures an S3 or S3-compatible (MinIO, R2, Spaces, ...) bucket.
type S3Config struct {
	Endpoint        string        `env:"ENDPOINT"`
	Region          string        `env:"REGION" envDefault:"us-east-1"`
	Bucket          string        `env:"BUCKET"`
	AccessKeyID     string        `env:"ACCESS_KEY_ID"`
	SecretAccessKey string        `env:"SECRET_ACCESS_KEY"`
	UsePathStyle    bool          `env:"USE_PATH_STYLE" envDefault:"true"`
	DisableACL      bool          `env:"DISABLE_ACL" envDefault:"false"`
	PublicBaseURL   string        `env:"PUBLIC_BASE_URL"`
	PresignExpiry   time.Duration `env:"PRESIGN_EXPIRY" envDefault:"15m"`
}


type S3 struct {
	cfg      S3Config
	client   *s3.Client
	transfer *transfermanager.Client
	presign  *s3.PresignClient
}

var _ application.Storage = (*S3)(nil)

func NewS3(ctx context.Context, cfg S3Config) (application.Storage, error) {
	if strings.TrimSpace(cfg.Bucket) == "" {
		return nil, apperr.InvalidInput("storage", "s3 bucket is required", nil)
	}

	if cfg.PresignExpiry <= 0 {
		cfg.PresignExpiry = defaultPresignExpiry
	}

	loadOptions := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(cfg.Region),
	}

	if cfg.AccessKeyID != "" && cfg.SecretAccessKey != "" {
		loadOptions = append(loadOptions, awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		))
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, loadOptions...)
	if err != nil {
		return nil, apperr.InvalidInput("storage", "failed to load s3 configuration", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		if cfg.Endpoint != "" {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
		}
		o.UsePathStyle = cfg.UsePathStyle
	})

	return &S3{
		cfg:      cfg,
		client:   client,
		transfer: transfermanager.New(client),
		presign:  s3.NewPresignClient(client),
	}, nil
}

func (a *S3) Upload(ctx context.Context, file io.Reader, size int64, key, contentType string, public bool) (string, error) {
	input := &transfermanager.UploadObjectInput{
		Bucket: aws.String(a.cfg.Bucket),
		Key:    aws.String(key),
		Body:   file,
	}

	if size > 0 {
		input.ContentLength = aws.Int64(size)
	}

	if contentType != "" {
		input.ContentType = aws.String(contentType)
	}

	if public && !a.cfg.DisableACL {
		input.ACL = tmtypes.ObjectCannedACLPublicRead
	}

	if _, err := a.transfer.UploadObject(ctx, input); err != nil {
		return "", apperr.Internal("storage", err)
	}

	return key, nil
}

func (a *S3) Delete(ctx context.Context, key string) error {
	_, err := a.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(a.cfg.Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return apperr.Internal("storage", err)
	}

	return nil
}

func (a *S3) GetURL(_ context.Context, key string) (string, error) {
	if a.cfg.PublicBaseURL != "" {
		return fmt.Sprintf("%s/%s", strings.TrimRight(a.cfg.PublicBaseURL, "/"), strings.TrimLeft(key, "/")), nil
	}

	if a.cfg.Endpoint != "" {
		return fmt.Sprintf("%s/%s/%s", strings.TrimRight(a.cfg.Endpoint, "/"), a.cfg.Bucket, strings.TrimLeft(key, "/")), nil
	}

	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", a.cfg.Bucket, a.cfg.Region, strings.TrimLeft(key, "/")), nil
}

func (a *S3) GetPresignedURL(ctx context.Context, key string) (string, error) {
	req, err := a.presign.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(a.cfg.Bucket),
		Key:    aws.String(key),
	}, func(o *s3.PresignOptions) {
		o.Expires = a.cfg.PresignExpiry
	})
	if err != nil {
		return "", apperr.Internal("storage", err)
	}

	return req.URL, nil
}

func (a *S3) Exists(ctx context.Context, key string) (bool, error) {
	_, err := a.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(a.cfg.Bucket),
		Key:    aws.String(key),
	})
	if err == nil {
		return true, nil
	}

	if isNotFound(err) {
		return false, nil
	}

	return false, apperr.Internal("storage", err)
}

func isNotFound(err error) bool {
	var notFound *s3types.NotFound
	if errors.As(err, &notFound) {
		return true
	}

	var noSuchKey *s3types.NoSuchKey
	if errors.As(err, &noSuchKey) {
		return true
	}

	var apiErr smithy.APIError
	if errors.As(err, &apiErr) {
		switch apiErr.ErrorCode() {
		case "NotFound", "NoSuchKey", "404":
			return true
		}
	}

	return false
}
