package db

import (
	mydb "FileStoreNetDisk_v3/db/mysql"
	"fmt"
)

// CreateDocument : 创建新的文件夹
func CreateDocument(username, documentName string, parentID int, documentSize int64) bool {

	stmt, err := mydb.DBConn().Prepare(
		"insert ignore into tbl_document (`user_name`,`document_name`,`parent_id`) values (?,?,?);")
	if err != nil {
		fmt.Println("Failed to insert, err:" + err.Error())
		return false
	}
	defer stmt.Close()

	ret, err := stmt.Exec(username, documentName, parentID)
	if err != nil {
		fmt.Println("Failed to insert, err:" + err.Error())
		return false
	}

	if rowsAffected, err := ret.RowsAffected(); nil == err && rowsAffected > 0 {
		UpdateDocument(username, documentName, parentID, documentSize)
		return true
	}
	return false
}

// OpenFolder : 通过 ( username documentName parentID) 找到文件(夹)更新其对应的文件(夹)名;
func OpenFolder(username, documentName string, parentID int) (int, error) {
	id, err := QueryUserID(username, parentID, documentName)
	return id, err
}

// GoUpFolder : 通过( username , parentID ) 找到上一层的文件夹;
func GoUpFolder(username string, parentID int) (int, error) {
	stmt, err := mydb.DBConn().Prepare(
		"select parent_id " +
			" from tbl_document where user_name=? and id = ? ")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()
	var resID int
	err = stmt.QueryRow(username, parentID).Scan(&resID)
	if err != nil {
		return resID, err
	}
	return resID, err
}

// RenameDocument : 通过 ( documentName username ) 找到文件(夹)更新其对应的文件(夹)名;
func RenameDocument(username, documentName, filename string) bool {

	stmt, err := mydb.DBConn().Prepare(
		"UPDATE tbl_document SET document_name = ? where user_name = ? and document_name = ?")
	if err != nil {
		fmt.Println("Failed to prepare statement, err:" + err.Error())
		return false
	}
	defer stmt.Close()

	_, err = stmt.Exec(filename, username, documentName)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	return true
}

// TableDocument : 文件表结构体
type TableDocument struct {
	UserName     string
	FileSha1     string
	ParentID     int
	DocumentName string
	DocumentSize int64
	UploadAt     string
	IsFile       int
}

// GetDocumentList : 从mysql批量获取文件元信息
func GetDocumentList(username string, parentID int) ([]TableDocument, error) {

	stmt, err := mydb.DBConn().Prepare(
		"select document_name , file_sha1 , document_size , update_at , is_file " +
			" from tbl_document where user_name=? and parent_id = ? order by is_file ASC")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(username, parentID)
	if err != nil {
		return nil, err
	}

	var userFiles []TableDocument
	for rows.Next() {
		file := TableDocument{}
		err = rows.Scan(&file.DocumentName, &file.FileSha1, &file.DocumentSize, &file.UploadAt, &file.IsFile)
		if err != nil {
			fmt.Println(err.Error())
			break
		}
		userFiles = append(userFiles, file)
	}
	return userFiles, nil

}

// RemoveDocument : 通过 ( documentName username ) 删除之间的索引 ;
func RemoveDocument(username, documentName string, parentID int, isFile int, documentSize int64) bool {

	if isFile == 1 {
		// 文件
		stmt, err := mydb.DBConn().Prepare(
			"DELETE FROM tbl_document WHERE user_name = ? and document_name = ? and parent_ID = ? ;")
		if err != nil {
			fmt.Println("Failed to prepare statement, err:" + err.Error())
			return false
		}
		defer stmt.Close()

		_, err = stmt.Exec(username, documentName, parentID)
		if err != nil {
			fmt.Println(err.Error())
			return false
		}
		UpdateDocument(username, documentName, parentID, -documentSize)
		return true
	} else {
		// 文件夹
		// 首先查找 当前 文件夹 里面 的东西
		documents, _ := GetDocumentList(username, parentID)

		if documents == nil {
			return true
		}
		for _, document := range documents {
			RemoveDocument(document.UserName, document.DocumentName,
				document.ParentID, document.IsFile, -document.DocumentSize)
		}

		// 然后删除文件夹
		stmt, err := mydb.DBConn().Prepare(
			"DELETE FROM tbl_document WHERE user_name = ? and document_name = ? and parentID = ? ;")
		if err != nil {
			fmt.Println("Failed to prepare statement, err:" + err.Error())
			return false
		}
		defer stmt.Close()

		_, err = stmt.Exec(username, documentName, parentID)
		if err != nil {
			fmt.Println(err.Error())
			return false
		}
		return true
	}

}

// UpdateDocument : 增删文件 对当前文件夹大小进行修正
func UpdateDocument(username, documentName string, parentID int, fileSize int64) bool {
	stmt, err := mydb.DBConn().Prepare(
		"UPDATE tbl_document SET document_size = document_size + ? " +
			"where user_name = ? and document_name = ? and parent_ID = ? ")
	if err != nil {
		fmt.Println("Failed to prepare statement, err:" + err.Error())
		return false
	}
	defer stmt.Close()

	_, err = stmt.Exec(fileSize, username, documentName, parentID)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	return true
}
