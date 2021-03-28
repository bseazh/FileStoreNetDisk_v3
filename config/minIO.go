package config

const (
	// Usessl : 使用 SSL
	// MinIOAccessKey : 访问Key
	// MinIOSecretKey : 访问密钥
	// MinIOGWEndpoint : gateway地址

	//UseSSL = false
	//MinIOGWEndpoint = "127.0.0.1:9000"
	//MinIOAccessKey = "minioadmin"
	//MinIOSecretKey = "minioadmin"


	UseSSL = true
	MinIOGWEndpoint = "play.min.io"
	MinIOAccessKey = "Q3AM3UQ867SPQQA43P2F"
	MinIOSecretKey = "zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG"

)

//MinIO 开启Server端服务
//1 . 在 Desktop 创建文件夹 "minIOTest"
//2 . 命令行执行 "minio server minIOTest"
//3 . 访问"http://127.0.0.1:9000"
//4 . 账号密码均为:"minioadmin"