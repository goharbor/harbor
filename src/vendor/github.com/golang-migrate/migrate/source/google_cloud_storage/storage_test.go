package googlecloudstorage

import (
	"testing"

	"github.com/fsouza/fake-gcs-server/fakestorage"
	"github.com/golang-migrate/migrate/source"
	st "github.com/golang-migrate/migrate/source/testing"
)

func Test(t *testing.T) {
	server := fakestorage.NewServer([]fakestorage.Object{
		{BucketName: "some-bucket", Name: "staging/migrations/1_foobar.up.sql", Content: []byte("1 up")},
		{BucketName: "some-bucket", Name: "staging/migrations/1_foobar.down.sql", Content: []byte("1 down")},
		{BucketName: "some-bucket", Name: "prod/migrations/1_foobar.up.sql", Content: []byte("1 up")},
		{BucketName: "some-bucket", Name: "prod/migrations/1_foobar.down.sql", Content: []byte("1 down")},
		{BucketName: "some-bucket", Name: "prod/migrations/3_foobar.up.sql", Content: []byte("3 up")},
		{BucketName: "some-bucket", Name: "prod/migrations/4_foobar.up.sql", Content: []byte("4 up")},
		{BucketName: "some-bucket", Name: "prod/migrations/4_foobar.down.sql", Content: []byte("4 down")},
		{BucketName: "some-bucket", Name: "prod/migrations/5_foobar.down.sql", Content: []byte("5 down")},
		{BucketName: "some-bucket", Name: "prod/migrations/7_foobar.up.sql", Content: []byte("7 up")},
		{BucketName: "some-bucket", Name: "prod/migrations/7_foobar.down.sql", Content: []byte("7 down")},
		{BucketName: "some-bucket", Name: "prod/migrations/not-a-migration.txt"},
		{BucketName: "some-bucket", Name: "prod/migrations/0-random-stuff/whatever.txt"},
	})
	defer server.Stop()
	driver := gcs{
		bucket:     server.Client().Bucket("some-bucket"),
		prefix:     "prod/migrations/",
		migrations: source.NewMigrations(),
	}
	err := driver.loadMigrations()
	if err != nil {
		t.Fatal(err)
	}
	st.Test(t, &driver)
}
