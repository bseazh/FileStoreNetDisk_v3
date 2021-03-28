package minIO

import (
	cfg "FileStoreNetDisk_v3/config"
	"context"
	"fmt"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/minio/minio-go/v7/pkg/encrypt"
	"log"
	"net/url"
	"os"
	"time"
)

// 全局变量
var (
	Client     *minio.Client
	err        error
	ctx        context.Context
	cancel     func()
	BucketName = "hhz-demo"
)

// InitClient : 连接 minIO 返回对应client
func InitClient(){

	// 初使化minio client对象。
	if Client, err = minio.New(cfg.MinIOGWEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinIOAccessKey, cfg.MinIOSecretKey, ""),
		Secure: cfg.UseSSL,
	}) ; err != nil {
		fmt.Println(err)
		return
	}
}

func init(){
	InitClient()
	CreateBucket(BucketName)
}

// CreateBucket : 创建名称为 "BucketName"
func CreateBucket( BucketName string ){
	ctx, cancel = context.WithTimeout(context.Background(), 100 * time.Second)
	defer cancel()

	// Region : Beijing
	location := "cn-north-1"

	if err = Client.MakeBucket(ctx, BucketName, minio.MakeBucketOptions{Region: location});
		err != nil {
		// 检查存储桶是否已经存在。
		exists, errBucketExists := Client.BucketExists(ctx, BucketName)
		if errBucketExists == nil && exists {
			log.Printf("We already own %s\n", BucketName)
		} else {
			log.Fatalln(err)
		}
	} else {
		log.Printf("Successfully created %s\n", BucketName)
	}
}

// FPutObject : 上传文件到MinIO集群(文件)
func FPutObject(objectName string, path string ) (err error) {

	if _, err = Client.FPutObject(context.Background(), BucketName, objectName, path, minio.PutObjectOptions{
		ContentType: "application/octet-stream",
	}); err != nil {
		log.Fatalln(err)
	}
	log.Printf("Successfully uploaded %s",objectName)
	return nil
}

// PutObject : 上传文件到MinIO集群( 文件流 )
func PutObject( objectName string, object *os.File ) error  {

	var (
		objectStat os.FileInfo
		n minio.UploadInfo
	)

	if objectStat , err = object.Stat();
		err != nil {
		log.Fatalln(err)
		return err
	}

	if n, err = Client.PutObject(context.Background(), BucketName, objectName, object, objectStat.Size(),
		minio.PutObjectOptions{ContentType: "application/octet-stream"});
		err != nil {
		log.Fatalln(err)
		return err
	}
	log.Println("Uploaded", objectName , " of size: ", n, "Successfully.")
	return nil
}

// FPutEncryptedObject : 上传文件到MinIO集群( 文件流 + 加密 )
func FPutEncryptedObject( objectName string, file *os.File ) error {

	var fstat os.FileInfo
	// Get file stats.
	fstat, err = file.Stat()
	if err != nil {
		log.Fatalln(err)
		return err
	}

	password := "~!@#$%^&*()_++_)(*&^%$#@!~"    // Specify your password.

	// New SSE-C where the cryptographic key is derived from a password and the objectname + bucketname as salt
	encryption := encrypt.DefaultPBKDF([]byte(password), []byte(BucketName+objectName))

	// Encrypt file content and upload to the server
	n, err := Client.PutObject(context.Background(), BucketName, objectName, file, fstat.Size(),
		minio.PutObjectOptions{ServerSideEncryption: encryption})
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("Uploaded", objectName , " of size: ", n, "Successfully.")
	return nil
}

func DownloadURL( filename string ) string {

	// Set request parameters
	reqParams := make(url.Values)
	reqParams.Set("response-content-disposition","attachment; filename=\""+filename+"\"")

	// Gernerate presigned get object url.
	presignedURL, err := Client.PresignedGetObject(context.Background(), BucketName, filename, time.Duration(1000)*time.Second, reqParams)
	if err != nil {
		log.Fatalln(err)
		return ""
	}
	log.Println(presignedURL)
	return presignedURL.String()
}
