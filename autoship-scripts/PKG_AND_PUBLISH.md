# Package & Publish `autoship-deploy` (RPM) — Complete Step-by-Step

This document explains every step to package the `autoship-scripts` into an RPM, publish it as a static YUM/DNF repo on GitHub Pages, and install/update it on EC2 hosts. It includes alternatives (fpm), systemd integration, `.env` handling, signing, CI automation, and troubleshooting.

Target outcome
- A reproducible RPM named `autoship-deploy-VERSION.rpm` that installs files under `/opt/autoship`, places a systemd unit at `/etc/systemd/system/autoship.service`, and does not contain real secrets.
- A static YUM/DNF repository served from GitHub Pages at `https://<your-username>.github.io/autoship-deploy/repo/`.
- EC2 hosts configured with `/etc/yum.repos.d/autoship-deploy.repo` can `yum install autoship-deploy` and later `yum update`.

Prerequisites (build machine)
- A Linux build host (Fedora/RHEL/CentOS/AlmaLinux/Ubuntu WSL) with `rpmbuild` or `fpm` available.
- `createrepo` installed for building repo metadata.
- Git configured and access to the `autoship-deploy` GitHub repo.
- (Optional) GPG key for signing packages.

Directories and conventions
- Use an `rpmbuild` tree under `~/rpmbuild` (default) for rpmbuild.
- Source tarball will contain only `autoship-scripts/` content.
- Do NOT include `/opt/autoship/.env` with secrets in the package; include `.env.example` only.

Part A — Build RPM (recommended: rpmbuild)

1) Prepare rpmbuild environment

```bash
mkdir -p ~/rpmbuild/{BUILD,RPMS,SOURCES,SPECS,SRPMS}
echo '%_topdir %(echo $HOME)/rpmbuild' > ~/.rpmmacros
```

2) Create source tarball (from repo root)

```bash
# from repo root where autoship-scripts/ exists
tar czf autoship-deploy-1.0.0.tar.gz autoship-scripts/
mv autoship-deploy-1.0.0.tar.gz ~/rpmbuild/SOURCES/
```

3) Create spec file `~/rpmbuild/SPECS/autoship-deploy.spec`

Use this minimal spec (copy and edit values):

```
Name:           autoship-deploy
Version:        1.0.0
Release:        1%{?dist}
Summary:        AutoShip host worker and helper scripts
License:        MIT
URL:            https://github.com/<youruser>/Auto-Ship
Source0:        autoship-deploy-1.0.0.tar.gz
BuildArch:      noarch
Requires:       python3, python3-watchdog, python3-requests, nginx, certbot

%description
Host worker that watches deployment requests and configures nginx, DNS, SSL.

%prep
%setup -q -n autoship-scripts

%install
rm -rf %{buildroot}
mkdir -p %{buildroot}/opt/autoship
cp -a * %{buildroot}/opt/autoship/
# install systemd unit
install -D -m 644 autoship.service %{buildroot}/etc/systemd/system/autoship.service
install -m 644 .env.example %{buildroot}/opt/autoship/.env.example

%files
/opt/autoship
/etc/systemd/system/autoship.service
%doc Deploy-package.md README.md

%post
# enable and start service if systemd available
if [ -x /usr/bin/systemctl ]; then
  systemctl daemon-reload || true
  systemctl enable --now autoship.service || true
fi

%preun
if [ $1 -eq 0 ]; then
  systemctl stop autoship.service || true
  systemctl disable autoship.service || true
fi

%changelog
* $(date '+%a %b %d %Y') ${USER} - 1.0.0-1
- Initial package
```

Notes for spec
- `Requires` should match package names on target OS. Adjust for Amazon Linux 2/2023.
- Mark `.env.example` as config; you may optionally change to `%config(noreplace) /opt/autoship/.env.example`.

4) Build RPM

```bash
rpmbuild -ba ~/rpmbuild/SPECS/autoship-deploy.spec
# RPM will be in ~/rpmbuild/RPMS/noarch/
```

5) Test the RPM locally (clean VM recommended)

```bash
sudo rpm -Uvh ~/rpmbuild/RPMS/noarch/autoship-deploy-1.0.0-1.noarch.rpm
# Verify files installed under /opt/autoship, unit exists
sudo systemctl status autoship.service
```

Part B — Quick packaging (alternative: fpm)

If you prefer very fast iteration, use `fpm`:

```bash
gem install --user-install fpm
cd autoship-scripts
fpm -s dir -t rpm -n autoship-deploy -v 1.0.0 --prefix /opt/autoship --config-files /opt/autoship/.env.example --rpm-os linux .
```

This produces `autoship-deploy-1.0.0.rpm` in the current dir. Adjust `--depends` flags if needed.

Part C — Create repo metadata and publish to GitHub Pages (GitHub Pages approach)

1) Prepare `repo/` directory and copy RPM(s)

```bash
mkdir repo
cp ~/rpmbuild/RPMS/noarch/autoship-deploy-1.0.0-1.noarch.rpm repo/
```

2) Run createrepo (generates repodata)

```bash
createrepo repo/
```

3) Commit `repo/` content to `gh-pages` branch of a GitHub repo named `autoship-deploy`

- Create repo `autoship-deploy` on GitHub
- Create `gh-pages` branch and push `repo/` contents there

```bash
git init repo
cd repo
git remote add origin git@github.com:<youruser>/autoship-deploy.git
git checkout -b gh-pages
git add .
git commit -m "Add repo metadata and rpm"
git push origin gh-pages --force
```

GitHub Pages will serve the directory at:

```
https://<youruser>.github.io/autoship-deploy/repo/
```

Part D — Client configuration on EC2

1) Create repo file on EC2 as `/etc/yum.repos.d/autoship-deploy.repo`:

```ini
[autoship-deploy]
name=Autoship Deploy Repo
baseurl=https://<youruser>.github.io/autoship-deploy/repo/
enabled=1
gpgcheck=0
```

2) Install package

```bash
sudo yum clean all
sudo yum install -y autoship-deploy
```

3) After install
- Copy `/opt/autoship/.env.example` to `/opt/autoship/.env` and fill values per host.
- Secure the file:

```bash
sudo chown root:root /opt/autoship/.env
sudo chmod 600 /opt/autoship/.env
```

- Start/enable service (if not already started by RPM postinstall):

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now autoship.service
sudo journalctl -u autoship.service -f
```

Part E — Signing RPMs (recommended for production)

1) Create or import a GPG key and export public key for clients

```bash
gpg --gen-key
gpg --export -a "Your Name" > public.gpg
```

2) Sign RPM during build with `rpmsign` or `gpg` configuration. Example (after rpmbuild):

```bash
rpmsign --addsign ~/rpmbuild/RPMS/noarch/autoship-deploy-1.0.0-1.noarch.rpm
```

3) On client, import public key and enable gpgcheck

```bash
sudo rpm --import public.gpg
# set gpgcheck=1 in /etc/yum.repos.d/autoship-deploy.repo
```

Part F — CI: GitHub Actions to build RPM and publish repo

High level steps for GitHub Actions workflow
- On push tag or release: create source tarball, build RPM (use a suitable runner or Docker with rpmbuild), sign RPM (if signing), create `repo/`, run `createrepo`, and push `repo/` contents to `gh-pages` branch.

I can generate a ready-to-use `build-rpm.yml` workflow if you want — include secrets for GPG private key and GitHub personal access token for pushing to gh-pages.

Part G — Atomic publishing best practices

- Build repo in a temporary directory, then move/rename atomically on the gh-pages branch to avoid partially updated `repodata/`.
- Example: build to `repo_tmp/`, commit `repo_tmp/` as `repo/` in a single commit replace, then push.

Part H — Handling `.env` and secrets

- Do NOT include `/opt/autoship/.env` in package. Include only `.env.example` and mark as `%config(noreplace)` if desired.
- After package installation, instruct the operator to create `/opt/autoship/.env` and set permissions to 600.
- Optionally provide a small post-install helper script to copy `.env.example` to `/opt/autoship/.env` with a safe prompt (do not automatically inject secrets).

Part I — Testing & troubleshooting checklist

- Validate RPM on a clean VM matching target OS.
- After install, verify:
  - Files are present in `/opt/autoship`.
  - Systemd unit exists and is active.
  - `/opt/autoship/.env` exists and is secure.
- If certbot/nginx errors occur: inspect `journalctl -u autoship.service`, test `sudo certbot --nginx -d example.com --staging` manually, check `/var/lib/autoship/certbot/logs/letsencrypt.log`.
- If DNS API calls fail: verify outbound connectivity and correct Cloudflare token/zone.

Part J — Cleanup and rollback

- To remove package:

```bash
sudo yum remove -y autoship-deploy
```

- To rollback repo content on GitHub Pages: restore previous commit on `gh-pages` branch and push.

Extras & recommendations
- Use GitHub Actions to automate: build, sign, create repo metadata, push to gh-pages.
- Consider hosting repo on S3/CloudFront for private repos + signed access.
- For enterprise, enable `gpgcheck=1` and sign RPMs.

Want me to generate now
- `~/rpmbuild/SPECS/autoship-deploy.spec` filled with your exact service unit and paths, or
- A GitHub Actions `build-rpm.yml` that builds RPM + createrepo + pushes to `gh-pages`, or
- An `fpm` one-liner + `after-install.sh` script for rapid testing.

Pick one and I will create the file.
