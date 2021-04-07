package db

import (
	mydb "FileStoreNetDisk_v3/db/mysql"
	"fmt"
	"time"
)

// UserFile : 用户文件表结构体
type UserFile struct {
	UserName     string
	FileSha1     string
	DocumentName string
	DocumentSize int64
	UploadAt     string
	IsFile       int
}

// OnUserFileUploadFinished : 更新用户文件表
func OnUserFileUploadFinished(username, filehash, filename string, filesize int64, parentID string) bool {
	stmt, err := mydb.DBConn().Prepare(
		"insert ignore into tbl_document (`user_name`,`file_sha1`,`document_name`," +
			"`document_size`,`update_at`,`parent_id`,`is_file`) values (?,?,?,?,?,?,?)")
	if err != nil {
		return false
	}
	defer stmt.Close()

	_, err = stmt.Exec(username, filehash, filename, filesize, time.Now(), parentID, 1)
	if err != nil {
		return false
	}
	return true
}

// QueryUserID : 获取 用户选中某一文件夹的 ID 号
func QueryUserID(username string, parentID int, documentName string) (int, error) {
	stmt, err := mydb.DBConn().Prepare(
		"select id " +
			" from tbl_document where user_name=? and parent_id = ? and document_name = ? ")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()
	var resID int
	err = stmt.QueryRow(username, parentID, documentName).Scan(&resID)
	if err != nil {
		return resID, err
	}
	return resID, err
}

// QueryUserFileMetas : 批量获取用户文件信息
func QueryUserFileMetas(username string, parentID int, limit int) ([]UserFile, error) {
	stmt, err := mydb.DBConn().Prepare(
		"select document_name , file_sha1 , document_size , update_at , is_file " +
			" from tbl_document where user_name=? and parent_id = ? order by is_file ASC limit ?")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(username, parentID, limit)
	if err != nil {
		return nil, err
	}

	var userFiles []UserFile
	for rows.Next() {
		file := UserFile{}
		err = rows.Scan(&file.DocumentName, &file.FileSha1, &file.DocumentSize, &file.UploadAt, &file.IsFile)
		if err != nil {
			fmt.Println(err.Error())
			break
		}
		userFiles = append(userFiles, file)
	}
	return userFiles, nil
}
