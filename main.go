package main

import (
	"fmt"
	"github.com/oxtoacart/bpool"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
)

const TEMPLATES_DIR = "./"
var templates map[string]*template.Template
var bufpool *bpool.BufferPool

func init() {
	bufpool = bpool.NewBufferPool(64)
	if templates == nil {
		templates = make(map[string]*template.Template)
	}

	layouts, err := filepath.Glob(TEMPLATES_DIR + "layouts/*.tmpl")
	if err != nil {
		log.Fatal(err)
	}

	includes, err := filepath.Glob(TEMPLATES_DIR + "includes/*.tmpl")
	if err != nil {
		log.Fatal(err)
	}

	for _, layout := range layouts {
		files := append(includes, layout)
		templates[filepath.Base(layout)] = template.Must(template.ParseFiles(files...))
	}

}

func renderTemplate(w http.ResponseWriter, name string, data map[string]interface{}) error {
	tmpl, ok := templates[name]
	if !ok {
		return fmt.Errorf("The template %s does not exist.", name)
	}

	// 这是个简单的写法，但是不能处理出错的情况，如果ExecuteTemplate出错，在这之前w已经写入了部分内容，此时处理错误已经来不及了。
	// 所以下面引入了bufpool来解决这个问题，它的性能也很不错的
	// w.Header().Set("Content-Type", "text/html; charset=utf-8")
	// return tmpl.ExecuteTemplate(w, "base", data)

	buf := bufpool.Get()
	defer bufpool.Put(buf)

	err := tmpl.ExecuteTemplate(buf, "base", data)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	buf.WriteTo(w)
	return nil
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request){
		data := map[string]interface{}{
			"Name": "tujiaw",
		}
		if err := renderTemplate(w, "index.tmpl", data); err != nil {
			fmt.Println(err)
		}
	})
	http.HandleFunc("/post", func(w http.ResponseWriter, r *http.Request){
		data := map[string]interface{}{
			"Content": "this is post content",
		}
		if err := renderTemplate(w, "post.tmpl", data); err != nil {
			fmt.Println(err)
		}
	})
	http.ListenAndServe(":8000", nil)
}