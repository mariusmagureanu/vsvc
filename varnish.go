package main

import (
	"fmt"
	"os"
	"os/exec"
	"sync"
	"text/template"
)

var tmpl *template.Template

const (
	vclTemplate = `
vcl 4.1;

import directors;
import std;

 {{- range . }}

backend {{ .Name }} {
    .host = "{{ .Name }}.{{ .Namespace }}.svc.cluster.local";
    .port = "{{ .Port }}";
}

 {{- end}}

sub vcl_recv {
    {{- range . }}
    if (req.http.host == "{{ .Name}}") {
        set req.backend_hint = {{ .Name }};
    }
    {{- end}}
}
`
	emptyVclTemplate = "vcl 4.1;\n backend default none;"

	varnishReloadCMD = "varnishreload"
)

type (
	// Backend maps to a k8s service which eventually
	// will be used as a backend for varnish.
	Backend struct {
		Name      string
		Namespace string
		Port      string
	}

	// BackendStore is a collection of backends, essentially
	// making up for a slice of k8s services wearing
	// the correct annotations.
	BackendStore struct {
		l              sync.RWMutex
		store          map[string]Backend
		varnishWorkDir string
		vclFilePath    string
	}
)

func initBackendStore(workDir, vclFile string) {
	backendStore = &BackendStore{}

	backendStore.store = make(map[string]Backend)
	backendStore.varnishWorkDir = workDir
	backendStore.vclFilePath = vclFile

	tmpl = template.Must(template.New("vcl").Parse(vclTemplate))
}

func (bs *BackendStore) add(b Backend) {
	bs.l.Lock()
	defer bs.l.Unlock()

	key := fmt.Sprintf("%s/%s", b.Namespace, b.Name)
	bs.store[key] = b
}

func (bs *BackendStore) delete(b Backend) {
	bs.l.Lock()
	defer bs.l.Unlock()

	key := fmt.Sprintf("%s/%s", b.Namespace, b.Name)
	delete(bs.store, key)
}

func (bs *BackendStore) updateVCL() error {
	var b []Backend

	bs.l.Lock()
	for _, v := range bs.store {
		b = append(b, v)
	}
	bs.l.Unlock()

	f, err := os.Create(bs.vclFilePath)
	if err != nil {
		return err
	}
	defer f.Close()

	if len(b) > 0 {
		err = tmpl.Execute(f, b)
		if err != nil {
			return err
		}
	} else {
		_, err = f.WriteString(emptyVclTemplate)
		if err != nil {
			return err
		}
	}

	Debug("updating vcl:", bs.vclFilePath)
	cmd := exec.Command(varnishReloadCMD, "-n", bs.varnishWorkDir)

	return cmd.Run()
}
