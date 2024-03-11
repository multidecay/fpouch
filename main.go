package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

// Confing struct
type Conf struct {
	NoUpload  bool
	NoSharing bool
	NoUI      bool
	StorePath string
}

func (c Conf) print() {
	fmt.Printf("store_path : %s \n", c.StorePath)
}

// ENDPOINT

func uploadUi() {

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

func indexUi() {

}

func downloadFile() {

}

// ROUTE

func setupRoutes(conf *Conf) {
	if !conf.NoUpload {
		http.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case "GET":
				if !conf.NoUI {
					w.Write([]byte("upload ui"))
					return
				}
				w.WriteHeader(404)
			case "POST":
				uploadStore(w, r, conf)
				w.Write([]byte("upload endpoint"))
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
					w.Write([]byte("upload ui"))
					return
				}
				w.WriteHeader(404)
			default:
				w.WriteHeader(404)
			}
		})
	}

	fmt.Println("fpouch - pouch for file upload and sharing, starting...")
	http.ListenAndServe(":4444", nil)
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
	flag.Parse()

	// clean the path from os path difference
	c.StorePath = filepath.Clean(c.StorePath)

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
