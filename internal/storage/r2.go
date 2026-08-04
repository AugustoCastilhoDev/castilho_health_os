// Package storage generates temporary, scoped access to files kept in
// Cloudflare R2 — an S3-compatible object store. Nothing in this package
// ever touches file bytes: it only mints short-lived presigned URLs so the
// browser can PUT/GET directly against R2, keeping this Go server a thin
// authorizer instead of a proxy for potentially large exam PDFs, and
// keeping patient files off both Postgres and this server's disk
// entirely — the LGPD-relevant boundary the clinic needs.
package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// R2Client wraps an S3 client pointed at one Cloudflare R2 bucket. R2
// implements the S3 API, so the official AWS SDK works against it
// unmodified — only the endpoint, region and path-style addressing differ
// from talking to real S3.
type R2Client struct {
	s3     *s3.Client
	bucket string
}

// R2Config is exactly what Cloudflare's dashboard (R2 -> Manage R2 API
// Tokens) gives you when you create a token: an account ID and a
// key/secret pair scoped to R2. See the project's setup guide for where to
// find each value.
type R2Config struct {
	AccountID       string
	AccessKeyID     string
	SecretAccessKey string
	Bucket          string
}

// NewR2Client builds a client from static R2 credentials. It never reads
// env vars itself — main.go is responsible for loading R2Config from
// config.Config, the same way every other credential in this app is wired.
func NewR2Client(ctx context.Context, cfg R2Config) (*R2Client, error) {
	endpoint := fmt.Sprintf("https://%s.r2.cloudflarestorage.com", cfg.AccountID)

	awsCfg, err := config.LoadDefaultConfig(ctx,
		// R2 ignores the region value but the SDK requires one to be set;
		// "auto" is what Cloudflare's own docs use.
		config.WithRegion("auto"),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("storage: loading aws config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.UsePathStyle = true // R2 requires path-style addressing, not virtual-hosted buckets
	})

	return &R2Client{s3: client, bucket: cfg.Bucket}, nil
}

// PresignUpload returns a temporary URL the front-end can PUT the file
// bytes to directly. Callers should build fileKey as
// "tenants/{tenant_id}/patients/{patient_id}/{uuid}-{filename}" — that
// tenant-scoped prefix is what keeps one clinic's files logically
// separated inside the shared bucket, on top of (never instead of) the
// tenant_id column on PatientDocument in Postgres.
func (c *R2Client) PresignUpload(ctx context.Context, fileKey, contentType string, expiresIn time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(c.s3)
	req, err := presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(c.bucket),
		Key:         aws.String(fileKey),
		ContentType: aws.String(contentType),
	}, s3.WithPresignExpires(expiresIn))
	if err != nil {
		return "", fmt.Errorf("storage: presigning upload for %q: %w", fileKey, err)
	}
	return req.URL, nil
}

// PresignDownload returns a temporary, read-only URL for an existing
// object — what backs "view/download this exam" in the UI. Never expose
// the bucket's own URL (public or not); every access must go through a
// presigned URL scoped to exactly one object with a short expiry, so a
// leaked link can't be replayed to browse the rest of the tenant's files
// or used after it expires.
func (c *R2Client) PresignDownload(ctx context.Context, fileKey string, expiresIn time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(c.s3)
	req, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(fileKey),
	}, s3.WithPresignExpires(expiresIn))
	if err != nil {
		return "", fmt.Errorf("storage: presigning download for %q: %w", fileKey, err)
	}
	return req.URL, nil
}

// DownloadObject reads an object's bytes directly into memory — the one
// exception to this package's "never touch file bytes" rule, needed when
// the server itself has to embed a file into something it generates (the
// clinic logo into a letterhead PDF) rather than handing the browser a
// presigned URL. Only ever used for small, server-controlled assets like a
// logo — never for arbitrary patient documents, which stay presigned-only.
func (c *R2Client) DownloadObject(ctx context.Context, fileKey string) ([]byte, error) {
	out, err := c.s3.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(fileKey),
	})
	if err != nil {
		return nil, fmt.Errorf("storage: downloading %q: %w", fileKey, err)
	}
	defer out.Body.Close()
	data, err := io.ReadAll(out.Body)
	if err != nil {
		return nil, fmt.Errorf("storage: reading %q: %w", fileKey, err)
	}
	return data, nil
}

// DeleteObject removes the object outright — used when a PatientDocument
// row is deleted, so the file doesn't outlive the metadata that made it
// findable (an orphaned object in the bucket is invisible but still real
// patient data sitting there).
func (c *R2Client) DeleteObject(ctx context.Context, fileKey string) error {
	_, err := c.s3.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(fileKey),
	})
	if err != nil {
		return fmt.Errorf("storage: deleting %q: %w", fileKey, err)
	}
	return nil
}
