package store

import (
	"database/sql"
	"fmt"
	slog "github.com/vearne/passwordbox/log"
	"github.com/vearne/passwordbox/model"
)

func CreateTable(db *sql.DB) error {
	var err error
	// table item
	sqlStmt := `
CREATE TABLE Item (
	id INTEGER PRIMARY KEY,
	title TEXT,
	IVCiphertext TEXT NOT NULL
);
	`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		slog.Error("%q: %s", err, sqlStmt)
		return err
	}
	// table meta
	sqlStmt = `
CREATE TABLE meta (
	id INTEGER PRIMARY KEY,
	hint TEXT
);
	`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		slog.Error("%q: %s", err, sqlStmt)
		return err
	}
	return nil
}

func InsertHint(db *sql.DB, hint string) error {
	tx, err := db.Begin()
	if err != nil {
		slog.Error("InsertHint, %v", err)
		return err
	}
	stmt, err := tx.Prepare("insert into meta(id, hint) values(?, ?)")
	if err != nil {
		slog.Error("InsertHint, %v", err)
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(0, hint)
	if err != nil {
		slog.Error("InsertHint, %v", err)
		return err
	}
	tx.Commit()
	return nil
}

func InsertItem(db *sql.DB, item *model.SimpleItem) error {
	tx, err := db.Begin()
	if err != nil {
		slog.Error("InsertItem, %v", err)
		return err
	}
	stmt, err := tx.Prepare("INSERT INTO item (title, IVCiphertext) VALUES(?, ?)")
	if err != nil {
		slog.Error("InsertItem, %v", err)
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(item.Title, item.IVCiphertext)
	if err != nil {
		slog.Error("InsertItem, %v", err)
		return err
	}
	tx.Commit()
	return nil
}
func UpdateItem(db *sql.DB, item *model.SimpleItem) error {
	tx, err := db.Begin()
	if err != nil {
		slog.Error("UpdateItem, %v", err)
		return err
	}
	stmt, err := tx.Prepare("update item set title = ?, IVCiphertext = ? where id = ?")
	if err != nil {
		slog.Error("UpdateItem, %v", err)
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(item.Title, item.IVCiphertext, item.ID)
	if err != nil {
		slog.Error("UpdateItem, %v", err)
		return err
	}
	tx.Commit()
	return nil
}

func CountItems(db *sql.DB, keyword string) (int, error) {
	sql := fmt.Sprintf("select count(*) from item where title like %q",
		"%"+keyword+"%")

	stmt, err := db.Prepare(sql)
	if err != nil {
		slog.Error("Get, %v", err)
		return -1, err
	}
	defer stmt.Close()
	var total int
	err = stmt.QueryRow().Scan(&total)
	if err != nil {
		slog.Error("Get, %v", err)
		return 0, err
	}
	return total, nil
}

func DeleteItem(db *sql.DB, itemId int) error {
	tx, err := db.Begin()
	if err != nil {
		slog.Error("UpdateItem, %v", err)
		return err
	}
	stmt, err := tx.Prepare("delete from item where id = ?")
	if err != nil {
		slog.Error("DeleteItem, %v", err)
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(itemId)
	if err != nil {
		slog.Error("DeleteItem, %v", err)
		return err
	}
	tx.Commit()
	return nil
}

func Query(db *sql.DB, keyword string, pageId, pageSize int) ([]*model.SimpleItem, error) {
	sql := "select id, title, IVCiphertext from item where title like %q limit %d, %d"
	sql = fmt.Sprintf(sql, "%"+keyword+"%", (pageId-1)*pageSize, pageSize)
	slog.Debug("sql:%v", sql)
	rows, err := db.Query(sql)
	if err != nil {
		slog.Error("query, %v", err)
		return nil, err
	}

	result := make([]*model.SimpleItem, 0)
	defer rows.Close()
	for rows.Next() {
		var id int
		var title string
		var IVCiphertext string
		err = rows.Scan(&id, &title, &IVCiphertext)
		if err != nil {
			slog.Error("query, %v", err)
			return nil, err
		}
		result = append(result, &model.SimpleItem{ID: id,
			Title: title, IVCiphertext: IVCiphertext})
	}

	return result, nil
}

func GetItem(db *sql.DB, itemId int) (*model.SimpleItem, error) {
	stmt, err := db.Prepare("select id, title, IVCiphertext from item where id = ?")
	if err != nil {
		slog.Error("Get, %v", err)
		return nil, err
	}
	defer stmt.Close()
	item := model.SimpleItem{}
	err = stmt.QueryRow(itemId).Scan(&item.ID, &item.Title, &item.IVCiphertext)
	if err != nil {
		slog.Error("Get, %v", err)
		return nil, err
	}
	return &item, nil
}

func GetHint(db *sql.DB) (string, error) {
	stmt, err := db.Prepare("select hint from meta where id = ?")
	if err != nil {
		slog.Error("Get, %v", err)
		return "", err
	}
	defer stmt.Close()
	var hint string
	err = stmt.QueryRow(0).Scan(&hint)
	if err != nil {
		slog.Error("Get, %v", err)
		return "", err
	}
	return hint, nil
}
