package main

import (
	"FileStoreNetDisk_v3/common"
	"bufio"
	"encoding/json"
	"FileStoreNetDisk_v3/config"
	dblayer "FileStoreNetDisk_v3/db"
	"FileStoreNetDisk_v3/mq"
	"FileStoreNetDisk_v3/store/oss"
	"FileStoreNetDisk_v3/store/minIO"
	"log"
	"os"
	"path"
)

// ProcessTransfer : 处理文件转移
func ProcessTransfer(msg []byte) bool {
	log.Println(string(msg))

	pubData := mq.TransferData{}
	err := json.Unmarshal(msg, &pubData)
	if err != nil {
		log.Println(err.Error())
		return false
	}

	fin, err := os.Open(pubData.CurLocation)
	if err != nil {
		log.Println(err.Error())
		return false
	}
	if pubData.DestStoreType == common.StoreOSS {
		// 上传 OSS 公有云
		err = oss.Bucket().PutObject(
			pubData.DestLocation,
			bufio.NewReader(fin))
		if err != nil {
			log.Println(err.Error())
			return false
		}

		_ = dblayer.UpdateFileLocation(
			pubData.FileHash,
			pubData.DestLocation)
		log.Printf("文件: \"%s\" 成功上传 到 OSS ",fin.Name())
		return true
	}else {
		// 上传 MinIO 私有云
		err = minIO.PutObject(pubData.FileHash+path.Ext(fin.Name()),fin)
		if err != nil {
			log.Println(err.Error())
			return false
		}

		_ = dblayer.UpdateFileLocation(
			pubData.FileHash,
			pubData.DestLocation)

		log.Printf("文件(加密):\"%s\" 成功上传 到 MinIO",fin.Name())
		return true
	}
}

func main() {
	if !config.AsyncTransferEnable {
		log.Println("异步转移文件功能目前被禁用，请检查相关配置")
		return
	}
	log.Println("文件转移服务启动中，开始监听转移任务队列...")
	mq.StartConsume(
		config.TransOSSQueueName,
		"transfer_oss",
		ProcessTransfer)
}
