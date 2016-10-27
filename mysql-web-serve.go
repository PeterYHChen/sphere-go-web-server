package main

import (
    //"fmt"
    "strconv"
    "net/http"
    "html/template"
    "log"
    "database/sql"
    _ "github.com/go-sql-driver/mysql"
)

type Data struct {
    Id string 
    Content string
}

const (
    DB_HOST = "tcp(127.0.0.1:3306)"
    DB_NAME = "sphere"
    DB_USER = "root"
    DB_PASS = "gaojingwen"
    DB_DATA_TABLE = "dataSet"
)

var db *sql.DB
// save data
// func (data *Data) save() error {

// }

// var templates = template.Must(template.ParseFiles("edit.html", "view.html"))

func loadData(id string) (*[]Data, error) {
    var idInt int
    var err error
    var rows *sql.Rows
    if id != "" {
        idInt, err = strconv.Atoi(id);
        rows, err = db.Query("select * from " + DB_DATA_TABLE + " where id = ?", idInt)
    } else {
        rows, err = db.Query("select * from " + DB_DATA_TABLE)
    }
    if err != nil {
        log.Fatal(err)
    }
    defer rows.Close()
    
    var dataSet []Data
    var content string
    for rows.Next() {
        err := rows.Scan(&idInt, &content)
        if err != nil {
            log.Fatal(err)
        }
        dataSet = append(dataSet, Data{Id: strconv.Itoa(idInt), Content: content})
    }
    err = rows.Err()
    if err != nil {
        log.Fatal(err)
    }

    return &dataSet, nil
}

func renderTemplate(w http.ResponseWriter, tmpl string, dataSet *[]Data) {
    t, err := template.ParseFiles(tmpl + ".html")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    err = t.Execute(w, dataSet)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

// edit the Content and store in database
func editHandler(w http.ResponseWriter, r *http.Request) {
    id := r.URL.Path[len("/edit/"):]
    data, _ := loadData(id)

    renderTemplate(w, "edit", data)
}

// view data from database
func viewHandler(w http.ResponseWriter, r *http.Request) {
    id := r.URL.Path[len("/view/"):]
    data, _ := loadData(id)
    renderTemplate(w, "view", data)
}

func saveHandler(w http.ResponseWriter, r *http.Request) {
    // id := r.URL.Path[len("/save/"):]
    // content := r.FormValue("content")
    // data := &Data{Id: id, Content: content}
   // data.save()
    // http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

func main() {
    var err error
    dns := DB_USER + ":" + DB_PASS + "@" + DB_HOST + "/"
    db, err = sql.Open("mysql", dns)
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    _, err = db.Exec("CREATE DATABASE IF NOT EXISTS "+ DB_NAME)
    if err != nil {
        log.Fatal(err)
    }

    _, err = db.Exec("USE " + DB_NAME)
    if err != nil {
        log.Fatal(err)
    }

    _, err = db.Exec("CREATE TABLE IF NOT EXISTS " + DB_DATA_TABLE + 
        " (id int(11) NOT NULL AUTO_INCREMENT, PRIMARY KEY (id), content TEXT NOT NULL)")
    if err != nil {
        log.Fatal(err)
    }

    http.HandleFunc("/view/", viewHandler)
    http.HandleFunc("/edit/", editHandler)
    http.ListenAndServe(":8080", nil)
}