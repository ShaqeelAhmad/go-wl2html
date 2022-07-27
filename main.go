package main

import (
	"encoding/xml"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const outdir = "./doc"

type document struct {
	Name      string `xml:"name,attr"`
	Copyright string `xml:"copyright"`
	Interface []struct {
		Name        string `xml:"name,attr"`
		Version     string `xml:"version,attr"`
		Description struct {
			Summary string `xml:"summary,attr"`
			Content string `xml:",chardata"`
		} `xml:"description"`
		Request []struct {
			Name        string `xml:"name,attr"`
			Description struct {
				Summary string `xml:"summary,attr"`
				Content string `xml:",chardata"`
			} `xml:"description"`
			Arg []struct {
				Name      string `xml:"name,attr"`
				Type      string `xml:"type,attr"`
				Interface string `xml:"interface,attr"`
				Summary   string `xml:"summary,attr"`
			} `xml:"arg"`
		} `xml:"request"`
		Event []struct {
			Name        string `xml:"name,attr"`
			Description struct {
				Summary string `xml:"summary,attr"`
				Content string `xml:",chardata"`
			} `xml:"description"`
			Arg []struct {
				Name      string `xml:"name,attr"`
				Type      string `xml:"type,attr"`
				Interface string `xml:"interface,attr"`
				Summary   string `xml:"summary,attr"`
			} `xml:"arg"`
		} `xml:"event"`
		Enum []struct {
			Name        string `xml:"name,attr"`
			Description struct {
				Summary string `xml:"summary,attr"`
				Content string `xml:",chardata"`
			} `xml:"description"`
			Entry []struct {
				Name    string `xml:"name,attr"`
				Value   string `xml:"value,attr"`
				Summary string `xml:"summary,attr"`
			} `xml:"entry"`
		} `xml:"enum"`
	} `xml:"interface"`
}

func generateIndex() {
	files, err := os.ReadDir(outdir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	idx := []struct {
		Name string
		Path string
	}{}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if file.Name() == "index.html" || file.Name() == "style.css" {
			continue
		}

		idx = append(idx, struct {
			Name string
			Path string
		}{
			Name: strings.TrimSuffix(file.Name(), filepath.Ext(file.Name())),
			Path: "/" + file.Name(),
		})
	}

	f, err := os.Create(filepath.Join(outdir, "index.html"))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer f.Close()

	tmpl.Execute(f, idx)
}

func handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	filename := outdir + r.URL.String()
	if r.URL.String() == "/" {
		filename = filepath.Join(outdir, "index.html")
	}
	file, err := os.ReadFile(filename)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404 error " + err.Error()))
		log.Println(err)
		return
	}

	ext := filepath.Ext(filename)
	switch ext {
	case ".css":
		w.Header().Set("Content-type", "text/css")
	case ".html":
		w.Header().Set("Content-type", "text/html")
	case ".jpg", ".jpeg":
		w.Header().Set("Content-type", "image/jpeg")
	case ".png":
		w.Header().Set("Content-type", "image/png")
	case ".gif":
		w.Header().Set("Content-type", "image/gif")
	}

	w.Write(file)
}

func serveFiles() {
	generateIndex()
	http.HandleFunc("/", handler)

	serveSite := "localhost:8000"
	fmt.Printf("Starting server at http://%s\n", serveSite)

	log.Fatal(http.ListenAndServe(serveSite, nil))
}

func main() {
	if len(os.Args) < 2 {
		serveFiles()
		return
	}

	if (os.Args[1] == "help" || os.Args[1] == "-h" || os.Args[1] == "-help") {
		fmt.Printf("usage: %s <input-file> [output-file]\n", os.Args[0])
		os.Exit(0)
	}

	filename := os.Args[1]

	outfile := ""
	if len(os.Args) >= 3 {
		outfile = os.Args[2]
	} else {
		os.MkdirAll(outdir, 0755)
		outfile = filepath.Base(filename)
		outfile = strings.TrimSuffix(outfile, filepath.Ext(outfile)) + ".html"
		outfile = filepath.Join(outdir, outfile)
	}

	file, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	doc := document{}
	err = xml.Unmarshal(file, &doc)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	tmpl, err := template.ParseFiles("templates/protocol.html")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	f, err := os.Create(outfile)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer f.Close()

	tmpl.Execute(f, doc)
}
