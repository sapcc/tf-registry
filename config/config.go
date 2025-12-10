// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company
// SPDX-License-Identifier: Apache-2.0

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
