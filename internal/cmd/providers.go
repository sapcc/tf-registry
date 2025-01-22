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
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/sapcc/tf-registry/config"
)

func getProviderVersions(provPath string) (config.ProviderVersions, error) {
	fmt.Println(" > inside getProviderVersions")
	fmt.Println(" >> provPath:", provPath)
	p := config.ProviderVersions{}
	versionDirs, err := fs.ReadDir(C.S3fsys, provPath)
	fmt.Println(" >> versionDirs:", versionDirs)
	if err != nil {
		return config.ProviderVersions{}, err
	}
	for _, v := range versionDirs {
		vers := map[string]string{"version": v.Name()}
		p.Versions = append(p.Versions, vers)
		fmt.Println(" >> found version: ", vers)
	}
	return p, nil
}

// httpGetVersions is a http handler for retrieving a list of module versions
// the registry server expects the versions to all be a set of
func httpGetProviderVersions(w http.ResponseWriter, r *http.Request) {
	fmt.Println(" > inside httpGetProviderVersions")
	p := config.Provider{
		Namespace: chi.URLParam(r, "namespace"),
		Provider:  chi.URLParam(r, "provider"),
	}

	provPath := filepath.Join(C.Pprefix, p.Namespace, p.Provider)
	fmt.Println(" >> provPath: ", provPath)

	provVers, err := getProviderVersions(provPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(provVers)
	if err != nil {
		fmt.Println(err)
	}
}

// function to get the shasum for a specific provider
func getSHASUM(downloadPath string) string {
	fmt.Println(" > inside getSHASUM")
	s3file := strings.Trim(downloadPath, "downla/")
	fmt.Println(" >> file to be open: ", s3file)
	f, err := C.S3fsys.Open(s3file)
	if err != nil {
		fmt.Println(err)
	}
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		log.Fatal(err)
	}
	sha256sum := hex.EncodeToString(h.Sum(nil))
	fmt.Println(" >> calculated sum: ", sha256sum)

	return sha256sum
}

// function to read public GPG key
func getGPGkey(keyFile string) string {
	fmt.Println(" > inside getGPGL", keyFile)
	content, err := os.ReadFile(keyFile)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(" >> content (first 10 chars): ", content[0:10])
	return string(content)
}

// httpGetProviderDownLoadURL is a http handler for retrieving the final download URL for a terraform module,
func httpGetProviderDownloadURL(w http.ResponseWriter, r *http.Request) {
	fmt.Println(" > inside httpGetProviderDownloadURL")
	fmt.Println(" >> building json response")

	p := config.Provider{
		Namespace: chi.URLParam(r, "namespace"),
		Provider:  chi.URLParam(r, "provider"),
		Version:   chi.URLParam(r, "version"),
		Os:        chi.URLParam(r, "os"),
		Arch:      chi.URLParam(r, "arch"),
	}
	downloadPrefix := filepath.Join("/download/", C.Pprefix, "/", p.Namespace, p.Provider, p.Version)
	signaturePrefix := filepath.Join("/", C.Pprefix, "/", p.Namespace, p.Provider, p.Version, "/signatures")
	filename := "terraform-provider-" + p.Provider + "_" + p.Version + "_" + p.Os + "_" + p.Arch + ".zip"
	fmt.Printf(" >> see complete json here: https://<server>/%s/%s/%s/download/%s/%s", p.Namespace, p.Provider, p.Version, p.Os, p.Arch)

	// building response json: https://developer.hashicorp.com/terraform/internals/provider-registry-protocol#protocols-1

	providerResponse := config.ProviderResp{}
	providerResponse.DownloadURL = downloadPrefix + "/" + filename
	fmt.Println(" >> downloadURL: ", providerResponse.DownloadURL)
	providerResponse.Filename = filename
	providerResponse.Shasum = getSHASUM(providerResponse.DownloadURL) // calculated ar GET time
	providerResponse.ShasumsURL = signaturePrefix + "/SHA256SUMS"
	providerResponse.Os = p.Os
	providerResponse.Arch = p.Arch
	providerResponse.ShasumsSignatureURL = signaturePrefix + "/SHA256SUMS.sig"
	providerResponse.SigningKeys = map[string]interface{}{"gpg_public_keys": []map[string]interface{}{{"key_id": "terraform-registry", "ascii_armor": getGPGkey(C.GpgKey)}}}

	// fields that the documentation list as required but does not seem to be needed for now
	// providerResponse.Protocols = []string{"4.0", "5.1"}
	// providerResponse.Trust_signature = ""
	// providerResponse.Source = "some org name"
	// providerResponse.Source_url = "/"

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(providerResponse)
	if err != nil {
		fmt.Println(err)
	}
}

// http handler for retrieving a terraform provider

func httpGetProvider(w http.ResponseWriter, r *http.Request) {
	fmt.Println(" > inside httpGetProvider")
	w.Header().Set("Content-Encoding", "application/octet-stream")
	w.Header().Set("Content-Type", "application/x-gzip")
	fs := http.StripPrefix("/download/", http.FileServer(http.FS(C.S3fsys)))
	fs.ServeHTTP(w, r)
}

func httpGetSignatures(w http.ResponseWriter, r *http.Request) {
	fmt.Println(" > inside httpGetSignatures")
	w.Header().Set("Content-Type", "text")
	fs := http.FileServer(http.FS(C.S3fsys))
	fs.ServeHTTP(w, r)
}
