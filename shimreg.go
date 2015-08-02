package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type Mapping struct {
	Method  string
	Regexp  *regexp.Regexp
	Handler func(http.ResponseWriter, *http.Request, [][]string)
}

type Handler struct {
	OutDir    string
	RegFormat bool
	Mappings  []*Mapping
}

func (h *Handler) GetPing(w http.ResponseWriter, r *http.Request, p [][]string) {
	w.Header().Add("X-Docker-Registry-Version", "0.0.1")
	w.WriteHeader(200)
	fmt.Fprint(w, "pong")
}

func (h *Handler) GetImageJson(w http.ResponseWriter, r *http.Request, p [][]string) {
	w.Header().Add("Content-Type", "application/json")
	idPrefix := p[0][2]
	layerLock.Lock()
	defer layerLock.Unlock()
	// pretend we have everything but the layer we're trying to save
	if strings.HasPrefix(layerId, idPrefix) {
		w.WriteHeader(http.StatusNotFound)
	} else {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{}"))
	}
}

func (h *Handler) PutImageResource(w http.ResponseWriter, r *http.Request, p [][]string) {
	imageId := p[0][2]
	resourceName := p[0][3]

	layerLock.Lock()
	defer layerLock.Unlock()
	if imageId != layerId {
		logger.Error("Client tried to push layer %s, rejecting", imageId)
		w.WriteHeader(http.StatusForbidden)
		return
	}

	if !h.RegFormat {
		logger.Debug("Considering %s for special handling")
		// Specifically skipping "checksum" here, but any new junk as well
		if resourceName != "layer" && resourceName != "json" {
			w.WriteHeader(http.StatusOK)
			return
		}
		if resourceName == "layer" {
			resourceName = "layer.tar"
		}
	}

	path := filepath.Join(h.OutDir, layerId, resourceName)
	logger.Info("Writing file: %s", filepath.Base(path))
	out, err := os.Create(path)
	if err == nil {
		defer out.Close()
		cnt, err := io.Copy(out, r.Body)
		if err == nil {
			logger.Debug("Wrote %d bytes", cnt)
		}
	}

	if err != nil {
		logger.Error("%s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

func (h *Handler) PutRepository(w http.ResponseWriter, r *http.Request, p [][]string) {
	w.Header().Add("X-Docker-Endpoints", r.Host)
	w.Header().Add("WWW-Authenticate", `Token signature=123abc,repository="dynport/test",access=write`)
	w.Header().Add("X-Docker-Token", "token")
	w.WriteHeader(http.StatusOK)
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
			r.Body.Close()
			return true
		}
	}
	r.Body.Close()
	return false
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logger.Debug("Got request %s %s", r.Method, r.URL.String())
	if ok := h.doHandle(w, r); !ok {
		logger.Debug("Returning 404")
		http.NotFound(w, r)
	}
}

func DummyResponse(status int) func(http.ResponseWriter, *http.Request, [][]string) {
	return func(w http.ResponseWriter, r *http.Request, p [][]string) {
		logger.Debug("Ignoring request, returning %d", status)
		w.WriteHeader(status)
	}
}

func NewHandler(outDir string, regFormat bool) (handler *Handler) {
	handler = &Handler{OutDir: outDir, RegFormat: regFormat}

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
