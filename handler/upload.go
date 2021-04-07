package handler

import (
	"FileStoreNetDisk_v3/mq"
	"FileStoreNetDisk_v3/store/minIO"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	cmn "FileStoreNetDisk_v3/common"
	cfg "FileStoreNetDisk_v3/config"
	dblayer "FileStoreNetDisk_v3/db"
	"FileStoreNetDisk_v3/meta"
	"FileStoreNetDisk_v3/store/oss"
	"FileStoreNetDisk_v3/util"
)

// UploadHandler ： 处理文件上传
func UploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		// 返回上传html页面
		data, err := ioutil.ReadFile("./static/view/index.html")
		if err != nil {
			io.WriteString(w, "View internel server error")
			return
		}
		io.WriteString(w, string(data))
		// 另一种返回方式:
		// 动态文件使用http.HandleFunc设置，静态文件使用到http.FileServer设置(见main.go)
		// 所以直接redirect到http.FileServer所配置的url
		// http.Redirect(w, r, "/static/view/index.html",  http.StatusFound)
	} else if r.Method == "POST" {
		// 接收文件流及存储到本地目录
		file, head, err := r.FormFile("file")
		if err != nil {
			fmt.Printf("Failed to get data, err:%s\n", err.Error())
			return
		}
		defer file.Close()

		fileMeta := meta.FileMeta{
			FileName: head.Filename,
			Location: "/tmp/NetDisk/" + head.Filename,
			UploadAt: time.Now().Format("2006-01-02 15:04:05"),
		}

		newFile, err := os.Create(fileMeta.Location)
		if err != nil {
			fmt.Printf("Failed to create file, err:%s\n", err.Error())
			return
		}
		defer newFile.Close()

		fileMeta.FileSize, err = io.Copy(newFile, file)
		if err != nil {
			fmt.Printf("Failed to save data into file, err:%s\n", err.Error())
			return
		}

		newFile.Seek(0, 0)
		fileMeta.FileSha1 = util.FileSha1(newFile)

		// 游标重新回到文件头部
		newFile.Seek(0, 0)

		if cfg.CurrentStoreType == cmn.StoreMinIO {
			// 文件写入OSS存储
			MinIOPath := "/minIO/" + fileMeta.FileSha1

			if !cfg.AsyncTransferEnable {
				// 文件写入MinIO存储（文件上传）
				//if err := minIO.FPutObject(fileMeta.FileSha1 + path.Ext(fileMeta.FileName),fileMeta.Location) ; err != nil {
				//	log.Println( err.Error() )
				//	return
				//}

				// 文件写入MinIO存储 ( 文件流读入 )
				if err := minIO.PutObject(fileMeta.FileSha1+path.Ext(fileMeta.FileName), newFile); err != nil {
					log.Println(err.Error())
					return
				}

				// 文件写入MinIO存储 ( 文件流读入 + 加密 )
				//if err := minIO.FPutEncryptedObject(fileMeta.FileSha1 + path.Ext(fileMeta.FileName), newFile);
				//	err != nil {
				//	log.Println(err.Error())
				//	return
				//}
				log.Println(" success write in MinIO ")
				fileMeta.Location = MinIOPath
			} else {
				// 写入异步转移任务队列
				data := mq.TransferData{
					FileHash:      fileMeta.FileSha1,
					CurLocation:   fileMeta.Location,
					DestLocation:  MinIOPath,
					DestStoreType: cmn.StoreMinIO,
				}
				pubData, _ := json.Marshal(data)
				pubSuc := mq.Publish(
					cfg.TransExchangeName,
					cfg.TransMinIORoutingKey,
					pubData,
				)
				if !pubSuc {
					// TODO: 当前发送转移信息失败，稍后重试
				}
			}
		} else if cfg.CurrentStoreType == cmn.StoreOSS {
			// 文件写入OSS存储
			ossPath := "oss/" + fileMeta.FileSha1
			// 判断写入OSS为同步还是异步
			if !cfg.AsyncTransferEnable {
				err = oss.Bucket().PutObject(ossPath, newFile)
				if err != nil {
					fmt.Println(err.Error())
					w.Write([]byte("Upload failed!"))
					return
				}
				fileMeta.Location = ossPath
			} else {
				// 写入异步转移任务队列
				data := mq.TransferData{
					FileHash:      fileMeta.FileSha1,
					CurLocation:   fileMeta.Location,
					DestLocation:  ossPath,
					DestStoreType: cmn.StoreOSS,
				}
				pubData, _ := json.Marshal(data)
				pubSuc := mq.Publish(
					cfg.TransExchangeName,
					cfg.TransOSSRoutingKey,
					pubData,
				)
				if !pubSuc {
					// TODO: 当前发送转移信息失败，稍后重试
				}
			}
		}

		// meta.UpdateFileMeta(fileMeta)
		_ = meta.UpdateFileMetaDB(fileMeta)

		// 更新用户文件表记录
		r.ParseForm()
		username := r.Form.Get("username")
		parentID := r.Form.Get("parentID")
		suc := dblayer.OnUserFileUploadFinished(username, fileMeta.FileSha1,
			fileMeta.FileName, fileMeta.FileSize, parentID)

		// TODO : 文件夹 添加
		if suc {
			http.Redirect(w, r, "/static/view/home.html", http.StatusFound)
		} else {
			w.Write([]byte("Upload Failed."))
		}
	}
}

// UploadSucHandler : 上传已完成
func UploadSucHandler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Upload finished!")
}

// GetFileMetaHandler : 获取文件元信息
func GetFileMetaHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	filehash := r.Form["filehash"][0]
	//fMeta := meta.GetFileMeta(filehash)
	fMeta, err := meta.GetFileMetaDB(filehash)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if fMeta != nil {
		data, err := json.Marshal(fMeta)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Write(data)
	} else {
		w.Write([]byte(`{"code":-1,"msg":"no such file"}`))
	}
}

// FileQueryHandler : 查询批量的文件元信息
func FileQueryHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	limitCnt, _ := strconv.Atoi(r.Form.Get("limit"))
	username := r.Form.Get("username")
	parentID, _ := strconv.Atoi(r.Form.Get("parentID"))
	documentName := r.Form.Get("documentName")
	//fileMetas, _ := meta.GetLastFileMetasDB(limitCnt)

	var ID int
	var err error
	if parentID == 0 {
		ID, err = dblayer.QueryUserID(username, parentID, documentName)
	} else {
		ID = parentID
	}
	userFiles, err := dblayer.QueryUserFileMetas(username, ID, limitCnt)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(userFiles)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

// DownloadHandler : 文件下载接口
func DownloadHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	fsha1 := r.Form.Get("filehash")
	fm, _ := meta.GetFileMetaDB(fsha1)

	f, err := os.Open(fm.Location)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octect-stream")
	// attachment表示文件将会提示下载到本地，而不是直接在浏览器中打开
	w.Header().Set("content-disposition", "attachment; filename=\""+fm.FileName+"\"")
	w.Write(data)
}

// FileMetaUpdateHandler ： 更新元信息接口(重命名)
func FileMetaUpdateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		r.ParseForm()
		opType := r.Form.Get("op")
		fileSha1 := r.Form.Get("filehash")
		newFileName := r.Form.Get("filename")
		username := r.Form.Get("username")
		//w.Write([]byte(opType+fileSha1+newFileName))
		if opType != "0" {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		curFileMeta := meta.GetFileMeta(fileSha1)
		curFileMeta.FileName = newFileName
		meta.UpdateFileMeta(curFileMeta)

		// 更新文件表中的元信息记录
		if suc := meta.RenameFileMetaDB(fileSha1, newFileName, username); !suc {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
		//w.Write(data)
	}
}

// FileDeleteHandler : 删除文件及元信息
func FileDeleteHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	fileSha1 := r.Form.Get("filehash")
	username := r.Form.Get("username")

	fMeta := meta.GetFileMeta(fileSha1)
	// 删除文件
	os.Remove(fMeta.Location)
	// 删除文件元信息
	meta.RemoveFileMeta(fileSha1)
	// 删除表文件信息
	meta.RemoveFileMetaDB(fileSha1, username)

	w.WriteHeader(http.StatusOK)
}

// TryFastUploadHandler : 尝试秒传接口
func TryFastUploadHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	// 1. 解析请求参数
	username := r.Form.Get("username")
	filehash := r.Form.Get("filehash")
	filename := r.Form.Get("filename")
	parentID := r.Form.Get("parentID")
	filesize, _ := strconv.Atoi(r.Form.Get("filesize"))

	// 2. 从文件表中查询相同hash的文件记录
	fileMeta, err := meta.GetFileMetaDB(filehash)
	if err != nil {
		fmt.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// 3. 查不到记录则返回秒传失败
	if fileMeta == nil {
		resp := util.RespMsg{
			Code: -1,
			Msg:  "秒传失败，请访问普通上传接口",
		}
		w.Write(resp.JSONBytes())
		return
	}

	// 4. 上传过则将文件信息写入用户文件表， 返回成功
	suc := dblayer.OnUserFileUploadFinished(
		username, filehash, filename, int64(filesize), parentID)
	if suc {
		resp := util.RespMsg{
			Code: 0,
			Msg:  "秒传成功",
		}
		w.Write(resp.JSONBytes())
		return
	}
	resp := util.RespMsg{
		Code: -2,
		Msg:  "秒传失败，请稍后重试",
	}
	w.Write(resp.JSONBytes())
	return
}

// DownloadURLHandler : 生成文件的下载地址
func DownloadURLHandler(w http.ResponseWriter, r *http.Request) {
	filehash := r.Form.Get("filehash")
	// 从文件表查找记录
	row, _ := dblayer.GetFileMeta(filehash)

	fmt.Println("################## URL ###############")
	// 判断文件存在OSS，还是Ceph，还是在本地
	if strings.HasPrefix(row.FileAddr.String, "/tmp/NetDisk/") {
		username := r.Form.Get("username")
		token := r.Form.Get("token")
		tmpUrl := fmt.Sprintf("http://%s/file/download?filehash=%s&username=%s&token=%s",
			r.Host, filehash, username, token)
		log.Println(" #### 进入了 本地下载的url ", tmpUrl)

		w.Write([]byte(tmpUrl))
	} else if strings.HasPrefix(row.FileAddr.String, "/minIO/") {
		signedURL := minIO.DownloadURL(filehash + path.Ext(row.FileName.String))
		w.Write([]byte(signedURL))
	} else if strings.HasPrefix(row.FileAddr.String, "oss/") {
		// oss下载url
		signedURL := oss.DownloadURL(row.FileAddr.String)
		w.Write([]byte(signedURL))
	}
}
