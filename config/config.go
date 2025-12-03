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

package config

import "io/fs"

const ProviderBasePath = "/providers"
const ModuleBasePath = "/modules"

type ServiceDiscoveryResp struct {
	ProvidersV1 string `json:"providers.v1"`
	ModulesV1   string `json:"modules.v1"`
}

type ProviderVersions struct {
	Versions []map[string]string `json:"versions"`
}

type Provider struct {
	Namespace string
	Provider  string
	Version   string
	Os        string
	Arch      string
}

type ProviderResp struct {
	DownloadURL         string         `json:"download_url"`
	Shasum              string         `json:"shasum"`
	Os                  string         `json:"os"`
	Arch                string         `json:"arch"`
	Filename            string         `json:"filename"`
	ShasumsURL          string         `json:"shasums_url"`
	ShasumsSignatureURL string         `json:"shasums_signature_url"`
	SigningKeys         map[string]any `json:"signing_keys"`
}

type ModuleVersions struct {
	Versions []map[string]string `json:"versions"`
}

// ModuleVersionsResp is our module versions response struct
type ModuleVersionsResp struct {
	Modules []ModuleVersions `json:"modules"`
}

// Module respresents a terraform module
type Module struct {
	Namespace string
	Name      string
	Provider  string
	Version   string
}

// Commandline parameters
type CmdLineParams struct {
	ServerCert string // https certificate
	ServerKey  string // https key
	GpgKeys    string // gpg public keys folder
	Pprefix    string
	Mprefix    string
	Port       string

	Bucket string
	S3fsys fs.FS
}
