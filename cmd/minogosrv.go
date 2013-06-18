package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/lukegb/minotar"
	"image"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
)

const (
	DEFAULT_SIZE = uint(180)
	MAX_SIZE     = uint(300)
	MIN_SIZE     = uint(8)

	STATIC_LOCATION = "static"

	LISTEN_ON = ":9999"
)

func serveStatic(w http.ResponseWriter, r *http.Request, inpath string) error {
	inpath = path.Clean(inpath)
	r.URL.Path = inpath

	if !strings.HasPrefix(inpath, "/") {
		inpath = "/" + inpath
		r.URL.Path = inpath
	}
	path := STATIC_LOCATION + inpath

	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	d, err := f.Stat()
	if err != nil {
		return err
	}

	http.ServeContent(w, r, d.Name(), d.ModTime(), f)
	return nil
}

func serveAssetPage(w http.ResponseWriter, r *http.Request) {
	err := serveStatic(w, r, r.URL.Path)
	if err != nil {
		notFoundPage(w, r)
	}
}

func indexPage(w http.ResponseWriter, r *http.Request) {
	err := serveStatic(w, r, "index.html")
	if err != nil {
		notFoundPage(w, r)
	}
}

type NotFoundHandler struct{}

func (h NotFoundHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(404)

	f, err := os.Open("static/404.html")
	if err != nil {
		fmt.Fprintf(w, "404 file not found")
		return
	}
	defer f.Close()

	io.Copy(w, f)
}

func notFoundPage(w http.ResponseWriter, r *http.Request) {
	nfh := NotFoundHandler{}
	nfh.ServeHTTP(w, r)
}
func serverErrorPage(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(500)
	fmt.Fprintf(w, "500 internal server error")
}

func rationalizeSize(inp string) uint {
	out64, err := strconv.ParseUint(inp, 10, 0)
	out := uint(out64)
	if err != nil {
		return DEFAULT_SIZE
	} else if out > MAX_SIZE {
		return MAX_SIZE
	} else if out < MIN_SIZE {
		return MIN_SIZE
	}
	return out
}

func fetchImageProcessThen(callback func(minotar.Skin) (image.Image, error)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)

		username := vars["username"]
		size := rationalizeSize(vars["size"])

		skin, err := minotar.FetchSkinForUser(username)
		if err != nil {
			if skin, err = minotar.FetchSkinForSteve(); err != nil {
				serverErrorPage(w, r)
				return
			}
		}

		img, err := callback(skin)
		if err != nil {
			serverErrorPage(w, r)
			return
		}

		imgResized := minotar.Resize(size, size, img)

		w.Header().Add("Content-Type", "image/png")
		minotar.WritePNG(w, imgResized)
	}
}
func skinPage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	username := vars["username"]

	userSkinURL := minotar.URLForUser(username)
	resp, err := http.Get(userSkinURL)
	if err != nil {
		notFoundPage(w, r)
		return
	}
	w.Header().Add("Content-Type", "image/png")
	defer resp.Body.Close()
	io.Copy(w, resp.Body)
}
func downloadPage(w http.ResponseWriter, r *http.Request) {
	headers := w.Header()
	headers.Add("Content-Disposition", "attachment; filename=\"skin.png\"")
	skinPage(w, r)
}

func main() {
	avatarPage := fetchImageProcessThen(func(skin minotar.Skin) (image.Image, error) {
		return skin.Head()
	})
	helmPage := fetchImageProcessThen(func(skin minotar.Skin) (image.Image, error) {
		return skin.Helm()
	})

	r := mux.NewRouter()
	r.NotFoundHandler = NotFoundHandler{}

	r.HandleFunc("/avatar/{username:[a-zA-Z0-9]+}{extension:(.png)?}", avatarPage)
	r.HandleFunc("/avatar/{username:[a-zA-Z0-9]+}/{size:[0-9]+}{extension:(.png)?}", avatarPage)

	r.HandleFunc("/helm/{username:[a-zA-Z0-9]+}{extension:(.png)?}", helmPage)
	r.HandleFunc("/helm/{username:[a-zA-Z0-9]+}/{size:[0-9]+}{extension:(.png)?}", helmPage)

	r.HandleFunc("/download/{username:[a-zA-Z0-9]+}{extension:(.png)?}", downloadPage)

	r.HandleFunc("/skin/{username:[a-zA-Z0-9]+}{extension:(.png)?}", skinPage)

	r.HandleFunc("/", indexPage)

	http.Handle("/", r)
	http.HandleFunc("/assets/", serveAssetPage)
	log.Fatalln(http.ListenAndServe(LISTEN_ON, nil))
}
