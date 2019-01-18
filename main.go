package main

import (
	"database/sql"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/pusher/pusher-http-go"
	"net/http"
	_ "github.com/mattn/go-sqlite3"
)


func initialiseDatabase(filepath string) *sql.DB  {
	db, err := sql.Open("sqlite3",filepath)

	if err != nil{
		panic(err)
	}

	if db == nil{
		panic("db nil")
	}

	return db
}

func migrateDatabase(db *sql.DB)  {
	sql := `CREATE TABLE IF NOT EXISTS posts(
                    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
                    content TEXT
            );`

	_, err := db.Exec(sql)
	if err != nil{
		panic(err)
	}
}

var client = pusher.Client{
	AppId:   "693970",
	Key:     "d2f8e5e37e867c18b9e8",
	Secret:  "40c563f87ae42977d316",
	Cluster: "ap1",
	Secure:  true,
}

type Post struct {

	ID int64 `json:"id"`
	Content string `json:"content"`
}

type PostCollection struct {
	Posts []Post `json:"item"`
}

func getPosts(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		rows, err := db.Query("SELECT * FROM posts ORDER BY id DESC")
		if err != nil {
			panic(err)
		}

		defer rows.Close()

		result := PostCollection{}

		for rows.Next() {
			post := Post{}
			err2 := rows.Scan(&post.ID, &post.Content)
			if err2 != nil {
				panic(err2)
			}

			result.Posts = append(result.Posts, post)
		}

		return c.JSON(http.StatusOK, result)
	}
}

func savePost(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		postContent := c.FormValue("content")
		stmt, err := db.Prepare("INSERT INTO posts (content) VALUES(?)")
		if err != nil {
			panic(err)
		}

		defer stmt.Close()

		result, err := stmt.Exec(postContent)
		if err != nil {
			panic(err)
		}

		insertedID, err := result.LastInsertId()
		if err != nil {
			panic(err)
		}

		post := Post{
			ID:      insertedID,
			Content: postContent,
		}

		client.Trigger("live-blog-stream", "new-post", post)

		return c.JSON(http.StatusOK, post)
	}
}

func main()  {

	e:= echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	db:= initialiseDatabase("./database/storage.db")
	migrateDatabase(db)

	e.File("/","public/index.html")
	e.File("/admin","public/admin.html")
	e.GET("/posts",getPosts(db))
	e.POST("/posts",savePost(db))

	e.Logger.Fatal(e.Start(":9000"))

}