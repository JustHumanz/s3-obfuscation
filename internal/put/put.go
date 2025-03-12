package s3put

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	s3list "github.com/justhumanz/s3-obfuscation/internal/list"
	"github.com/justhumanz/s3-obfuscation/pkg"
	"github.com/schollz/progressbar/v3"
)

type S3Put struct {
	S3Session *session.Session
	S3Base    pkg.S3Base
}

func (s *S3Put) UploadFile() {
	d := s3manager.NewDownloader(s.S3Session)
	indexEnc := s3list.GetIndex(d, s.S3Base)
	indexTmp, err := pkg.CreateTempFile()
	if err != nil {
		log.Fatalln("Error opening file:", err)
	}
	defer os.Remove(indexTmp.Name()) // Clean up the file when done
	defer indexTmp.Close()

	uploader := s3manager.NewUploader(s.S3Session, func(u *s3manager.Uploader) {
		u.PartSize = 50 * 1024 * 1024
		u.LeavePartsOnError = true
	})

	pkg.DecryptFile(indexEnc, indexTmp.Name(), s.S3Base.GpGpass)
	data := make(map[string]interface{})
	err = pkg.ReadJSON(indexTmp.Name(), &data)
	if err != nil {
		log.Fatalln("Error decoding JSON:", err)
	}

	for _, FileName := range s.S3Base.FilesName {
		_, err := os.Stat(FileName)
		if err != nil {
			log.Printf("Invalid file dir %v", err)
			continue
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

		pkg.UpdateDataIndex(data, NewDataStruct, true)

		tmpFile, err := pkg.CreateTempFile()
		if err != nil {
			log.Fatalln(err)
		}

		fmt.Println("[info] Encrypt File:", FileName)
		err = pkg.EncryptFile(FileName, tmpFile.Name(), s.S3Base.GpGpass)
		if err != nil {
			log.Fatalf("failed to stat file %v, %v", tmpFile, err)
		}

		fileInfo, err := tmpFile.Stat()
		if err != nil {
			log.Fatalf("failed to stat file %v, %v", tmpFile, err)
		}

		bar := progressbar.DefaultBytes(
			fileInfo.Size(),
			"[info] Upload s3: "+FileName,
		)

		reader := &CustomReader{
			fp:   tmpFile,
			size: fileInfo.Size(),
			bar:  bar,
		}

		uploadKey := strings.Join(uploadFileMd5, "/")
		_, err = uploader.Upload(&s3manager.UploadInput{
			Bucket: aws.String(s.S3Base.BktName),
			Key:    aws.String(uploadKey),
			Body:   reader,
		})
		if err != nil {
			log.Fatalf("failed to put file %v, %v", FileName, err)
		}

		bar.Finish()

		tmpFile.Close()
		os.Remove(tmpFile.Name())

		//log.Println(output.Location)
		//pkg.PrintIndex(data, "")
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

type CustomReader struct {
	fp   *os.File
	size int64
	read int64
	bar  *progressbar.ProgressBar
}

func (r *CustomReader) Read(p []byte) (int, error) {
	return r.fp.Read(p)
}

func (r *CustomReader) ReadAt(p []byte, off int64) (int, error) {
	n, err := r.fp.ReadAt(p, off)
	if err != nil {
		return n, err
	}

	r.read += int64(n)
	_ = r.bar.Set64(r.read)

	return n, err
}

func (r *CustomReader) Seek(offset int64, whence int) (int64, error) {
	return r.fp.Seek(offset, whence)
}
