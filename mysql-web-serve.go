package main

import (
    //"fmt"
    "strconv"
    "net/http"
    "html/template"
    "log"
    "database/sql"
    _ "github.com/go-sql-driver/mysql"
    // "errors"
)

type Data struct {
    Id int 
    Content string
}

const (
    DB_HOST = "tcp(127.0.0.1:3306)"
    DB_NAME = "sphere"
    DB_USER = "root"
    DB_PASS = "gaojingwen"
    DB_DATA_TABLE = "dataSet"
    DB_DATA_TABLE_COLUMN_CONTENT = "content"
)

var db *sql.DB

// save data
func (data *Data) save() error {
    stmt1, err := db.Prepare("INSERT INTO " + DB_DATA_TABLE + 
        "(id, " + DB_DATA_TABLE_COLUMN_CONTENT + ") VALUES(?, ?)" + 
        " ON DUPLICATE KEY UPDATE " + DB_DATA_TABLE_COLUMN_CONTENT + "= ?")
    if err != nil {
        log.Fatal(err)
    }
    defer stmt1.Close()

    _, err = stmt1.Exec(data.Id, data.Content, data.Content)
    if err != nil {
        log.Fatal(err)
    }
    return nil
}

// var templates = template.Must(template.ParseFiles("edit.html", "view.html"))

func loadData(id int) (*Data, error) {
    var content string
    err := db.QueryRow("SELECT " + DB_DATA_TABLE_COLUMN_CONTENT + 
        " FROM " + DB_DATA_TABLE + " where id = ?", id).Scan(&content)
    return &Data{Id : id, Content : content}, err 
}

func loadAllData() (*[]Data, error) {
    rows, err := db.Query("SELECT * FROM " + DB_DATA_TABLE)
    if err != nil {
        log.Fatal(err)
    }
    defer rows.Close()
    
    var dataSet []Data
    var id int
    var content string
    for rows.Next() {
        err := rows.Scan(&id, &content)
        if err != nil {
            log.Fatal(err)
        }
        dataSet = append(dataSet, Data{Id: id, Content: content})
    }
    err = rows.Err()
    if err != nil {
        log.Fatal(err)
    }

    return &dataSet, nil
}

func renderTemplateShowAll(w http.ResponseWriter, tmpl string, dataSet *[]Data) {
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

func renderTemplate(w http.ResponseWriter, tmpl string, data *Data) {
    t, err := template.ParseFiles(tmpl + ".html")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    err = t.Execute(w, data)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

func handler(w http.ResponseWriter, r *http.Request) {
    http.Redirect(w, r, "/view/", http.StatusFound)
}

func addHandler(w http.ResponseWriter, r *http.Request) {
    var cnt int
    _ = db.QueryRow("SELECT COUNT(*) FROM " + DB_DATA_TABLE).Scan(&cnt)

    nextId := strconv.Itoa(cnt+1)
    http.Redirect(w, r, "/edit/" + nextId, http.StatusFound)
}

// edit the Content and store in database
func editHandler(w http.ResponseWriter, r *http.Request) {
    idstring := r.URL.Path[len("/edit/"):]
    id, err := strconv.Atoi(idstring)
    if err != nil {
        renderTemplate(w, "empty", nil)
    } else {
        data, err := loadData(id)
        if err == nil || err == sql.ErrNoRows {
            renderTemplate(w, "edit", data)
        } else {
            renderTemplate(w, "empty", nil)
        }
    }
}

// view data from database
func viewHandler(w http.ResponseWriter, r *http.Request) {
    idstring := r.URL.Path[len("/view/"):]

    // view specific data
    if idstring != "" {
        id, err := strconv.Atoi(idstring)
        if err != nil {
            renderTemplate(w, "empty", nil)
            return
        }
        data, err := loadData(id)
        if err == nil {
            renderTemplate(w, "view", data)
        } else if err == sql.ErrNoRows {
            renderTemplate(w, "not-found", data)
        } else {
            renderTemplate(w, "empty", nil)
        }
    // view data set
    } else {
        dataSet, _ := loadAllData()
        renderTemplateShowAll(w, "viewAll", dataSet)
    }
}

func saveHandler(w http.ResponseWriter, r *http.Request) {
    idstring := r.URL.Path[len("/save/"):]
    id, err := strconv.Atoi(idstring)
    if err != nil {
        renderTemplate(w, "empty", nil)
        return
    }
    content := r.FormValue("content")
    data := &Data{Id: id, Content: content}
    err = data.save()
    if err != nil {
        renderTemplate(w, "not-found", data)
    } else {
        http.Redirect(w, r, "/view/" + idstring, http.StatusFound)
    }
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

    http.HandleFunc("/", handler)
    http.HandleFunc("/view/", viewHandler)
    http.HandleFunc("/edit/", editHandler)
    http.HandleFunc("/add/", addHandler)
    http.HandleFunc("/save/", saveHandler)
    http.ListenAndServe(":8080", nil)
}