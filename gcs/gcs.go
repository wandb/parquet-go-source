package gcs

import (
	"context"

	"cloud.google.com/go/storage"
	"github.com/wandb/parquet-go/source"
)

type GcsFile struct {
	ProjectId  string
	BucketName string
	Ctx        context.Context

	Client         *storage.Client
	externalClient bool
	Bucket         *storage.BucketHandle
	FilePath       string
	FileReader     *storage.Reader
	FileWriter     *storage.Writer
}

func NewGcsFileWriter(ctx context.Context, projectId string, bucketName string, name string) (source.ParquetFile, error) {
	res := &GcsFile{
		ProjectId:  projectId,
		BucketName: bucketName,
		Ctx:        ctx,
		FilePath:   name,
	}
	return res.Create(name)
}

func NewGcsFileWriterWithClient(ctx context.Context, client *storage.Client, projectId string, bucketName string, name string) (source.ParquetFile, error) {
	res := &GcsFile{
		ProjectId:      projectId,
		BucketName:     bucketName,
		Ctx:            ctx,
		Client:         client,
		externalClient: true,
		FilePath:       name,
	}
	return res.Create(name)
}

func NewGcsFileReader(ctx context.Context, projectId string, bucketName string, name string) (source.ParquetFile, error) {
	res := &GcsFile{
		ProjectId:  projectId,
		BucketName: bucketName,
		Ctx:        ctx,
		FilePath:   name,
	}
	return res.Open(name)
}

func NewGcsFileReaderWithClient(ctx context.Context, client *storage.Client, projectId string, bucketName string, name string) (source.ParquetFile, error) {
	res := &GcsFile{
		ProjectId:      projectId,
		BucketName:     bucketName,
		Ctx:            ctx,
		Client:         client,
		externalClient: true,
		FilePath:       name,
	}
	return res.Open(name)
}

func (self *GcsFile) Create(name string) (source.ParquetFile, error) {
	var err error
	gcs := new(GcsFile)
	if self.Client == nil {
		gcs.Client, err = storage.NewClient(self.Ctx)
		gcs.externalClient = false
	} else {
		gcs.Client = self.Client
		gcs.externalClient = self.externalClient
	}
	gcs.FilePath = name
	if err != nil {
		return gcs, err
	}
	// must use existing bucket
	gcs.Bucket = gcs.Client.Bucket(self.BucketName)
	obj := gcs.Bucket.Object(name)
	gcs.FileWriter = obj.NewWriter(self.Ctx)
	return gcs, err

}
func (self *GcsFile) Open(name string) (source.ParquetFile, error) {
	var err error
	gcs := new(GcsFile)
	if self.Client == nil {
		gcs.Client, err = storage.NewClient(self.Ctx)
		gcs.externalClient = false
	} else {
		gcs.Client = self.Client
		gcs.externalClient = self.externalClient
	}
	gcs.FilePath = name
	if err != nil {
		return gcs, err
	}
	// must use existing bucket
	gcs.Bucket = gcs.Client.Bucket(self.BucketName)
	obj := gcs.Bucket.Object(name)
	gcs.FileReader, err = obj.NewReader(self.Ctx)
	return gcs, err
}
func (self *GcsFile) Seek(offset int64, pos int) (int64, error) {
	//Not implemented
	return 0, nil
}

func (self *GcsFile) Read(b []byte) (cnt int, err error) {
	var n int
	ln := len(b)
	for cnt < ln {
		n, err = self.FileReader.Read(b[cnt:])
		cnt += n
		if err != nil {
			break
		}
	}
	return cnt, err
}

func (self *GcsFile) Write(b []byte) (n int, err error) {
	return self.FileWriter.Write(b)
}

func (self *GcsFile) Close() error {
	if self.FileReader != nil {
		if err := self.FileReader.Close(); err != nil {
			return err
		}
	}
	if self.FileWriter != nil {
		if err := self.FileWriter.Close(); err != nil {
			return err
		}
	}
	if self.Client != nil && !self.externalClient {
		err := self.Client.Close()
		self.Client = nil
		if err != nil {
			return err
		}
	}
	return nil
}
