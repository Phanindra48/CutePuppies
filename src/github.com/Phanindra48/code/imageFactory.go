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

func (m *ImageManager) NewImage(photo Photo) *Image {
	return &Image{photo.ID, photo.Title, photo.URL(SizeThumbnail), photo.URL(SizeMedium640), 0, 0}
}

func (m *ImageManager) UpdateVotes(puppy_id int, up_vote bool) {
	sqlStmt := "update votes set "
	if up_vote == true {
		sqlStmt += " up_votes = up_votes + 1"
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
	createSqlStmt := `
	create table if not exists votes (id integer not null primary key, puppy_id integer unique, title string, thumbnail string, large string, up_votes integer, down_votes integer);
	delete from votes;
	`
	_, err := m.db.Exec(createSqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, createSqlStmt)
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

func (m *ImageManager) FindOldPuppies(ids []string) []*Image {

	//sqlStmt = "select * from votes where puppy_id in (?" + strings.Repeat(",?", len(ids)-1) + ")"

	query := fmt.Sprintf("select * from votes where puppy_id in (%s)",
		strings.Join(strings.Split(strings.Repeat("?", len(ids)), ""), ","))

	stmt, err := m.db.Prepare(query)
	if err != nil {
		log.Fatal(err)
	}

	var params []interface{}
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
		rows.Scan(&id, &dbImage.ID, &dbImage.Title, &dbImage.Thumbnail, &dbImage.Large, &dbImage.UpVotes, &dbImage.DownVotes)
		rs = append(rs, &dbImage)
	}

	if err = rows.Err(); err != nil {
		log.Fatal(err)
	}

	return rs
}
