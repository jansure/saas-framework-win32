package main

import (
	"flag"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"path"
	"path/filepath"
	"strings"
)

var (
	dir                 string
	port                string
	logging             bool
	depth               int
	auth                string
	debug               bool
	disable_sys_command bool
	rootDir             = "."
)

//var cpuprof string
//commandsFile        string

const MAX_MEMORY = 1 * 1024 * 1024
const VERSION = "1.0"

func main() {

	//fmt.Println(len(os.Args), os.Args)
	if len(os.Args) > 1 && os.Args[1] == "-v" {
		fmt.Println("Version " + VERSION)
		os.Exit(0)
	}

	flag.StringVar(&dir, "dir", "/tmp", "Specify a directory to server files from.")
	flag.StringVar(&port, "port", ":8080", "Port to bind the file server")
	flag.BoolVar(&logging, "log", true, "Enable Log (true/false)")
	flag.StringVar(&auth, "auth", "", "'username:pass' Basic Auth")
	flag.IntVar(&depth, "depth", 5, "Depth directory crawler")
	//flag.StringVar(&commandsFile, "commands", "", "Path to external commands file.json")
	flag.BoolVar(&debug, "debug", false, "Make external assets expire every request")
	flag.BoolVar(&disable_sys_command, "disable_cmd", true, "Disable sys comands")

	//flag.StringVar(&cpuprof, "cpuprof", "", "write cpu and mem profile")
	flag.Parse()

	envDir := os.Getenv("FILESERVER_DIR")
	if envDir != "" {
		dir = envDir
	}
	envPort := os.Getenv("FILESERVER_PORT")
	if envPort != "" {
		port = envPort
	}
	envAuth := os.Getenv("FILESERVER_AUTH")
	if envAuth != "" {
		auth = envAuth
	}
	envCmd := os.Getenv("FILESERVER_COMMAND")
	if envCmd != "" {
		disable_sys_command = false
	}

	if logging == false {
		log.SetOutput(ioutil.Discard)
	}
	// If no path is passed to app, normalize to path formath
	if dir == "." {
		dir, _ = filepath.Abs(dir)
	}

	if _, err := os.Stat(dir); err != nil {
		log.Fatalf("Directory %s not exist", dir)
	}

	// normalize dir, ending with... /
	if strings.HasSuffix(dir, "/") == false {
		dir = dir + "/"
	}

	// build index files in background
	go Build_index(dir)

	mux := http.NewServeMux()

	statics := &ServeStaticFromBinary{
		MountPoint: "/-/assets/",
		DataDir:    "data/"}

	mux.Handle("/-/assets/", makeGzipHandler(statics.ServeHTTP))

	mux.Handle("/-/api/dirs", makeGzipHandler(http.HandlerFunc(SearchHandle)))
	mux.Handle("/", BasicAuth(http.HandlerFunc(handleReq), auth))

	log.Printf("Listening on port %s .....\n", port)
	if debug {
		log.Print("Serving data dir in debug mode.. no assets caching.\n")
	}
	http.ListenAndServe(port, mux)

}

func handleReq(w http.ResponseWriter, r *http.Request) {

	//Is_Ajax := strings.Contains(r.Header.Get("Accept"), "application/json")
	if r.Method == "PUT" {
		AjaxUpload(w, r)
		return
	}
	if r.Method == "POST" {
		WebCommandHandler(w, r)
		return
	}

	log.Print("Request: ", r.RequestURI)
	// See bug #9. For some reason, don't arrive index.html, when asked it..
	if strings.HasSuffix(r.URL.Path, "/") && r.FormValue("get_file") != "true" {
		log.Printf("Index dir %s", r.URL.Path)
		handleDir(w, r)
	} else {
		log.Printf("downloading file %s", path.Clean(dir+r.URL.Path))
		r.Header.Del("If-Modified-Since")
		http.ServeFile(w, r, path.Clean(dir+r.URL.Path))
		//http.ServeContent(w, r, r.URL.Path)
		//w.Write([]byte("this is a test inside file handler"))

	}

}

func handleDir(w http.ResponseWriter, r *http.Request) {

	var d string = ""
	//var s string = ""
	//log.Printf("len %d,, %s", len(r.URL.Path), dir)
	if len(r.URL.Path) == 1 {
		// handle root dir
		d = dir
	} else {
		d += dir + r.URL.Path[1:]
	}

	// handle json format of dir...
	if r.FormValue("format") == "json" {

		w.Header().Set("Content-Type", "application/json")
		result := &DirJson{w, d}
		err := result.Get()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	if r.FormValue("format") == "zip" {
		result := &DirZip{w, d}
		err := result.Get()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	/*if r.FormValue("format") == "unzip" {
		result := &DirUnZip{w, s,d}
		err := result.()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}*/

	// If we dont receive json param... we are asking, for genric app ui...
	template_file, err := Asset("data/main.html")
	if err != nil {
		log.Fatalf("Cant load template main")
	}

	t := template.Must(template.New("listing").Delims("[%", "%]").Parse(string(template_file)))
	v := map[string]interface{}{
		"Path":        r.URL.Path,
		"version":     VERSION,
		"sys_command": disable_sys_command,
	}
	w.Header().Set("Content-Type", "text/html")
	t.Execute(w, v)

}

func upLoadAndUnzip(w http.ResponseWriter, r *http.Request) {

	var destPath = "/tmp/"
	//var destPath = r.Header.Get("destPath")
	if r.Method == "GET" {
		fmt.Fprintln(w, "滚蛋")
		return
	}

	err := r.ParseMultipartForm(1024 * 1024 * 100)
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	file, handler, err := r.FormFile("uploadfile")
	defer file.Close()
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}

	var dir  = flag.String("d", destPath, "location dir")
	os.Mkdir(*dir, 0777)

	log.Println(handler.Filename)

	f, err := os.OpenFile(*dir+handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
	defer f.Close()
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}

	io.Copy(f, file)

	err = DeCompress(f.Name(), *dir)
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}

	os.Remove(f.Name())

	fmt.Fprintln(w, "upload ok!")
}


func AjaxUpload(w http.ResponseWriter, r *http.Request) {
	reader, err := r.MultipartReader()
	if err != nil {
		fmt.Print(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	pa := r.URL.Path[1:]
	fmt.Println(pa)
	r.ParseForm()
	destPath := r.Form.Get("destPath")

	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		var ff string
		if dir != "." {
			if !isDirExists(dir+destPath) {
				os.MkdirAll(dir+"/"+destPath, os.ModePerm)
			}
			ff = dir + destPath + "/" + pa + part.FileName()
		} else {
			if !isDirExists(pa+destPath) {
				os.MkdirAll(pa+"/"+destPath, os.ModePerm)
			}
			ff = pa + destPath+ "/" + part.FileName()
		}
		fmt.Println(part)
		dst, err := os.Create(ff)
		defer dst.Close()

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if _, err := io.Copy(dst, part); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if strings.EqualFold("true",r.Form.Get("unzip")) {

			if !isDirExists(dir+destPath) {
				os.MkdirAll(dir+"/"+destPath, os.ModePerm)
			}
			if err := zipDeCompressCurrentPath(ff, dir+destPath + "/"); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

		}
	}

	fmt.Fprint(w, "ok")
	return
}
