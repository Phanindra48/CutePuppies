package main
import (
	_ "github.com/mattn/go-sqlite3"
	"log"
	"strconv"
	"database/sql"
	"os"
	"fmt"
	"strings"
)
type ImageManager struct {
	images []*Image
	db     *sql.DB
}
func NewImageManager() *ImageManager {
	return &ImageManager{}
}

// All returns the list of all the Tasks in the TaskManager.
func (m *ImageManager) All() []*Image {
	return m.images
}

func (m *ImageManager) GetPuppiesResponse(searchResponse *SearchResponse) *PuppiesResponse {
	page, err := strconv.Atoi(searchResponse.Page)
	pages, err := strconv.Atoi(searchResponse.Pages)
	perPage, err := strconv.Atoi(searchResponse.PerPage)
	total, err := strconv.Atoi(searchResponse.Total)
	if err != nil {
		log.Println(err)
		return nil
	}
	return &PuppiesResponse{page, pages, perPage, total, m.images}
}

func (m *ImageManager) Save(image *Image) error {
	for _, im := range m.images {
		if im.ID == image.ID {
			return nil
		}
	}

	m.images = append(m.images, cloneImage(image))
	return nil
}

func (m *ImageManager) GetPuppiesCount() int {
	query := "select count(id) from votes"

	rows, err := m.db.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	count := 0
	for rows.Next() {
		rows.Scan(&count)
	}

	return count
}

func (m *ImageManager) GetPuppiesByMostVotes(pageId int,uid int) []*Image {
	perPage := 20
	if pageId != 0 {
		pageId-- 
	}
	start := perPage * pageId
	query := "select v.*,coalesce(uv.choice,2) as choice" +
			" from votes v left join uservotes uv on uv.puppy_id = v.puppy_id and uv.user_id = ?" +
			" order by up_votes desc limit ?,?"
	stmt, err := m.db.Prepare(query)
	if err != nil {
		log.Fatal(err)
	}

	defer stmt.Close()
	rows, err := stmt.Query(uid,start, perPage)
	if err != nil {
		log.Fatal(err)
	}

	defer rows.Close()

	var rs []*Image

	for rows.Next() {
		var dbImage Image
		var id int
		rows.Scan(&id, &dbImage.ID, &dbImage.Title, &dbImage.Thumbnail, &dbImage.Large, &dbImage.UpVotes, &dbImage.DownVotes,&dbImage.UserChoice)
		rs = append(rs, &dbImage)
	}

	if err = rows.Err(); err != nil {
		log.Fatal(err)
	}

	return rs

}
func (m *ImageManager) NewImage(photo Photo) *Image {
	return &Image{photo.ID, photo.Title, photo.URL(SizeThumbnail), photo.URL(SizeMedium640), 0, 0, 2}
}

func (m *ImageManager) UpdateVotes(puppy_id int, up_vote bool,user_id int) {
	var choice = 0
	sqlStmt := "update votes set "
	if up_vote == true {
		sqlStmt += " up_votes = up_votes + 1"
		choice = 1
	} else {
		sqlStmt += " down_votes = down_votes + 1"
	}

	sqlStmt += " where puppy_id = ?"

	stmt, err := m.db.Prepare(sqlStmt)
	if err != nil {
		log.Fatal(err)
	}

	defer stmt.Close()

	res, _ := stmt.Exec(puppy_id)

	affect, _ := res.RowsAffected()

	fmt.Println(affect)
	fmt.Printf("update votes puppy_id: %d,up_vote: %v,uer_id: %d\n",puppy_id, choice,user_id)

	//insert user votes 
	sqlStmt = "INSERT OR REPLACE INTO uservotes (user_id,puppy_id,choice) VALUES (?,?,?)"

	stmt, err = m.db.Prepare(sqlStmt)
	if err != nil {
		log.Fatal(err)
	}

	defer stmt.Close()

	res, _ = stmt.Exec(user_id,puppy_id, choice)
	affect, _ = res.RowsAffected()

	fmt.Printf("inserted rows %d\n",affect)


	return
}

func (m *ImageManager) InitDB(removeDb bool) error {
	if removeDb == true {
		os.Remove("./" + DatabaseName)
	}

	db, err := sql.Open("sqlite3", "./"+DatabaseName)
	if err != nil {
		log.Fatal(err)
		return err
	}

	m.db = db
	return nil
}

func (m *ImageManager) GetDB() *sql.DB {
	return m.db
}

func (m *ImageManager) CreateTables() {
	fmt.Printf("create tables\n")
	createSqlStmt := `
	create table if not exists votes (id integer not null primary key, puppy_id integer unique, title string, thumbnail string, large string, up_votes integer, down_votes integer);
	delete from votes;
	`
	_, err := m.db.Exec(createSqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, createSqlStmt)
	}

	createUserTable := `
	create table if not exists users (user_id integer not null primary key,username string);
	delete from users;
	`
	_, user_err := m.db.Exec(createUserTable)
	if user_err != nil {
		log.Printf("%q: %s\n", user_err, createUserTable)
	}
	//	CREATE TABLE tblData (qid TEXT, anid INTEGER, value INTEGER, PRIMARY KEY(qid, anid))
	//CREATE UNIQUE INDEX vote_index on uservotes(user_id,puppy_id,choice);
	createUserVotesTable := `
	create table if not exists uservotes (user_id integer not null,puppy_id integer,choice int,PRIMARY KEY(user_id, puppy_id));
	delete from uservotes;
	`
	_, uservotes_err := m.db.Exec(createUserVotesTable)
	if uservotes_err != nil {
		log.Printf("%q: %s\n", uservotes_err, createUserVotesTable)
	}

}

func (m *ImageManager) InsertPuppies(images []*Image) {
	tx, err := m.db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	stmt, err := tx.Prepare("insert into votes(puppy_id, title, thumbnail, large, up_votes, down_votes) values(?, ?, ?, ?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}

	defer stmt.Close()

	for _, im := range images {
		_, err = stmt.Exec(im.ID, im.Title, im.Thumbnail, im.Large, im.UpVotes, im.DownVotes)
		if err != nil {
			log.Fatal(err)
		}
	}

	tx.Commit()
}

func (m *ImageManager) InsertUser(username string) {
	sqlStmt := "insert into users(username) select ? where not exists(select 1 from users where username = ?)"

	stmt, err := m.db.Prepare(sqlStmt)
	if err != nil {
		log.Fatal(err)
	}

	defer stmt.Close()

	res, _ := stmt.Exec(username,username)
	affect, _ := res.RowsAffected()

	fmt.Printf("inserted rows %d\n",affect)
}

func (m *ImageManager) getUser(username string) int{

	if len(username) > 0 {
		
		sql := fmt.Sprintf("select user_id from users where username = '%s' LIMIT 1",username)
		fmt.Printf("sql -> %s \n",sql)
		rows, err := m.db.Query(sql)
		if err != nil {
			log.Fatal(err)
		}

		defer rows.Close()
		var user_id int
		for rows.Next() {
			rows.Scan(&user_id)
		}

		fmt.Printf("user %s id %v\n",username,user_id)

		return user_id
	} else {
		return 0
	}
}


func (m *ImageManager) FindOldPuppies(ids []string,uid int) []*Image {

	//sqlStmt = "select * from votes where puppy_id in (?" + strings.Repeat(",?", len(ids)-1) + ")"

	query := fmt.Sprintf("select v.*,coalesce(uv.choice,2) as choice" +
		" from votes v left join uservotes uv on uv.puppy_id = v.puppy_id and uv.user_id = ?" +
		" where v.puppy_id in (%s)",
		strings.Join(strings.Split(strings.Repeat("?", len(ids)), ""), ","))
	fmt.Println("query")
	stmt, err := m.db.Prepare(query)
	if err != nil {
		log.Fatal(err)
	}

	var params []interface{}
	//append user id first
	params = append(params, uid)
	for _, id := range ids {
		params = append(params, id)
	}
	defer stmt.Close()
	rows, err := stmt.Query(params...)
	if err != nil {
		log.Fatal(err)
	}

	defer rows.Close()

	var rs []*Image

	for rows.Next() {
		var dbImage Image
		var id int
		rows.Scan(&id, &dbImage.ID, &dbImage.Title, &dbImage.Thumbnail, &dbImage.Large, &dbImage.UpVotes, &dbImage.DownVotes, &dbImage.UserChoice)
		rs = append(rs, &dbImage)
	}

	if err = rows.Err(); err != nil {
		log.Fatal(err)
	}

	return rs
}
