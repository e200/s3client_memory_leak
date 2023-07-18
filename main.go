package main

import (
	"bytes"
	"log"
	"net/http"
	"net/http/pprof"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/mazen160/go-random"
)

func main() {
	router := mux.NewRouter()

	router.HandleFunc("/debug/pprof/", pprof.Index)
	router.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	router.HandleFunc("/debug/pprof/profile", pprof.Profile)
	router.HandleFunc("/debug/pprof/symbol", pprof.Symbol)

	router.Handle("/debug/pprof/goroutine", pprof.Handler("goroutine"))
	router.Handle("/debug/pprof/heap", pprof.Handler("heap"))
	router.Handle("/debug/pprof/allocs", pprof.Handler("allocs"))
	router.Handle("/debug/pprof/threadcreate", pprof.Handler("threadcreate"))
	router.Handle("/debug/pprof/block", pprof.Handler("block"))

	srv := &http.Server{
		Handler: router,
		Addr:    "127.0.0.1:8000",
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	cycles := 300

	sess, err := session.NewSession(&aws.Config{
		Region:           aws.String("us-west-2"),
		Endpoint:         aws.String("http://localhost:4566"),
		S3ForcePathStyle: aws.Bool(true),
		Credentials: credentials.NewStaticCredentialsFromCreds(credentials.Value{
			AccessKeyID:     "AccessKeyID",
			SecretAccessKey: "SecretAccessKey",
			ProviderName:    "AssumeRoleCredentialsProvider",
		}),
	})
	if err != nil {
		log.Fatal(err)
	}

	// S3 service client the Upload manager will use.
	s3Svc := s3.New(sess)

	// Create an uploader with S3 client and default options
	uploader := s3manager.NewUploaderWithClient(s3Svc)

	go func() {
		for i := 0; i < cycles; i++ {
			filename, err := uuid.NewRandom()
			if err != nil {
				log.Fatal(err)
			}

			content, err := random.Bytes(2 * 1024 * 1024) // 2mb of data
			if err != nil {
				log.Fatal(err)
			}

			_, err = uploader.Upload(&s3manager.UploadInput{
				Bucket: aws.String("uploads"),
				Key:    aws.String(filename.String()),
				Body:   bytes.NewBuffer(content),
			})
			if err != nil {
				log.Fatal(err)
			}

			log.Default().Println(i)
		}
	}()

	log.Fatal(srv.ListenAndServe())
}
