package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	s3del "github.com/justhumanz/s3-obfuscation/internal/del"
	s3get "github.com/justhumanz/s3-obfuscation/internal/get"
	s3init "github.com/justhumanz/s3-obfuscation/internal/init"
	s3list "github.com/justhumanz/s3-obfuscation/internal/list"
	s3put "github.com/justhumanz/s3-obfuscation/internal/put"
	"github.com/justhumanz/s3-obfuscation/pkg"
)

var (
	_bucket           = flag.String("bucket", "test", "bucket name of s3")
	_bucketPassphrase = flag.String("passphrase", "", "the passphrase of object files")
)

// operation
const (
	_get  = "get"
	_put  = "put"
	_del  = "del"
	_list = "list"
	_init = "init"
)

// Here, you can choose the region of your bucket
func main() {
	accessKey := os.Getenv("ACCESS_KEY_ID")
	secretKey := os.Getenv("SECRET_ACCESS_KEY")
	endpoint := os.Getenv("ENDPOINT")
	region := os.Getenv("REGION")

	sess, err := session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(accessKey, secretKey, ""),
		Endpoint:    aws.String(endpoint),
		Region:      aws.String(region),
	})

	if err != nil {
		fmt.Println("Error creating session:", err)
	}

	flag.Parse()

	bkt_name := *_bucket
	Pass := *_bucketPassphrase
	if Pass == "" {
		reader := bufio.NewReader(os.Stdin)
		fmt.Println("Empty Passphrase!!!")
		for {
			fmt.Print("-> ")
			text, _ := reader.ReadString('\n')
			text = strings.Replace(text, "\n", "", -1)

			if text != "" {
				Pass = text
				break
			} else {
				fmt.Println("Bye")
				os.Exit(0)
			}
		}
	}

	args := flag.Args()
	if len(args) == 0 {
		log.Fatal("Please specify a subcommand.")
	}

	cmd, args := args[0], args[1:]

	switch cmd {
	case _list:
		//Get list of bkt object
		S3List := s3list.S3List{
			S3Session: sess,
			S3Base: pkg.S3Base{
				BktName: bkt_name,
				GpGpass: Pass,
			},
		}

		S3List.ListBkt()

	case _get:
		//Get/Download the object
		flag := flag.NewFlagSet(_get, flag.ExitOnError)
		outPut := flag.String("o", "/tmp", "output directory")
		flag.Parse(args)
		args := flag.Args()

		S3Get := s3get.S3Download{
			S3Session: sess,
			S3Base: pkg.S3Base{
				FilesName: args,
				BktName:   bkt_name,
				GpGpass:   Pass,
			},
			S3OutputDir: *outPut,
		}
		outPutDir, err := os.Stat(S3Get.S3OutputDir)
		if err != nil {
			log.Fatalf("Invalid dir %v", err)
		}

		if outPutDir.IsDir() {
			if len(S3Get.S3Base.FilesName) > 0 {
				S3Get.DownloadFile()
			} else {
				log.Fatalf("Empty target obj")
			}
		} else {
			log.Fatalf("output dir %s is a file", S3Get.S3OutputDir)
		}

	case _put:
		Upload := s3put.S3Put{
			S3Session: sess,
			S3Base: pkg.S3Base{
				FilesName: args,
				BktName:   bkt_name,
				GpGpass:   Pass,
			},
		}

		Upload.UploadFile()

	case _del:
		Delete := s3del.S3Del{
			S3Session: sess,
			S3Base: pkg.S3Base{
				FilesName: args,
				BktName:   bkt_name,
				GpGpass:   Pass,
			},
		}

		Delete.DeleteObj()

	case _init:
		reader := bufio.NewReader(os.Stdin)
		fmt.Println("This action will remove your existing index file, run this only once (y/n)")
		for {
			fmt.Print("-> ")
			text, _ := reader.ReadString('\n')
			text = strings.Replace(text, "\n", "", -1)

			if strings.Compare("yes", text) == 0 || strings.Compare("y", text) == 0 {
				break
			} else {
				fmt.Println("Bye")
				os.Exit(0)
			}
		}

		Init := s3init.S3Init{
			S3Session: sess,
			S3Base: pkg.S3Base{
				BktName: bkt_name,
				GpGpass: Pass,
			},
		}
		Init.Init()
	default:
		log.Fatalf("Unrecognized command %q. "+
			"Command must be one of: branch, checkout", cmd)
	}

}
