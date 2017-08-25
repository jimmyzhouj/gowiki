// go wiki study program
//

package main

import (
    "fmt"
    "io/ioutil"
    "net/http"
    "html/template"
    "regexp"
    //"strings"
   //errors"
)

type Page struct {
    Title string
    Body  []byte
    Disp template.HTML
}

var templates = template.Must(template.ParseFiles("template/edit.html", "template/view.html"))

var validPath = regexp.MustCompile("^/(|edit|save|view)/([a-zA-Z0-9]+)$")
var repPageName = regexp.MustCompile("\\[([a-zA-Z0-9]+)\\]")


func (p *Page) save() error {
    filename := "data/" + p.Title + ".txt"
    return ioutil.WriteFile(filename, p.Body, 0600)

}

func loadPage(title string) (*Page, error) {
    filename := "data/" + title + ".txt"
    body, err := ioutil.ReadFile(filename)
    if err != nil {
        return nil, err
    }
    return &Page{Title: title, Body: body}, nil
}

func renderTemplate(w http.ResponseWriter, tmp string, p *Page) {
    err := templates.ExecuteTemplate(w, tmp+".html", p)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

func makeHandler(fn func (http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // fmt.Println(r.URL.Path)
        if r.URL.Path == "/" {
            fn(w, r, "")
            return
        }

        m := validPath.FindStringSubmatch(r.URL.Path)
        if m == nil {
            http.NotFound(w, r)
            return
        }
        fn(w, r, m[2])
    }
}

func rootHandler(w http.ResponseWriter, r *http.Request, title string) {
    http.Redirect(w, r, "/view/FrontPage", http.StatusFound)
}

func replace(data []byte) []byte {
    fmt.Println("str is ", string(data))
    matched := repPageName.FindStringSubmatch(string(data))
    fmt.Println("matched  ", matched)
    out := []byte("<a href=\"/view/" + matched[1] + "\">" +matched[1] + "</a>")
    return out
    // str := string(data[:])
    // fmt.Println("str is ", str)    
    // ret := "<a href=\"/view/PageName\">PageName</a>"
    // ret = strings.Replace(ret, "PageName", str, -1)
    // fmt.Println("ret is ", ret)    
    // return []byte(ret)
}


func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
    p, err := loadPage(title)
    if err != nil {
        http.Redirect(w, r, "/edit/" + title, http.StatusFound)
        return
    }
    fmt.Println(p.Body)

    escape := []byte(template.HTMLEscapeString(string(p.Body)))
    //data := repPageName.ReplaceAllFunc(p.Body, replace)
    //p.Disp = template.HTML(data)
    p.Disp = template.HTML(repPageName.ReplaceAllFunc(escape, replace))
    renderTemplate(w, "view", p)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
    p, err := loadPage(title)
    if err != nil {
        p = &Page{Title: title}
    }
    renderTemplate(w, "edit", p)
}    

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
    body := r.FormValue("body")  
    p := &Page{Title:title, Body:[]byte(body)}
    err := p.save()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    http.Redirect(w, r, "/view/" + title, http.StatusFound)  
}

func main() {

    http.HandleFunc("/view/", makeHandler(viewHandler))
    http.HandleFunc("/edit/", makeHandler(editHandler))
    http.HandleFunc("/save/", makeHandler(saveHandler))
    http.HandleFunc("/", makeHandler(rootHandler))

    http.ListenAndServe(":8080", nil)
}

