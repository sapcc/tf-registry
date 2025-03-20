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
	"os"
	"time"

	"github.com/sapcc/tf-registry/config"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jszwec/s3fs"
	"github.com/sapcc/go-api-declarations/bininfo"
	"github.com/spf13/cobra"
)

// Globals

var C = &config.CmdLineParams{}

var RootCmd = &cobra.Command{
	Use:     "tf-registry",
	Short:   "tf-registry is a private terraform registry witha s3 backend",
	RunE:    RunRootCmd,
	Version: bininfo.Version(),
}

func init() {
	RootCmd.PersistentFlags().StringVar(&C.ServerCert, "serverCert", "", "path to https server certificate")
	RootCmd.PersistentFlags().StringVar(&C.ServerKey, "serverKey", "", "path to https server key")
	RootCmd.PersistentFlags().StringVar(&C.GpgKeys, "gpgKeys", "", "path to gpg public keys folder")
	RootCmd.PersistentFlags().StringVar(&C.Pprefix, "pprefix", "", "optional path prefix for local providers")
	RootCmd.PersistentFlags().StringVar(&C.Mprefix, "mprefix", "", "optional path prefix for local modules")
	RootCmd.PersistentFlags().StringVar(&C.Port, "port", "8001", "port for HTTP server")
	RootCmd.PersistentFlags().StringVar(&C.Bucket, "bucket", "", "aws s3 bucket name containing terraform providers")
	err := RootCmd.MarkPersistentFlagRequired("serverCert")
	if err != nil {
		fmt.Println(err)
	}
	err = RootCmd.MarkPersistentFlagRequired("serverKey")
	if err != nil {
		fmt.Println(err)
	}
	err = RootCmd.MarkPersistentFlagRequired("gpgKeys")
	if err != nil {
		fmt.Println(err)
	}
	err = RootCmd.MarkPersistentFlagRequired("pprefix")
	if err != nil {
		fmt.Println(err)
	}
	err = RootCmd.MarkPersistentFlagRequired("mprefix")
	if err != nil {
		fmt.Println(err)
	}
	err = RootCmd.MarkPersistentFlagRequired("bucket")
	if err != nil {
		fmt.Println(err)
	}
}

// TF Registry Server
func RunRootCmd(cmd *cobra.Command, args []string) error {
	fmt.Printf(" > Starting tf-registry webserver on 0.0.0.0:%s...\n", C.Port)
	fmt.Println(" > Connecting to storage backend...")

	fmt.Println(" > Using env vars to connect to aws")

	sess := session.Must(session.NewSession())

	C.S3fsys = s3fs.New(s3.New(sess), C.Bucket)
	bucketRoot := "."
	fmt.Println(" > bucketroot:", bucketRoot)
	stat, errr := fs.Stat(C.S3fsys, bucketRoot)
	if errr != nil {
		fmt.Println(errr)
		os.Exit(1)
	}

	fmt.Println(" > fsstat:", stat)

	fmt.Printf(" > Connection successful, serving terraform registry from: s3://%s/\n", C.Bucket)

	// Configure a go-chi router
	r := chi.NewRouter()
	r.Use(middleware.RealIP)
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)
	r.Use(middleware.GetHead)

	// is http server alive?
	r.Get("/alive", httpIAmAlive)

	// is s3 connection working?
	r.Get("/s3_reachable", httpS3Status)
	// ROUTES

	// GET / returns our static service discovery resp
	r.Get("/", httpGetServiceDiscovery)

	// GET /.well-known/terraform.json returns our static service discovery resp
	r.Get("/.well-known/terraform.json", httpGetServiceDiscovery)

	// PROVIDERS

	// GET /:namespace/:name/:provider/versions returns a list of versions for the specified module path
	r.Get("/{namespace}/{provider}/versions", httpGetProviderVersions)

	// GET /:namespace/:name/:provider/:version/download responds with a 204 and X-Terraform-Get header pointing to the download path
	r.Get("/{namespace}/{provider}/{version}/download/{os}/{arch}", httpGetProviderDownloadURL)

	// GET /download/ provides an http fileserver for downloading modules as gzipped tarballs
	r.Get("/download/{namespace}/{provider}/{system}/{version}/*", httpGetModule)
	r.Get("/download/{namespace}/{provider}/{version}/*", httpGetProvider)

	r.Get(config.ProviderBasePath+"/{namespace}/{provider}/{version}/signatures/*", httpGetSignatures)

	// MODULES
	r.Get(config.ModuleBasePath+"/{namespace}/{name}/{provider}/versions", httpGetModuleVersions)

	r.Get(config.ModuleBasePath+"/{namespace}/{name}/{provider}/{version}/download", httpGetModuleDownloadURL)

	// FILES

	// r.Get("/files/{file}", httpGetFile)
	r.Handle("/files/*", http.FileServer(http.FS(C.S3fsys)))

	err := chi.Walk(r, func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		fmt.Printf("[%s]: '%s' has %d middlewares\n", method, route, len(middlewares))
		return nil
	})
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(" > starting TLS server")

	server := &http.Server{
		Addr:              ":" + C.Port,
		Handler:           r,
		ReadHeaderTimeout: 3 * time.Second,
	}

	err = server.ListenAndServeTLS(C.ServerCert, C.ServerKey)
	// err = http.ListenAndServeTLS(":"+C.Port, C.ServerCert, C.ServerKey, r)
	if err != nil {
		fmt.Println(err)
	}
	return nil
}

// httpGetServiceDiscovery is a http handler for returning the
// base path for the modules API provided by this registry
func httpGetServiceDiscovery(w http.ResponseWriter, r *http.Request) {
	fmt.Println(" > inside httpGetServiceDiscovery")
	s := config.ServiceDiscoveryResp{ProvidersV1: config.ProviderBasePath, ModulesV1: config.ModuleBasePath}
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(s)
	if err != nil {
		fmt.Println(err)
	}
}

func httpIAmAlive(w http.ResponseWriter, r *http.Request) {
	fmt.Println(" > inside httpIAmAlive")
	w.Header().Set("Content-Type", "text")
	_, err := w.Write([]byte("internal terraform provier registry is alive"))
	if err != nil {
		fmt.Println(err)
	}
}

func httpS3Status(w http.ResponseWriter, r *http.Request) {
	fmt.Println(" > inside httpS3Status")
	statusFile := "status/donotremove.txt"
	content, err := fs.ReadFile(C.S3fsys, statusFile)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	w.Header().Set("Content-Type", "text")
	_, err = w.Write(content)
	if err != nil {
		fmt.Println(err)
	}
}
