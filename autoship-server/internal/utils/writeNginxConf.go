package utils
import (
    "fmt"
    "os"
    "os/exec"
)



func writeNginxConf(subdomain string, hostPort int) error {
    // Step 1: Temporary HTTP config for Certbot challenge
    if subdomain == "" || hostPort <= 0 {
        return fmt.Errorf("invalid subdomain or host port") 
    }

    fmt.Println("Writing temporary NGINX config for Certbot challenge...")
    tempConf := fmt.Sprintf(`
	server {
		listen 80;
		server_name %s;

		location /.well-known/acme-challenge/ {
			root /var/www/certbot;
		}
	}
	`, subdomain)

    tempPath := fmt.Sprintf("/etc/nginx/sites-available/%s-temp.conf", subdomain)
    err := os.WriteFile(tempPath, []byte(tempConf), 0644)
    if err != nil {
        return fmt.Errorf("writing temp nginx conf: %w", err)
    }

    tempSymlink := fmt.Sprintf("/etc/nginx/sites-enabled/%s-temp.conf", subdomain)
    _ = os.Remove(tempSymlink)
    err = os.Symlink(tempPath, tempSymlink)
    if err != nil {
        return fmt.Errorf("symlinking temp nginx conf: %w", err)
    }

    // Reload NGINX for Certbot to succeed
    if err := exec.Command("systemctl", "reload", "nginx").Run(); err != nil {
        return fmt.Errorf("reloading nginx: %w", err)
    }

    // Step 2: Generate SSL with Certbot (ensure certbot is installed)
    certbotCmd := exec.Command("certbot", "certonly", "--webroot", "-w", "/var/www/certbot", "--agree-tos", "--no-eff-email", "--email", "your-email@example.com", "-d", subdomain, "--non-interactive")
    certbotCmd.Stdout = os.Stdout
    certbotCmd.Stderr = os.Stderr
    if err := certbotCmd.Run(); err != nil {
        return fmt.Errorf("certbot failed: %w", err)
    }

    // Step 3: Final HTTPS + HTTP redirect conf
    finalConf := fmt.Sprintf(`
	server {
		listen 80;
		server_name %s;

		location / {
			return 301 https://$host$request_uri;
		}
	}

	server {
		listen 443 ssl;
		server_name %s;

		ssl_certificate /etc/letsencrypt/live/%s/fullchain.pem;
		ssl_certificate_key /etc/letsencrypt/live/%s/privkey.pem;

		ssl_protocols TLSv1.2 TLSv1.3;
		ssl_ciphers HIGH:!aNULL:!MD5;

		location / {
			proxy_pass http://localhost:%d;
			proxy_http_version 1.1;
			proxy_set_header Host $host;
			proxy_set_header X-Real-IP $remote_addr;
			proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
			proxy_set_header X-Forwarded-Proto $scheme;
		}
	}
	`, subdomain, subdomain, subdomain, subdomain, hostPort)

    finalPath := fmt.Sprintf("/etc/nginx/sites-available/%s.conf", subdomain)
    err = os.WriteFile(finalPath, []byte(finalConf), 0644)
    if err != nil {
        return fmt.Errorf("writing final nginx conf: %w", err)
    }

    finalSymlink := fmt.Sprintf("/etc/nginx/sites-enabled/%s.conf", subdomain)
    _ = os.Remove(finalSymlink)
    err = os.Symlink(finalPath, finalSymlink)
    if err != nil {
        return fmt.Errorf("symlinking final nginx conf: %w", err)
    }

    // Clean up temp conf
    _ = os.Remove(tempPath)
    _ = os.Remove(tempSymlink)

    // Reload NGINX again with final HTTPS config
    if err := exec.Command("systemctl", "reload", "nginx").Run(); err != nil {
        return fmt.Errorf("reloading nginx after final conf: %w", err)
    }

    return nil
}

// WriteNginxProxyConf writes a NGINX configuration for proxying requests for static websites
func WriteNginxProxyConf(subdomain, targetURL string) error {
	tempConf := fmt.Sprintf(`
	server {
		listen 80;
		server_name %s;

		location /.well-known/acme-challenge/ {
			root /var/www/certbot;
		}
	}
	`, subdomain)

	tempPath := fmt.Sprintf("/etc/nginx/sites-available/%s-temp.conf", subdomain)
	_ = os.WriteFile(tempPath, []byte(tempConf), 0644)
	_ = os.Symlink(tempPath, fmt.Sprintf("/etc/nginx/sites-enabled/%s-temp.conf", subdomain))
	_ = exec.Command("systemctl", "reload", "nginx").Run()

	certbotCmd := exec.Command("certbot", "certonly", "--webroot", "-w", "/var/www/certbot", "--agree-tos", "--no-eff-email", "--email", "your-email@example.com", "-d", subdomain, "--non-interactive")
	certbotCmd.Stdout = os.Stdout
	certbotCmd.Stderr = os.Stderr
	if err := certbotCmd.Run(); err != nil {
		return fmt.Errorf("certbot failed: %w", err)
	}

	finalConf := fmt.Sprintf(`
	server {
		listen 80;
		server_name %s;
		location / {
			return 301 https://$host$request_uri;
		}
	}

	server {
		listen 443 ssl;
		server_name %s;

		ssl_certificate /etc/letsencrypt/live/%s/fullchain.pem;
		ssl_certificate_key /etc/letsencrypt/live/%s/privkey.pem;

		location / {
			proxy_pass %s;
			proxy_set_header Host $host;
			proxy_set_header X-Real-IP $remote_addr;
			proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
			proxy_set_header X-Forwarded-Proto $scheme;
		}
	}
	`, subdomain, subdomain, subdomain, subdomain, targetURL)

	finalPath := fmt.Sprintf("/etc/nginx/sites-available/%s.conf", subdomain)
	_ = os.WriteFile(finalPath, []byte(finalConf), 0644)
	_ = os.Symlink(finalPath, fmt.Sprintf("/etc/nginx/sites-enabled/%s.conf", subdomain))
	_ = os.Remove(tempPath)
	_ = os.Remove(fmt.Sprintf("/etc/nginx/sites-enabled/%s-temp.conf", subdomain))

	return exec.Command("systemctl", "reload", "nginx").Run()
}


// calling each func in dnsRecord.go, subdomain.go, and writeNginxConf.go and connecting them through the main flow
// using Projects.go file
// automatic routing for static s3 still left to do
// automate Hostinger DNS API integration 
// AddDNSRecordWithRetry adds a DNS record with retries in case of failure.

// data in .env file