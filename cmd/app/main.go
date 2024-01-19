package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

type album struct {
	ID     string `json:"id"`
	Title  string `json:"Title"`
	Artist string `json:"artist"`
	Price  string `json:"price"`
}

var albums = []album{
	{ID: "1", Title: "Jeru", Artist: "Jacob Benjamin Gyllenhaal", Price: "100"},
	{ID: "2", Title: "Jeru2", Artist: "Jacob Benjamin Gyllenhaal2", Price: "1002"},
}

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func pregister(c *gin.Context) {
	err := godotenv.Load("../../internal/database.env")
	if err != nil {
		log.Println(":", err)
	}
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	db, err := sql.Open("mysql", dbUser+":"+dbPassword+"@tcp(127.0.0.1:3306)/"+dbName)
	if err != nil {
		log.Println("Error connecting to the database:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
	}
	defer db.Close()
	var newUser User
	if err := c.ShouldBindJSON(&newUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error2": err.Error()})
		return
	}
	if err := registerUser(db, newUser); err != nil {
		log.Println("Error registering user:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully"})
}

func registerUser(db *sql.DB, newUser User) error {
	statement, err := db.Prepare("Insert into accounts (login, password) values (?,?)")
	if err != nil {
		log.Println("Error with preparing quiry", err)
		return err
	}
	_, err = statement.Exec(newUser.Username, newUser.Password)
	defer statement.Close()
	if err != nil {
		log.Println("Error with executing quiry", err)
		return err
	}
	log.Println("User was successfully registered")
	log.Println(newUser.Username, newUser.Password)
	return nil
}

func plogin(c *gin.Context) {
	err := godotenv.Load("../../internal/database.env")
	if err != nil {
		log.Println(":", err)
	}
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	db, err := sql.Open("mysql", dbUser+":"+dbPassword+"@tcp(127.0.0.1:3306)/"+dbName)
	if err != nil {
		log.Println("Error connecting to the database:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
	}
	defer db.Close()
	var oldUser User
	if err := c.ShouldBindJSON(&oldUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error2": err.Error()})
		return
	}
	if err := findUser(db, oldUser); err != nil {
		if _, ok := err.(*PasswordIsWrong); ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Password is incorrect"})
		}
		log.Println("Error finding user:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	session := sessions.Default(c)
	session.Set("user", oldUser.Username)
	err = session.Save()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save session"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully"})
}

type PasswordIsWrong struct{}

func (e *PasswordIsWrong) Error() string {
	return "Password is wrong"
}

func findUser(db *sql.DB, oldUser User) error {

	stmt, err := db.Prepare("SELECT password FROM accounts WHERE login = ?")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	var hashedPassword string
	// Query the database
	err = stmt.QueryRow(oldUser.Username).Scan(&hashedPassword)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Println("No user with that username.")
			return &PasswordIsWrong{}
		}
		log.Fatal(err)
	}

	if oldUser.Password != hashedPassword {
		log.Println("The password is incorrect.")
		return &PasswordIsWrong{}
	}

	return nil
}

func gregister(c *gin.Context) {
	c.HTML(http.StatusOK, "form.html", gin.H{"title": "Start"})
}

func glogin(c *gin.Context) {
	c.HTML(http.StatusOK, "form2.html", gin.H{"title": "Start2"})
}

func gmain(c *gin.Context) {

	if !userIsLoggedIn(c) {
		c.Redirect(http.StatusTemporaryRedirect, "/login")
		c.Abort()
	} else {
		// User is authenticated, let them proceed
		c.HTML(http.StatusOK, "main.html", gin.H{"title": "Start3"})
	}
}
func pmain(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	if err := session.Save(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save session"})
		return
	}
	c.Redirect(http.StatusFound, "/login")
}
func userIsLoggedIn(c *gin.Context) bool {
	session := sessions.Default(c)
	user := session.Get("user")
	return user != nil
}

func postAlbum(c *gin.Context) {
	var newAlbum album
	if err := c.BindJSON(&newAlbum); err != nil {
		return
	}
	albums = append(albums, newAlbum)
	c.IndentedJSON(http.StatusCreated, newAlbum)
}

func getAlbum(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, albums)
}

func getAlbumById(c *gin.Context) {
	id := c.Param("id")
	for _, a := range albums {
		if a.ID == id {
			c.IndentedJSON(http.StatusOK, a)
			return
		}
	}
	c.IndentedJSON(http.StatusNotFound, gin.H{"answer": "no album with such ID"})
}

func main() {
	r := gin.Default()
	store := cookie.NewStore([]byte("secret"))
	fmt.Println(&store)
	r.Use(sessions.Sessions("mysession", store))
	r.LoadHTMLGlob("D:/vscodeprojects/golang/project1/internal/templates/*")
	r.POST("/register", pregister)
	r.GET("/register", gregister)
	r.GET("/login", glogin)
	r.GET("/albums/:id", getAlbumById)
	r.GET("/albums", getAlbum)
	r.POST("/albums", postAlbum)
	r.POST("/login", plogin)
	r.GET("/main", gmain)
	r.POST("/main", pmain)
	r.Run()
}
