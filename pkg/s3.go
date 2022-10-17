package pkg

import (
	"context"
	"io"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type s3Helper struct {
	bucket string
	client *s3.Client
	cache  map[string]time.Time
	tmp    map[string]time.Time
}

func NewS3Helper(ctx context.Context, bucket string) (*s3Helper, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}
	return &s3Helper{
		bucket: bucket,
		client: s3.NewFromConfig(cfg),
		cache:  make(map[string]time.Time),
		tmp:    make(map[string]time.Time),
	}, nil
}

func (s *s3Helper) RefreshContext(ctx context.Context) error {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return err
	}
	s.client = s3.NewFromConfig(cfg)
	return nil
}

type S3object struct {
	objKey string
	body   io.ReadCloser
	err    error
}

func (s S3object) Key() string {
	return s.objKey
}

func (s S3object) Reader() io.ReadCloser {
	return s.body
}

var _ EncryptedObject = S3object{}

// call to aws api to list all objects within bucket set by AWS_S3_BUCKET
// returns list of objects that do not match in memory <obj name: modified date> map
func (s *s3Helper) GetUpdatedObjects(ctx context.Context) ([]EncryptedObject, error) {
	objects, err := s.client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: &s.bucket,
	})
	if err != nil {
		return nil, err
	}

	updatedObjKeys := []*string{}
	for _, object := range objects.Contents {
		// include objects that do not exist in in-memory cache
		// or objects that have different modified times
		_, exists := s.cache[*object.Key]
		if !exists || !object.LastModified.Equal(s.cache[*object.Key]) {
			updatedObjKeys = append(updatedObjKeys, object.Key)
			s.tmp[*object.Key] = *object.LastModified
		}
	}

	var wg sync.WaitGroup
	ch := make(chan S3object)

	for _, key := range updatedObjKeys {
		wg.Add(1)
		go s.getS3Object(ctx, *key, &wg, ch)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	result := []S3object{}
	for obj := range ch {
		if obj.err != nil {
			return nil, err
		}
		result = append(result, obj)
	}

	return convert(result), nil
}

// goroutine func
// aws call for details of specific object. returned via channel
func (s *s3Helper) getS3Object(ctx context.Context, key string, wg *sync.WaitGroup, ch chan<- S3object) {
	defer wg.Done()
	object := S3object{}
	result, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &s.bucket,
		Key:    &key,
	})
	if err != nil {
		object.err = err
		ch <- object
	}
	object.body = result.Body
	object.objKey = key
	ch <- object
}

// decodes object keys and casts s3object to encryptedobject type for use gpg use
func convert(originals []S3object) []EncryptedObject {
	converted := []EncryptedObject{}
	for _, o := range originals {
		converted = append(converted, EncryptedObject(o))
	}
	return converted
}

// updates cache to reflect successful changes from current iteration
func (s *s3Helper) UpdateCache() {
	for key, dur := range s.tmp {
		s.cache[key] = dur
	}
	s.tmp = make(map[string]time.Time) // clear tmp in prep for next run
}
