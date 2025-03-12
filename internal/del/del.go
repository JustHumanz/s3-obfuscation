package s3del

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	s3list "github.com/justhumanz/s3-obfuscation/internal/list"
	"github.com/justhumanz/s3-obfuscation/pkg"
)

type S3Del struct {
	S3Session *session.Session
	S3Base    pkg.S3Base
}

func (s *S3Del) DeleteObj() {
	d := s3manager.NewDownloader(s.S3Session)
	uploader := s3manager.NewUploader(s.S3Session, func(u *s3manager.Uploader) {
		u.PartSize = 50 * 1024 * 1024
		u.LeavePartsOnError = true
	})
	svc := s3.New(s.S3Session)

	indexEnc := s3list.GetIndex(d, s.S3Base)
	indexTmp, err := pkg.CreateTempFile()
	if err != nil {
		log.Fatalln("Error opening file:", err)
	}
	defer os.Remove(indexTmp.Name()) // Clean up the file when done
	defer indexTmp.Close()

	err = pkg.DecryptFile(indexEnc, indexTmp.Name(), s.S3Base.GpGpass)
	if err != nil {
		log.Fatalf("failed Decrypt index enc %v", err)
	}

	data := make(map[string]interface{})
	err = pkg.ReadJSON(indexTmp.Name(), &data)
	if err != nil {
		log.Fatalln("Error decoding JSON:", err)
	}

	for _, FileName := range s.S3Base.FilesName {
		if !strings.HasPrefix(FileName, "/") {
			FileName = filepath.Join("/", FileName)
		}

		uploadFileList := strings.Split(FileName, "/")
		uploadFileMd5 := []string{}

		newDataMap := make(map[string]interface{})
		for _, v := range uploadFileList {
			if v == "" {
				continue
			}

			pathBase64 := pkg.EncodeBase64(v)
			pathMd5 := pkg.ComputeMD5(pathBase64)

			if v == uploadFileList[len(uploadFileList)-1] {
				newDataMap[pathMd5] = map[string]interface{}{
					"name": pathBase64,
					"type": "file",
				}
			} else {
				newDataMap[pathMd5] = map[string]interface{}{
					"name": pathBase64,
					"type": "dir",
					"sub":  make(map[string]interface{}),
				}
			}

			uploadFileMd5 = append(uploadFileMd5, pathMd5)
		}

		NewDataStruct := pkg.UpdateIndex{
			Index: uploadFileMd5,
			Map:   newDataMap,
		}

		// Delete the obj
		delObj := strings.Join(uploadFileMd5, "/")
		_, err = svc.DeleteObject(&s3.DeleteObjectInput{
			Bucket: aws.String(s.S3Base.BktName),
			Key:    aws.String(delObj),
		})
		if err != nil {
			log.Fatalf("Unable to delete object %q from bucket %q, %v", delObj, s.S3Base.BktName, err)
		}

		err = svc.WaitUntilObjectNotExists(&s3.HeadObjectInput{
			Bucket: aws.String(s.S3Base.BktName),
			Key:    aws.String(delObj),
		})
		if err != nil {
			log.Fatalf("Error occurred while waiting for object %q to be deleted, %v", delObj, err)
		}

		pkg.UpdateDataIndex(data, NewDataStruct, false)
	}

	indexTmpEnc := indexTmp.Name() + ".enc"
	err = pkg.SaveJson(data, indexTmp.Name())
	if err != nil {
		log.Fatalf("failed save file, %v", err)
	}

	err = pkg.EncryptFile(indexTmp.Name(), indexTmpEnc, s.S3Base.GpGpass)
	if err != nil {
		log.Fatalf("failed enc index enc %v", err)
	}

	f, err := os.Open(indexTmpEnc)
	if err != nil {
		log.Fatalf("failed open index enc %v", err)
	}

	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(s.S3Base.BktName),
		Key:    aws.String(pkg.S3Index),
		Body:   f,
	})
	if err != nil {
		log.Fatalf("failed to upload file, %v", err)
	}
}
