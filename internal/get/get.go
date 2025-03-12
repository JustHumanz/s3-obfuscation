package s3get

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/justhumanz/s3-obfuscation/pkg"
	"github.com/schollz/progressbar/v3"
)

type S3Download struct {
	S3Session   *session.Session
	S3Base      pkg.S3Base
	S3OutputDir string
}

func (s *S3Download) DownloadFile() {
	downloader := s3manager.NewDownloader(s.S3Session)

	//Do iteration for each filename
	for _, FileName := range s.S3Base.FilesName {
		downloadFileList := strings.Split(FileName, "/")
		downloadFileMd5 := []string{}

		for _, v := range downloadFileList {
			//Ignore the empty filename
			if v == "" {
				continue
			}

			pathBase64 := pkg.EncodeBase64(strings.TrimSuffix(v, "\n"))
			pathMd5 := pkg.ComputeMD5(pathBase64)
			downloadFileMd5 = append(downloadFileMd5, pathMd5)
		}

		downloadFilename := strings.Join(downloadFileMd5, "/")
		objContent, err := s.getObjContent(downloadFilename)
		if err != nil {
			log.Fatalln(err)
		}

		bar := progressbar.DefaultBytes(
			*objContent.ContentLength,
			"[info] downloading s3: "+FileName,
		)

		tmpFile, err := pkg.CreateTempFile()
		if err != nil {
			log.Fatalln(err)
		}
		defer os.Remove(tmpFile.Name()) // Clean up the file when done
		defer tmpFile.Close()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go s.updateProgressBar(ctx, bar, tmpFile.Name())
		_, err = downloader.DownloadWithContext(ctx, tmpFile,
			&s3.GetObjectInput{
				Bucket: aws.String(s.S3Base.BktName),
				Key:    aws.String(downloadFilename),
			})
		if err != nil {
			log.Fatalf("Unable to download item %v", err)
		}
		bar.Finish()

		outPutFile := filepath.Join(s.S3OutputDir, downloadFileList[len(downloadFileList)-1])
		fmt.Println("[info] Decrypt File:", outPutFile)
		err = pkg.DecryptFile(tmpFile.Name(), outPutFile, s.S3Base.GpGpass)
		if err != nil {
			log.Fatalf("failed enc index enc %v", err)
		}
	}
}

func (s *S3Download) getObjContent(encFileName string) (*s3.GetObjectOutput, error) {
	svc := s3.New(s.S3Session)
	res, err := svc.GetObject(&s3.GetObjectInput{
		Bucket: &s.S3Base.BktName,
		Key:    aws.String(encFileName),
	})
	if err != nil {
		return nil, err
	}

	return res, nil

}

func (s *S3Download) updateProgressBar(ctx context.Context, bar *progressbar.ProgressBar, filePath string) {
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			info, err := os.Stat(filePath)
			if err != nil {
				continue
			}
			_ = bar.Set64(info.Size())
		}
	}
}
