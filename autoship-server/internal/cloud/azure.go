package cloud

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v6"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
)

// azureProvider is the Azure counterpart of awsProvider: static sites go to Blob
// Storage (the static-website "$web" container by default) and dynamic container
// ports are opened with Network Security Group rules.
type azureProvider struct {
	// storage
	blobClient *azblob.Client
	container  string
	publicBase string // base URL for returned links, no trailing slash

	// firewall (resolved lazily in AuthorizePort; optional for static-only use)
	subscriptionID string
	resourceGroup  string
	nsgName        string
}

// newAzureProvider validates the Blob Storage credentials required for static
// hosting. Firewall/NSG settings are optional here and only enforced when a
// dynamic project needs a port opened (see AuthorizePort).
func newAzureProvider() (Provider, error) {
	account := os.Getenv("AZURE_STORAGE_ACCOUNT")
	key := os.Getenv("AZURE_STORAGE_KEY")
	if account == "" || key == "" {
		return nil, fmt.Errorf("AZURE_STORAGE_ACCOUNT and AZURE_STORAGE_KEY must be set")
	}

	cred, err := azblob.NewSharedKeyCredential(account, key)
	if err != nil {
		return nil, fmt.Errorf("invalid Azure storage credentials: %w", err)
	}
	serviceURL := fmt.Sprintf("https://%s.blob.core.windows.net/", account)
	client, err := azblob.NewClientWithSharedKeyCredential(serviceURL, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Azure blob client: %w", err)
	}

	container := os.Getenv("AZURE_BLOB_CONTAINER")
	if container == "" {
		container = "$web" // default static-website container
	}

	// AZURE_PUBLIC_BASE_URL lets you return the static-website endpoint
	// (https://<account>.z##.web.core.windows.net) instead of the raw blob URL.
	publicBase := os.Getenv("AZURE_PUBLIC_BASE_URL")
	if publicBase == "" {
		publicBase = fmt.Sprintf("https://%s.blob.core.windows.net/%s", account, container)
	}
	publicBase = strings.TrimRight(publicBase, "/")

	return &azureProvider{
		blobClient:     client,
		container:      container,
		publicBase:     publicBase,
		subscriptionID: os.Getenv("AZURE_SUBSCRIPTION_ID"),
		resourceGroup:  os.Getenv("AZURE_RESOURCE_GROUP"),
		nsgName:        os.Getenv("AZURE_NSG_NAME"),
	}, nil
}

func (a *azureProvider) Name() string { return "azure" }

// UploadStaticSite uploads a folder recursively to the Blob container and returns
// the URL to its index.html.
func (a *azureProvider) UploadStaticSite(localPath, keyPrefix string) (string, error) {
	ctx := context.Background()

	err := filepath.Walk(localPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(localPath, path)
		if err != nil {
			return err
		}
		blobName := filepath.ToSlash(filepath.Join(keyPrefix, relPath))

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		contentType := contentTypeByExtension(path)
		_, err = a.blobClient.UploadFile(ctx, a.container, blobName, file, &azblob.UploadFileOptions{
			HTTPHeaders: &blob.HTTPHeaders{BlobContentType: &contentType},
		})
		if err != nil {
			return fmt.Errorf("failed to upload blob %s: %w", blobName, err)
		}
		return nil
	})
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/%s/index.html", a.publicBase, keyPrefix), nil
}

// AuthorizePort opens an inbound TCP rule for port on the configured NSG. The
// rule name is derived from the port (idempotent across re-deploys) and an
// unused priority is chosen by scanning existing rules.
func (a *azureProvider) AuthorizePort(port int) error {
	if a.subscriptionID == "" || a.resourceGroup == "" || a.nsgName == "" {
		return fmt.Errorf("AZURE_SUBSCRIPTION_ID, AZURE_RESOURCE_GROUP and AZURE_NSG_NAME must be set; cannot open port %d", port)
	}

	ctx := context.Background()

	// Uses the standard Azure auth chain (env service principal or managed
	// identity): AZURE_CLIENT_ID, AZURE_TENANT_ID, AZURE_CLIENT_SECRET.
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return fmt.Errorf("failed to obtain Azure credentials: %w", err)
	}
	client, err := armnetwork.NewSecurityRulesClient(a.subscriptionID, cred, nil)
	if err != nil {
		return fmt.Errorf("failed to create NSG rules client: %w", err)
	}

	ruleName := fmt.Sprintf("autoship-port-%d", port)
	priority, err := a.nextFreePriority(ctx, client, ruleName)
	if err != nil {
		return err
	}

	poller, err := client.BeginCreateOrUpdate(ctx, a.resourceGroup, a.nsgName, ruleName, armnetwork.SecurityRule{
		Properties: &armnetwork.SecurityRulePropertiesFormat{
			Protocol:                 to.Ptr(armnetwork.SecurityRuleProtocolTCP),
			Access:                   to.Ptr(armnetwork.SecurityRuleAccessAllow),
			Direction:                to.Ptr(armnetwork.SecurityRuleDirectionInbound),
			Priority:                 to.Ptr(priority),
			SourceAddressPrefix:      to.Ptr("*"),
			SourcePortRange:          to.Ptr("*"),
			DestinationAddressPrefix: to.Ptr("*"),
			DestinationPortRange:     to.Ptr(strconv.Itoa(port)),
			Description:              to.Ptr("Auto-opened for container hosting"),
		},
	}, nil)
	if err != nil {
		return fmt.Errorf("failed to create NSG rule for port %d: %w", port, err)
	}
	if _, err := poller.PollUntilDone(ctx, nil); err != nil {
		return fmt.Errorf("NSG rule creation did not complete for port %d: %w", port, err)
	}
	return nil
}

// nextFreePriority lists existing inbound rules and returns a priority that is
// free. If a rule with ruleName already exists, its current priority is reused
// so a re-deploy updates the rule in place instead of colliding.
func (a *azureProvider) nextFreePriority(ctx context.Context, client *armnetwork.SecurityRulesClient, ruleName string) (int32, error) {
	used := map[int32]bool{}
	pager := client.NewListPager(a.resourceGroup, a.nsgName, nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return 0, fmt.Errorf("failed to list NSG rules: %w", err)
		}
		for _, rule := range page.Value {
			if rule == nil || rule.Properties == nil || rule.Properties.Priority == nil {
				continue
			}
			if rule.Name != nil && *rule.Name == ruleName {
				return *rule.Properties.Priority, nil // reuse existing slot
			}
			used[*rule.Properties.Priority] = true
		}
	}

	// Auto-opened rules live in 1000..4095; user/base rules typically sit below.
	for p := int32(1000); p <= 4095; p++ {
		if !used[p] {
			return p, nil
		}
	}
	return 0, fmt.Errorf("no free NSG priority available in range 1000-4095")
}
