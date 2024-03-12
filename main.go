package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

// Confing struct
type Conf struct {
	NoUpload  bool
	NoSharing bool
	NoUI      bool
	Port      int
	StorePath string
}

// ENDPOINT

func uploadUi(w http.ResponseWriter, r *http.Request) {
	layout := `<!DOCTYPE html>
	<html lang="en">
	<head>
		<meta charset="UTF-8">
		<meta http-equiv="X-UA-Compatible" content="IE=edge">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<link rel="shortcut icon" href="#" />
	</head>
		<body>
			{{. }}
		</body>
	</html>
	`
	tpl, err := template.New("fpouch").Parse(layout)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(500)
		return
	}

	payload := `<form method="POST" action="/upload" enctype="multipart/form-data" style="display:grid; gap: 50; margin: auto;">
		<section id="uploads" style="display: grid; gaps: 50;">
			<input name="files" type="file" placeholder="Put file here" required/>
		</section>
		<button type="button" style="background:white; color: navy; border: 1px solid navy;" id="addUpload">Add upload</button>
		<button type="submit" style="margin-top: .5em;">Upload now !</button>
	</form>
	<script>

		function newUploadInput(){
			var x = document.createElement("input");
			x.name = "files";
			x.type = "file";
			x.required = true;
			return(x);
		}
		const addUploadButton = document.getElementById('addUpload');
		const uploadForm = document.getElementById('uploads');

		addUploadButton.addEventListener('click', ()=>{
			uploadForm.insertBefore(newUploadInput(),null);
		})

	</script>
	`

	tpl.Execute(w, template.HTML(payload))
}

func uploadStore(w http.ResponseWriter, r *http.Request, c *Conf) {
	r.ParseMultipartForm(10 << 20)

	for _, handler := range r.MultipartForm.File["files"] {
		tempFile, err := os.CreateTemp(c.StorePath, handler.Filename+"-*")
		if err != nil {
			fmt.Println(err)
			break
		}

		file, err := handler.Open()
		if err != nil {
			fmt.Println(err)
			break
		}

		fileBytes, err := io.ReadAll(file)
		if err != nil {
			fmt.Println(err)
		}

		tempFile.Write(fileBytes)

	}
	fmt.Fprintf(w, "Successfully Uploaded File\n")
}

func indexUi(w http.ResponseWriter, r *http.Request, c *Conf) {
	var files []string

	filepath.Walk(c.StorePath, func(path string, info fs.FileInfo, err error) error {
		if !info.IsDir() {
			files = append(files, filepath.Base(path))
		}
		return nil
	})

	if c.NoUI {
		res, err := json.Marshal(files)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(500)
			return
		}

		w.Header().Add("Content-type", "application/json")
		w.Write(res)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	layout := `<!DOCTYPE html>
	<html lang="en">
	<head>
		<meta charset="UTF-8">
		<meta http-equiv="X-UA-Compatible" content="IE=edge">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<link rel="shortcut icon" href="#" />
	</head>
		<body>
			{{. }}
		</body>
	</html>
	`

	tpl, err := template.New("fpouch").Parse(layout)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(500)
		return
	}

	payload := "<section style='display: grid; gap: .5em;'>\n"
	for _, file := range files {
		payload += fmt.Sprintf("<a href='/%s'>%s</a> \n", file, file)
	}
	payload += "</section>"

	tpl.Execute(w, template.HTML(payload))
}

// ROUTE

func setupRoutes(conf *Conf) {
	if !conf.NoUpload {
		http.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case "GET":
				if !conf.NoUI {
					uploadUi(w, r)
					return
				}
				w.WriteHeader(404)
			case "POST":
				uploadStore(w, r, conf)
			default:
				w.WriteHeader(404)
			}
		})
	}

	if !conf.NoSharing {

		http.HandleFunc("/index", func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case "GET":
				if !conf.NoUI {
					indexUi(w, r, conf)
					return
				}
				w.WriteHeader(404)
			default:
				w.WriteHeader(404)
			}
		})

		fs := http.FileServer(http.Dir(conf.StorePath))
		http.Handle("/", http.StripPrefix("/", fs))
	}
	port := strconv.Itoa(conf.Port)
	fmt.Println("fpouch - pouch for file upload and sharing, starting at port " + port)
	http.ListenAndServe(":"+port, nil)
}

// UTILS

func IsPathNotExists(path string) bool {
	_, err := os.Stat(path)
	return os.IsNotExist(err)
}

// options: no-ui, no-sharing, no-upload, set-password=[password]
func main() {
	wd, _ := os.Getwd()
	c := Conf{}

	flag.BoolVar(&c.NoSharing, "no-sharing", false, "disable index file.")
	flag.BoolVar(&c.NoUpload, "no-upload", false, "disable file upload.")
	flag.BoolVar(&c.NoUI, "no-ui", false, "disable upload and sharing will json.")
	flag.StringVar(&c.StorePath, "store-path", wd, "place for store uploaded file and indexed for share.")
	flag.IntVar(&c.Port, "port", 6942, "port to listen")
	flag.Parse()

	// clean the path from os path difference
	c.StorePath = filepath.Clean(c.StorePath)

	// prevent negative number
	c.Port = int(uint16(c.Port))
	if c.Port == 0 {
		c.Port = 6942
	}

	// if the path not exist then create one
	if IsPathNotExists(c.StorePath) {
		os.MkdirAll(c.StorePath, os.ModePerm)
	}

	if c.NoSharing && c.NoUpload {
		fmt.Println("no-sharing and no-upload cannot exist together")
		return
	}

	setupRoutes(&c)
}
