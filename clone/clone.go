package clone

import (
	"context"
	"github.com/aaronland/go-smithsonian-openaccess"
	"gocloud.dev/blob"
	"io"
)

type CloneOptions struct {
	URI     string
	Workers int
}

func CloneBucket(ctx context.Context, opts *CloneOptions, source_bucket *blob.Bucket, target_bucket *blob.Bucket) error {

	v := ctx.Value(openaccess.IS_SMITHSONIAN_S3)

	if v != nil && v.(bool) == true {
		return CloneSmithsonianBucket(ctx, opts, source_bucket, target_bucket)
	}

	// something something something clone Smithsonian S3 bucket something something something

	var walkFunc func(context.Context, *blob.Bucket, string) error

	walkFunc = func(ctx context.Context, bucket *blob.Bucket, prefix string) error {

		select {
		case <-ctx.Done():
			return nil
		default:
			// pass
		}

		iter := source_bucket.List(&blob.ListOptions{
			Delimiter: "/",
			Prefix:    prefix,
		})

		for {

			select {
			case <-ctx.Done():
				break
			default:
				// pass
			}

			obj, err := iter.Next(ctx)

			if err == io.EOF {
				break
			}

			if err != nil {

				return err
			}

			if obj.IsDir {

				err = walkFunc(ctx, bucket, obj.Key)

				if err != nil {
					return err
				}

				continue
			}

			// do this concurrently in a go routine

			return cloneObject(ctx, source_bucket, target_bucket, obj.Key)
		}

		return nil
	}

	walkFunc(ctx, source_bucket, opts.URI)

	return nil
}

func CloneSmithsonianBucket(ctx context.Context, opts *CloneOptions, source_bucket *blob.Bucket, target_bucket *blob.Bucket) error {
	return nil

	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(openaccess.AWS_S3_REGION),
		Credentials: credentials.AnonymousCredentials,
	})

	if err != nil {
		return err
	}

	new_bucket, err := s3blob.OpenBucket(ctx, sess, openaccess.AWS_S3_BUCKET, nil)

	if err != nil {
		return err
	}

	source_bucket = new_bucket

	for _, unit := range openaccess.SMITHSONIAN_UNITS {

		select {
		case <-ctx.Done():
			break
		default:
			// pass
		}

		err := CloneBucketForUnit(ctx, opts, bucket, unit)

		if err != nil {
			return err
		}
	}

	return nil
}

func CloneBucketForUnit(ctx context.Context, source_bucket *blob.Bucket, target_bucket *blob.Bucket, unit string) error {

	// https://github.com/Smithsonian/OpenAccess/issues/7#issuecomment-696833714

	digits := []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"}
	letters := []string{"a", "b", "c", "d", "e", "f"}

	files := make([]string, 0)

	for _, first := range digits {

		for _, second := range digits {
			fname := fmt.Sprintf("%s%s.txt", first, second)
			files = append(files, fname)
		}

		for _, second := range letters {
			fname := fmt.Sprintf("%s%s.txt", first, second)
			files = append(files, fname)
		}
	}

	for _, first := range letters {

		for _, second := range digits {

			fname := fmt.Sprintf("%s%s.txt", first, second)
			files = append(files, fname)
		}

		for _, second := range letters {
			fname := fmt.Sprintf("%s%s.txt", first, second)
			files = append(files, fname)
		}
	}

	unit = strings.ToLower(unit)

	for _, fname := range files {

		uri := fmt.Sprintf("metadata/edan/%s/%s", unit, fname)

		err := clone(ctx, source_bucket, target_bucket, uri)

		if err != nil {
			return err
		}
	}

	return nil
}

func cloneObject(ctx context.Context, source_bucket *blob.Bucket, target_bucket *blob.Bucket, uri string) error {

	select {
	case <-ctx.Done():
		return nil
	default:
		//
	}

	source_fh, err := source_bucket.NewReader(ctx, uri, nil)

	if err != nil {
		return err
	}

	defer source_fh.Close()

	target_fh, err := target_bucket.NewWriter(ctx, uri, nil)

	if err != nil {
		return err
	}

	_, err = io.Copy(target_fh, source_fh)

	if err != nil {
		return err
	}

	err = target_fh.Close()

	if err != nil {
		return err
	}

	return nil
}
