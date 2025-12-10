<!--
SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company

SPDX-License-Identifier: Apache-2.0
-->

# tf-registry
Self Hosted Terraform Registry backed by S3
Based on the initial work of [terraform-registry](https://github.com/nrkno/terraform-registry)

## Usage


`tf-registry` Provides a simple http server that implements the [Terraform Provider Registry Protocol](https://developer.hashicorp.com/terraform/internals/provider-registry-protocol).

```
Terraform Registry Server

Usage: tf-registry [flags] 

Flags:
  -- bucket string
    	aws s3 bucket name containing terraform providers
  -- port string
    	port for HTTPS server (default "443")
  -- profile string
    	aws named profile to assume (default "default")
  -- serverCert
      path to https server certificate
  -- serverKey
      path to https server key
  -- gpgKey
      path to gpg public key
```

### Uploading Providers

Pre requisites:
A GPG key pair was created and stored in vault. 


Process is explained below:
1. build new provider from source. name must be terraform-provider-<name>
2. zip it to terraform-provide-<name>_<version>_<os>_<arch>.zip
3. get the sha256 sum of the file and put it in the SHA256SUM file: shasum -a 256 >> SHA256SUM
4. get the detached signature file: gpg -b SHA256SUM 
5. upload the provider zip file to s3: bucket/org/name/version/provider.zip. e.g. terraform-registry-1/cp/daybreak/1.0.0/terraform-provider-daybreak_1.0.0_linux_amd64.zip
6. upload SHA256SUM and SHA256SUM.SIG to s3: bucket/org/name/version/, e.g. terraform-registry-1/cp/daybreak/1.0.0/signatures

### Starting server


E.g: start the server: 
```
./tf-registry -bucket terraform-registry-1 -port 443 -serverCert certs/localhost.crt -serverKey certs/localhost.key -gpgKey certs/gpg.pub
```
Certs and key will be mounted to /certs, gpg key via /gpg/gpg.pub

## TODO

- manage providers versioning



