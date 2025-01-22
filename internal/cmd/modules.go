/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"path/filepath"

	"github.com/sapcc/tf-registry/config"

	"github.com/go-chi/chi/v5"
)

func httpGetModuleVersions(w http.ResponseWriter, r *http.Request) {
	fmt.Println(" > inside httpGetModuleVersions")
	m := config.Module{
		Namespace: chi.URLParam(r, "namespace"),
		Name:      chi.URLParam(r, "name"),
		Provider:  chi.URLParam(r, "provider"),
	}
	modPath := filepath.Join(C.Mprefix, m.Namespace, m.Name, m.Provider)
	fmt.Println(" > modpath:", modPath)
	modVers, err := getModuleVersions(modPath)
	if err != nil {
		// TODO handle module not found with 404
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(modVers)
	if err != nil {
		fmt.Println(err)
	}
}

func getModuleVersions(modPath string) (config.ModuleVersionsResp, error) {
	fmt.Println(" > inside getModuleVersions")
	fmt.Println(" >> modpath:", modPath)

	m := config.ModuleVersions{}
	versionDirs, err := fs.ReadDir(C.S3fsys, modPath)
	fmt.Println(" >> versionDirs:", versionDirs)
	if err != nil {
		return config.ModuleVersionsResp{}, err
	}
	for _, v := range versionDirs {
		vers := map[string]string{"version": v.Name()}
		m.Versions = append(m.Versions, vers)
		fmt.Println(" >> found version: ", vers)
	}
	return config.ModuleVersionsResp{
		Modules: []config.ModuleVersions{m},
	}, nil
}

func httpGetModuleDownloadURL(w http.ResponseWriter, r *http.Request) {
	fmt.Println(" > inside httpGetModuleDownloadURL")
	m := config.Module{
		Namespace: chi.URLParam(r, "namespace"),
		Name:      chi.URLParam(r, "name"),
		Provider:  chi.URLParam(r, "provider"),
		Version:   chi.URLParam(r, "version"),
	}
	tfGetHeader := filepath.Join(
		"/download/modules",
		m.Namespace,
		m.Name,
		m.Provider,
		m.Version,
		m.Name+".tgz",
	)
	fmt.Println(" > tfGetHeader:", tfGetHeader)
	w.Header().Set("X-Terraform-Get", tfGetHeader)
	w.WriteHeader(http.StatusNoContent)
}

func httpGetModule(w http.ResponseWriter, r *http.Request) {
	fmt.Println(" > inside httpGetModule")
	w.Header().Set("Content-Encoding", "application/octet-stream")
	w.Header().Set("Content-Type", "application/x-gzip")
	fs := http.StripPrefix("/download/", http.FileServer(http.FS(C.S3fsys)))
	fs.ServeHTTP(w, r)
}
