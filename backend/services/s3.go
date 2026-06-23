package services

import (
	"context"
	"fmt"
	"io"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	appConfig "solicitud-permisos/internal/config"
	"solicitud-permisos/utils"
)

type S3Service struct {
	client *s3.Client
	bucket string
	region string
}

var S3 *S3Service

func InitS3(cfg appConfig.S3Config) error {
	if cfg.AccessKeyID == "" || cfg.SecretAccessKey == "" || cfg.BucketName == "" || cfg.Region == "" {
		log.Println("S3 no configurado, carga de archivos deshabilitada")
		return nil
	}

	awsCfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(cfg.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.AccessKeyID,
			cfg.SecretAccessKey,
			"",
		)),
	)
	if err != nil {
		return fmt.Errorf("error configurando S3: %w", err)
	}

	S3 = &S3Service{
		client: s3.NewFromConfig(awsCfg, func(o *s3.Options) {
			o.UsePathStyle = true
		}),
		bucket: cfg.BucketName,
		region: cfg.Region,
	}

	log.Println("Servicio S3 inicializado correctamente")
	return nil
}

func (s *S3Service) UploadFile(ctx context.Context, file io.Reader, originalFilename string, cedula string) (string, error) {
	ext := strings.ToLower(filepath.Ext(originalFilename))
	if ext == "" {
		ext = ".bin"
	}

	key := fmt.Sprintf("permisos/%s/%d_%s%s",
		cedula,
		time.Now().UnixMilli(),
		utils.GenerateRandomString(8),
		ext,
	)

	contentType := getContentType(ext)

	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        file,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", fmt.Errorf("error subiendo archivo a S3: %w", err)
	}

	url := fmt.Sprintf("https://s3.%s.amazonaws.com/%s/%s", s.region, s.bucket, key)
	return url, nil
}

func (s *S3Service) UploadAnuncioFile(ctx context.Context, file io.Reader, originalFilename string, cedula string) (string, error) {
	ext := strings.ToLower(filepath.Ext(originalFilename))
	if ext == "" {
		ext = ".bin"
	}

	key := fmt.Sprintf("anuncios/%s/%d_%s%s",
		cedula,
		time.Now().UnixMilli(),
		utils.GenerateRandomString(8),
		ext,
	)

	contentType := getContentType(ext)

	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        file,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", fmt.Errorf("error subiendo archivo a S3: %w", err)
	}

	url := fmt.Sprintf("https://s3.%s.amazonaws.com/%s/%s", s.region, s.bucket, key)
	return url, nil
}

func (s *S3Service) IsEnabled() bool {
	return s != nil && s.client != nil
}

func getContentType(ext string) string {
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".pdf":
		return "application/pdf"
	default:
		return "application/octet-stream"
	}
}

func (s *S3Service) GeneratePresignedURL(key string, expiration time.Duration) (string, error) {
	if !s.IsEnabled() {
		return "", fmt.Errorf("servicio S3 no está habilitado")
	}

	presignClient := s3.NewPresignClient(s.client)
	req, err := presignClient.PresignGetObject(context.Background(), &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(expiration))
	if err != nil {
		return "", fmt.Errorf("error generando presigned URL: %w", err)
	}
	return req.URL, nil
}

func (s *S3Service) GetObject(ctx context.Context, key string) (io.ReadCloser, string, error) {
	if !s.IsEnabled() {
		return nil, "", fmt.Errorf("servicio S3 no está habilitado")
	}

	out, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, "", err
	}

	contentType := "application/octet-stream"
	if out.ContentType != nil {
		contentType = *out.ContentType
	}

	return out.Body, contentType, nil
}

func (s *S3Service) GeneratePresignedURLFromS3URL(s3Url string, expiration time.Duration) (string, error) {
	if !s.IsEnabled() {
		return "", fmt.Errorf("servicio S3 no está habilitado")
	}

	prefix := fmt.Sprintf("https://s3.%s.amazonaws.com/%s/", s.region, s.bucket)
	if !strings.HasPrefix(s3Url, prefix) {
		return "", fmt.Errorf("formato de URL S3 inválido")
	}

	key := strings.TrimPrefix(s3Url, prefix)
	return s.GeneratePresignedURL(key, expiration)
}

func (s *S3Service) ExtractKeyFromURL(s3Url string) (string, error) {
	if !s.IsEnabled() {
		return "", fmt.Errorf("servicio S3 no está habilitado")
	}

	prefix := fmt.Sprintf("https://s3.%s.amazonaws.com/%s/", s.region, s.bucket)
	if !strings.HasPrefix(s3Url, prefix) {
		return "", fmt.Errorf("formato de URL S3 inválido")
	}

	return strings.TrimPrefix(s3Url, prefix), nil
}

func (s *S3Service) FileExists(ctx context.Context, key string) (bool, error) {
	if !s.IsEnabled() {
		return false, fmt.Errorf("servicio S3 no está habilitado")
	}

	_, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		errStr := err.Error()
		if strings.Contains(errStr, "NotFound") || strings.Contains(errStr, "404") || strings.Contains(errStr, "NoSuchKey") {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

