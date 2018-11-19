package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

type Verse struct {
	Book    string
	Chapter int
	Verse   int
	Text    string
}

//ParseChapterVerse extract chapter and verse from x:x format
func ParseChapterVerse(colonJoined string) (int, int) {
	fmt.Printf("colonJoined: %v\n", colonJoined)

	result := strings.Split(colonJoined, ":")
	ch, err := strconv.Atoi(result[0])
	if err != nil {
		panic(err)
	}
	v, err := strconv.Atoi(result[1])
	if err != nil {
		panic(err)
	}
	return ch, v

}

// IsNumberedBook determines if this is numbered book like 1John or 2Timothy.
func IsNumberedBook(firstPart string) bool {
	// firstPart is the very first element in the parsed string.
	if _, err := strconv.Atoi(firstPart); err == nil {
		return true
	}
	return false
}

func pullKJVText() {
	fmt.Println("Starting sqlite3 db creation. ")
	database, _ := sql.Open("sqlite3", "data/kjv.sqlite3.db")
	defer database.Close()

	statement, _ := database.Prepare("create table if not exists kjv(book string not null, chapter int, verse int, text string)")
	statement.Exec()

	url := "https://raw.githubusercontent.com/R4wm/bible/master/data/bible.txt"

	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		verse := Verse{}
		brokenString := strings.Fields(scanner.Text())
		fmt.Println("broken: ", brokenString)

		if brokenString[0] == "Song" {
			//This is Song of Solomon book, special case where book name has multiple words
			verse.Book = fmt.Sprintf("%s %s %s", brokenString[0], brokenString[1], brokenString[2])
			verse.Chapter, verse.Verse = ParseChapterVerse(brokenString[3])
			verse.Text = strings.Join(brokenString[4:], " ")
		} else if IsNumberedBook(brokenString[0]) {
			verse.Book = brokenString[0] + brokenString[1]
			verse.Chapter, verse.Verse = ParseChapterVerse(brokenString[2])
			verse.Text = strings.Join(brokenString[3:], " ")
		} else {
			verse.Book = brokenString[0]
			verse.Chapter, verse.Verse = ParseChapterVerse(brokenString[1])
			verse.Text = strings.Join(brokenString[2:], " ")
		}

		fmt.Printf("verse: %v\n", verse)
		sqlInsertStr := `INSERT OR REPLACE INTO kjv(book, chapter, verse, text) values(?, ?, ?, ?)`
		stmt, err := database.Prepare(sqlInsertStr)
		if err != nil {
			panic(err)
		}

		defer stmt.Close()
		stmt.Exec(verse.Book, verse.Chapter, verse.Verse, verse.Text)

	}

}

func main() {

	pullKJVText()

	// fmt.Println("Starting sqlite3 db creation. ")
	// database, _ := sql.Open("sqlite3", "data/kjv.sqlite3.db")
	// defer database.Close()

	// statement, _ := database.Prepare("create table if not exists kjv(book string not null, chapter int, verse int, text string)")
	// statement.Exec()

	// fmt.Println("Database done.")

}