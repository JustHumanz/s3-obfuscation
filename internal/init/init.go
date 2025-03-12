package s3init

import (
	"encoding/json"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/justhumanz/s3-obfuscation/pkg"
)

type S3Init struct {
	S3Session *session.Session
	S3Base    pkg.S3Base
}

func (s *S3Init) Init() {
	tmpIndex, err := pkg.CreateTempFile()
	if err != nil {
		log.Fatalln(err)
	}
	defer os.Remove(tmpIndex.Name()) // Clean up the file when done
	defer tmpIndex.Close()

	data := map[string]interface{}{}
	jsonData, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		log.Fatalln("error marshaling JSON: %v", err)
	}

	// Write the JSON data to the file
	err = os.WriteFile(tmpIndex.Name(), jsonData, 0644)
	if err != nil {
		log.Fatalln("error writing to file: %v", err)
	}

	tmpIndexEnc := tmpIndex.Name() + ".enc"

	f, err := os.Create(tmpIndexEnc)
	if err != nil {
		log.Fatalln(err)
	}

	err = pkg.EncryptFile(tmpIndex.Name(), f.Name(), s.S3Base.GpGpass)
	if err != nil {
		log.Fatalln(err)
	}

	uploader := s3manager.NewUploader(s.S3Session)

	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(s.S3Base.BktName),
		Key:    aws.String(pkg.S3Index),
		Body:   f,
	})
	if err != nil {
		log.Fatalf("failed to upload file, %v", err)
	}
}
