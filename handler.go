package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"net/http"
	"regexp"
	"path/filepath"
	"strings"
	"sync"
)

var layerLock sync.Mutex
var layerToSave string = ""

type Mapping struct {
	Method  string
	Regexp  *regexp.Regexp
	Handler func(http.ResponseWriter, *http.Request, [][]string)
}

type Handler struct {
	OutDir   string
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
	layerLock.Lock()
	defer layerLock.Unlock()
	// pretend we have everything but the layer we're trying to save
	if strings.HasPrefix(layerToSave, idPrefix) {
		w.WriteHeader(http.StatusNotFound)
	} else {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{}"))
	}
}

func (h *Handler) PutImageResource(w http.ResponseWriter, r *http.Request, p [][]string) {
	imageId := p[0][2]
	resourceName := p[0][3]
	path := filepath.Join(h.OutDir, imageId, resourceName)

	err := os.MkdirAll(filepath.Dir(path), 0755)
	if err == nil {
		logger.Info("Writing file: ", filepath.Base(path))

		out, err := os.Create(path)
		if err == nil {
			defer out.Close()
			cnt, err := io.Copy(out, r.Body)
			if err == nil {
				logger.Debug(fmt.Sprintf("Wrote %d bytes", cnt))
			}
		}
	}

	if err != nil {
		logger.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

// formerly created the _index file, holding a list of dicts of the image layers
func (h *Handler) PutRepository(w http.ResponseWriter, r *http.Request, p [][]string) {
	h.WriteJsonHeader(w)
	h.WriteEndpointsHeader(w, r)
	w.Header().Add("WWW-Authenticate", `Token signature=123abc,repository="dynport/test",access=write`)
	w.Header().Add("X-Docker-Token", "token")
	w.WriteHeader(http.StatusOK)

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logger.Error(err.Error())
		return
	}
	datas := []map[string]interface{}{}
	if err := json.Unmarshal(body, &datas); err != nil {
		logger.Error(err.Error())
		return
	}
	layerLock.Lock()
	defer layerLock.Unlock()
	for _, layer := range datas {
		layerToSave = layer["id"].(string)
	}
	logger.Info(fmt.Sprintf("Will save %s", layerToSave))
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

func NewHandler(outDir string) (handler *Handler) {
	handler = &Handler{OutDir: outDir}

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
