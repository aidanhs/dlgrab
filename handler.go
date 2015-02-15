package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
)

type Mapping struct {
	Method  string
	Regexp  *regexp.Regexp
	Handler func(http.ResponseWriter, *http.Request, [][]string)
}

type Handler struct {
	DataDir  string
	Mappings []*Mapping
}

func (h *Handler) WriteJsonHeader(w http.ResponseWriter) {
	w.Header().Add("Content-Type", "application/json")
}

func (h *Handler) WriteEndpointsHeader(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("X-Docker-Endpoints", r.Host)
}

func (h *Handler) GetPing(w http.ResponseWriter, r *http.Request, p [][]string) {
	w.Header().Add("X-Docker-Registry-Version", "0.0.1")
	w.WriteHeader(200)
	fmt.Fprint(w, "pong")
}

func (h *Handler) GetImageJson(w http.ResponseWriter, r *http.Request, p [][]string) {
	idPrefix := p[0][2]
	if paths, err := filepath.Glob(h.DataDir + "/images/" + idPrefix + "*"); err == nil {
		if len(paths) > 0 {
			image := &Image{paths[0]}
			file, err := os.Open(image.Dir + "/json")
			if err == nil {
				if file, err := os.Open(image.LayerPath()); err == nil {
					if stat, err := file.Stat(); err == nil {
						w.Header().Add("X-Docker-Size", fmt.Sprintf("%d", stat.Size()))
					}
				}
				w.WriteHeader(http.StatusOK)
				io.Copy(w, file)
				return
			}
		}
	}
	w.WriteHeader(http.StatusNotFound)
}

func (h *Handler) PutImageResource(w http.ResponseWriter, r *http.Request, p [][]string) {
	imageId := p[0][2]
	tagName := p[0][3]

	err := writeFile(h.DataDir+"/images/"+imageId+"/"+tagName, r.Body)
	if err != nil {
		logger.Error(err.Error())
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

func (h *Handler) PutRepository(w http.ResponseWriter, r *http.Request, p [][]string) {
	h.WriteJsonHeader(w)
	h.WriteEndpointsHeader(w, r)
	w.Header().Add("WWW-Authenticate", `Token signature=123abc,repository="dynport/test",access=write`)
	w.Header().Add("X-Docker-Token", "token")
	w.WriteHeader(http.StatusOK)
	repoName := p[0][2]
	repo := &Repository{h.DataDir + "/repositories/" + repoName}
	err := writeFile(repo.IndexPath(), r.Body)
	if err != nil {
		logger.Error(err.Error())
	}
}

func (h *Handler) Map(t, re string, f func(http.ResponseWriter, *http.Request, [][]string)) {
	if h.Mappings == nil {
		h.Mappings = make([]*Mapping, 0)
	}
	h.Mappings = append(h.Mappings, &Mapping{t, regexp.MustCompile("/v(\\d+)/" + re), f})
}

func (h *Handler) doHandle(w http.ResponseWriter, r *http.Request) (ok bool) {
	for _, mapping := range h.Mappings {
		if r.Method != mapping.Method {
			continue
		}
		if res := mapping.Regexp.FindAllStringSubmatch(r.URL.String(), -1); len(res) > 0 {
			mapping.Handler(w, r, res)
			return true
		}
	}
	return false
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logger.Info(fmt.Sprintf("Got request %s %s", r.Method, r.URL.String()))
	if ok := h.doHandle(w, r); !ok {
		logger.Info("returning 404")
		http.NotFound(w, r)
	}
}

func DummyResponse(status int) (func(http.ResponseWriter, *http.Request, [][]string)) {
	return func (w http.ResponseWriter, r *http.Request, p [][]string) {
		logger.Info(fmt.Sprintf("ignoring request, returning %d", status))
		w.WriteHeader(status)
	}
}

func NewHandler(dataDir string) (handler *Handler) {
	handler = &Handler{DataDir: dataDir}

	// this isn't a full registry
	handler.Map("GET", "users", DummyResponse(http.StatusNotImplemented))
	handler.Map("GET", "images/(.*?)/ancestry", DummyResponse(http.StatusNotImplemented))
	handler.Map("GET", "images/(.*?)/layer", DummyResponse(http.StatusNotImplemented))
	handler.Map("GET", "repositories/(.*?)/tags", DummyResponse(http.StatusNotImplemented))
	handler.Map("GET", "repositories/(.*?)/images", DummyResponse(http.StatusNotImplemented))
	handler.Map("PUT", "repositories/(.*?)/tags/(.*)", DummyResponse(http.StatusOK))
	handler.Map("PUT", "repositories/(.*?)/images", DummyResponse(http.StatusNoContent))

	// dummies
	handler.Map("GET", "_ping", handler.GetPing)

	// images
	handler.Map("GET", "images/(.*?)/json", handler.GetImageJson)
	handler.Map("PUT", "images/(.*?)/(.*)", handler.PutImageResource)

	// repositories
	handler.Map("PUT", "repositories/(.*?)/$", handler.PutRepository)
	return
}
