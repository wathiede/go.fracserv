package main

import (
	"bytes"
	"cache"
	"flag"
	"fmt"
	"fractal"
	"fractal/debug"
	"fractal/julia"
	"fractal/mandelbrot"
	"fractal/solid"
	"html/template"
	"image/png"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

var factory map[string]func(o fractal.Options) (fractal.Fractal, error)
var port string
var cacheDir string
var disableCache bool
var pngCache cache.Cache

type cachedPng struct {
	Timestamp time.Time
	Bytes []byte
}

func (c cachedPng) Size() int {
	return len(c.Bytes)
}

func init() {
	flag.StringVar(&port, "port", "8000", "webserver listen port")
	flag.StringVar(&cacheDir, "cacheDir", "/tmp/fractals",
		"directory to store rendered tiles. Directory must exist")
	flag.BoolVar(&disableCache, "disableCache", false,
		"never serve from disk cache")
	flag.Parse()

	factory = map[string]func(o fractal.Options) (fractal.Fractal, error){
		"debug": debug.NewFractal,
		"solid": solid.NewFractal,
		"mandelbrot": mandelbrot.NewFractal,
		"julia": julia.NewFractal,
		//"glynn": glynn.NewFractal,
		//"lyapunov": lyapunov.NewFractal,
	}

	// TODO(wathiede): load tiles on startup
	pngCache = *cache.NewCache()
}

func main() {
	fmt.Printf("Listening on:\n")
	host, err := os.Hostname()
	if err != nil {
		log.Fatal("Failed to get hostname from os:", err)
	}
	fmt.Printf("  http://%s:%s/\n", host, port)

	s := "static/"
	_, err = os.Open(s)
	if os.IsNotExist(err) {
		log.Fatalf("Directory %s not found, please run for directory containing %s\n", s, s)
	}

	http.Handle("/"+s, http.StripPrefix("/"+s, http.FileServer(http.Dir(s))))
	http.HandleFunc("/", IndexServer)
	log.Fatal(http.ListenAndServe(":" + port, nil))
}

func drawFractalPage(w http.ResponseWriter, req *http.Request, fracType string) {
	t, err := template.ParseFiles(fmt.Sprintf("templates/%s.html", fracType))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = t.Execute(w, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func fsNameFromURL(url string) string {
	cleanup := func(r rune) rune {
		switch r {
		case '?':
			return '/'
		case '&':
			return ','
		}
		return r
	}
	return strings.Map(cleanup, url)
}

func savePngFromCache(cacheKey string) {
	cacher, ok := pngCache.Get(cacheKey)
	if !ok {
		log.Printf("Attempt to save %q to disk, but image not in cache",
			cacheKey)
		return
	}

	cachefn := cacheDir + cacheKey
	d := path.Dir(cachefn)
	if _, err := os.Stat(d); err != nil {
		log.Printf("Creating cache dir for %q", d)
		err = os.Mkdir(d, 0700)
	}

	_, err := os.Stat(cachefn)
	if err == nil {
		log.Printf("Attempt to save %q to %q, but file already exists",
			cacheKey, cachefn)
		return
	}

	outf, err := os.OpenFile(cachefn, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Failed to open tile %q for save: %s", cachefn, err)
		return
	}
	cp := cacher.(cachedPng)
	outf.Write(cp.Bytes)
	outf.Close()

	err = os.Chtimes(cachefn, cp.Timestamp, cp.Timestamp)
	if err != nil {
		log.Printf("Error setting atime and mtime on %q: %s", cachefn, err)
	}
}

func drawFractal(w http.ResponseWriter, req *http.Request, fracType string) {
	if disableCache {
		i, err := factory[fracType](fractal.Options{req.URL.Query()})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		png.Encode(w, i)
		return
	}

	cacheKey := fsNameFromURL(req.URL.RequestURI())
	cacher, ok := pngCache.Get(cacheKey)
	// TODO(wathiede): log cache hits as expvar
	if !ok {
		// No png in cache, create one
		i, err := factory[fracType](fractal.Options{req.URL.Query()})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		b := &bytes.Buffer{}
		png.Encode(b, i)
		cacher = cachedPng{time.Now(), b.Bytes()}
		pngCache.Add(cacheKey, cacher)

		// Async save image to disk
		// TODO make this a channel and serialize saving of images
		go savePngFromCache(cacheKey)
	}

	cp := cacher.(cachedPng)


	// Using this instead of io.Copy, sets Last-Modified which helps given
	// the way the maps API makes lots of re-requests
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Last-Modified", cp.Timestamp.Format(http.TimeFormat))
	w.Header().Set("Expires",
		cp.Timestamp.Add(time.Hour).Format(http.TimeFormat))
	w.Write(cp.Bytes)
}

func IndexServer(w http.ResponseWriter, req *http.Request) {
	// TODO(wathiede): check fracType against keys in factory before handing
	// off to handler to prevent directory traversal exploit from malformed
	// urls.
	fracType := req.URL.Path[1:]
	if fracType != "" {
		//log.Println("Found fractal type", fracType)

		if len(req.URL.Query()) != 0 {
			drawFractal(w, req, fracType)
		} else {
			drawFractalPage(w, req, fracType)
		}
		return
	}

	t, err := template.ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = t.Execute(w, factory)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
