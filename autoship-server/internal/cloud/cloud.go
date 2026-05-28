// Package cloud abstracts the cloud-specific pieces of Auto-Ship (static-site
// storage and per-container firewall rules) behind a single Provider interface,
// so the rest of the server is agnostic to whether we run on AWS or Azure.
//
// The active provider is chosen at startup from the CLOUD_PROVIDER env var
// (see Init) and accessed everywhere else via Get().
package cloud

import (
	"fmt"
	"path/filepath"
	"strings"
)

// StorageProvider hosts a static site by uploading a built folder and returning
// the public URL to its index.html. Backed by S3 (AWS) or Blob Storage (Azure).
type StorageProvider interface {
	UploadStaticSite(localPath, keyPrefix string) (string, error)
}

// FirewallProvider opens an inbound TCP port so a dynamic container is reachable
// from the internet. Backed by an EC2 security group (AWS) or an NSG (Azure).
type FirewallProvider interface {
	AuthorizePort(port int) error
}

// Provider bundles every cloud capability the server needs for one backend.
type Provider interface {
	StorageProvider
	FirewallProvider
	// Name returns the provider identifier, e.g. "aws" or "azure".
	Name() string
}

// active holds the provider selected by Init.
var active Provider

// Init selects and initializes the cloud provider from name (typically the value
// of CLOUD_PROVIDER). An empty name defaults to "aws" to preserve the original
// behavior. It must be called once at startup before Get.
func Init(name string) error {
	name = strings.ToLower(strings.TrimSpace(name))
	if name == "" {
		name = "aws"
	}

	var (
		p   Provider
		err error
	)
	switch name {
	case "aws":
		p, err = newAWSProvider()
	case "azure":
		p, err = newAzureProvider()
	default:
		return fmt.Errorf("unknown CLOUD_PROVIDER %q (expected \"aws\" or \"azure\")", name)
	}
	if err != nil {
		return fmt.Errorf("failed to initialize %q cloud provider: %w", name, err)
	}

	active = p
	return nil
}

// Get returns the active provider. It panics if Init was never called, which
// would be a programming error rather than a recoverable runtime condition.
func Get() Provider {
	if active == nil {
		panic("cloud.Get() called before cloud.Init()")
	}
	return active
}

// contentTypeByExtension returns a basic content type from a file extension.
// Shared by the AWS and Azure storage uploaders so served assets get correct
// Content-Type headers instead of defaulting to a download.
func contentTypeByExtension(filePath string) string {
	switch strings.ToLower(filepath.Ext(filePath)) {
	case ".html", ".htm":
		return "text/html"
	case ".css":
		return "text/css"
	case ".js", ".mjs":
		return "application/javascript"
	case ".json":
		return "application/json"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".svg":
		return "image/svg+xml"
	case ".ico":
		return "image/x-icon"
	case ".webp":
		return "image/webp"
	case ".woff":
		return "font/woff"
	case ".woff2":
		return "font/woff2"
	default:
		return "application/octet-stream"
	}
}
