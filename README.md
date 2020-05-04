# tmpl
golang html/template模板继承实例

golang的模板包使用的是html/template库，使用的时候通常首先我们就会关注它的模板继承（模板嵌套）怎么写，
毕竟这影响到整体网页渲染接口的写法，以及接口是否优雅和可扩展性。

# base.tmpl
首先，我们定义一个基础模板layouts/base.tmpl
```cassandraql
{{ define "base" }}
<html>
<head>
    {{ template "title" . }}
</head>
<body>
    {{ template "content" . }}
</body>
</html>
{{ end }}

{{ define "title" }}<title>default title</title>{{ end }}
{{ define "content" }}<h1>default body</h1>{{ end }}
```
base模板中包含两个子模板分别是title和content，同时定义了它们的默认值，这个是为了防止子模板没有被重定义的时候运行出错。
如果没有定义默认模板，那么它的扩展模板必须要定义它。

# index.tmpl
定义一个主页includes/index.tmpl，它是基础模板的扩展
```cassandraql

{{ define "title"}}<title>Index Title-{{ .Name }}</title>{{ end }}
```
这里只是简单的定义了title，同时通过参数传入Name，由于没有重定义content，所以默认显示base中的content

# post.tmpl
定义一个显示文章内容的模板includes/post.tmpl
```cassandraql
{{ define "content"}}
<h1>Post Header</h1>
<div>{{ .Content }}</div>
{{ end }}
```
重新定义了content，通过参数传入Content，由于没有重定义title，所以默认显示base中的title

# 渲染接口
```cassandraql
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
```

# 测试
测试代码main函数如下：
```cassandraql
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
```
在本地浏览器地址栏输入主页地址：  
http://localhost:8000/  
浏览器上输出：  
标题是：Index Title-tujiaw  
内容是：default body  

在本地浏览器地址栏输入post地址：  
http://localhost:8000/post  
浏览器上输出：  
标题是：default title  
内容是：Post Header this is post content  
