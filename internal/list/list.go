package s3list

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/justhumanz/s3-obfuscation/pkg"
)

const Index = "/tmp/index"

type S3List struct {
	S3Session  *session.Session
	S3Download *s3manager.Downloader
	S3Base     pkg.S3Base
}

// List the obj on selected bucket
func (s *S3List) ListBkt() {
	s.S3Download = s3manager.NewDownloader(s.S3Session)

	indexFile := s.GetIndex()
	err := pkg.DecryptFile(indexFile, Index, s.S3Base.GpGpass)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	file, err := os.Open(Index)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()
	defer os.Remove(Index)

	data := make(map[string]interface{})
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&data)
	if err != nil {
		fmt.Println("Error decoding JSON:", err)
		return
	}

	pkg.PrintIndex(data, "")
}

// Download index file
func (s *S3List) GetIndex() string {
	tmpFile, err := pkg.CreateTempFile()
	if err != nil {
		log.Fatalln(err)
	}

	_, err = s.S3Download.Download(tmpFile,
		&s3.GetObjectInput{
			Bucket: aws.String(s.S3Base.BktName),
			Key:    aws.String(pkg.S3Index),
		})
	if err != nil {
		log.Fatalf("Unable to download item %v", err)
	}

	return tmpFile.Name()
}

// Wrapper for GetIndex
func GetIndex(d *s3manager.Downloader, b pkg.S3Base) string {
	s := S3List{
		S3Download: d,
		S3Base:     b,
	}
	return s.GetIndex()
}
