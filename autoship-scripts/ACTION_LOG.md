# Autoship packaging & service — action log and commands

This document records what I changed in the `autoship-scripts` folder, and lists the exact commands to build the source tarball, create the RPM, publish a GitHub Pages YUM repo, and install the package on EC2. Each command is followed by a short explanation of what it does.

Files created/edited
- autoship-deploy.spec — RPM spec used by `rpmbuild` (installs /opt/autoship and systemd unit).
- autoship.service — systemd unit example used by the package.
- PKG_AND_PUBLISH.md — full guide (detailed steps and best practices).
- PKG_AND_PUBLISH.md and Deploy-package.md reviewed and linked.
- .env.example — example environment file template (kept as config in package).
- autoship.service (copy) and other helper docs were added to `autoship-scripts/`.
- ssl_utils.py, nginx_utils.py, response_utils.py, main.py, config.py were reviewed and adjusted earlier to handle certbot/webroot and atomic responses.

One-shot commands to build and package (run from repo root)

1) Prepare rpmbuild tree and macro (do once per builder account)

```sh
mkdir -p ~/rpmbuild/{BUILD,RPMS,SOURCES,SPECS,SRPMS}
# tell rpmbuild to use ~/rpmbuild as topdir
echo '%_topdir %(echo $HOME)/rpmbuild' > ~/.rpmmacros
```
Explanation: creates the standard rpmbuild directories and a macro so rpmbuild stores build artifacts under your home.

2) Create source tarball containing only `autoship-scripts/`

```sh
VERSION=1.0.0
# create compressed tarball with the directory content
tar czf autoship-deploy-${VERSION}.tar.gz autoship-scripts/
# move tarball to rpmbuild SOURCES directory
mv autoship-deploy-${VERSION}.tar.gz ~/rpmbuild/SOURCES/
```
Explanation: rpmbuild expects a Source0 tarball in SOURCES. The tarball should unpack to the directory name `autoship-scripts` because the spec uses `%setup -n autoship-scripts`.

3) Copy the spec and run rpmbuild

```sh
cp autoship-scripts/autoship-deploy.spec ~/rpmbuild/SPECS/
# build binary and source RPMs
rpmbuild -ba ~/rpmbuild/SPECS/autoship-deploy.spec
```
Explanation: `rpmbuild -ba` runs prep/build/install steps defined in the spec and produces RPMs in `~/rpmbuild/RPMS/`.

4) Locate the built RPM

```sh
ls -l ~/rpmbuild/RPMS/noarch/
```
Explanation: lists the produced RPM(s). For a noarch package the RPM lives under `noarch`.

5) Prepare the static repo directory and metadata

```sh
mkdir repo
cp ~/rpmbuild/RPMS/noarch/autoship-deploy-1.0.0-1.noarch.rpm repo/
# generate repodata/ directory with createrepo
createrepo repo/
```
Explanation: `createrepo` builds the repodata metadata that yum/dnf needs to consume a repo.

6) Publish `repo/` to gh-pages branch (example using ssh remote)

```sh
cd repo
git init
git remote add origin git@github.com:<youruser>/autoship-deploy.git
git checkout -b gh-pages
git add .
git commit -m "Add rpm and repodata"
git push origin gh-pages --force
```
Explanation: pushes the static repo to GitHub Pages; GitHub will serve files from gh-pages branch.

Client-side setup (on each EC2 host)

1) Add repo file `/etc/yum.repos.d/autoship-deploy.repo`

Create a file with:

```ini
[autoship-deploy]
name=Autoship Deploy Repo
baseurl=https://<youruser>.github.io/autoship-deploy/repo/
enabled=1
gpgcheck=0
```
Explanation: points yum/dnf to the GitHub Pages URL where the repo metadata is served.

2) Install package

```sh
sudo yum clean all
sudo yum install -y autoship-deploy
```
Explanation: installs the RPM and runs the `%post` scriptlets (enabling the systemd unit if present).

3) After install: create per-host `.env` and secure it

```sh
sudo cp /opt/autoship/.env.example /opt/autoship/.env
# edit /opt/autoship/.env and set actual values (CLOUDFLARE token, zone id, EC2 public IP, etc.)
sudo chown root:root /opt/autoship/.env
sudo chmod 600 /opt/autoship/.env
# Ensure service is running
sudo systemctl daemon-reload
sudo systemctl enable --now autoship.service
sudo journalctl -u autoship.service -f
```
Explanation: package doesn't contain real secrets — operator must populate `.env` per-host.

Commands & tests for certbot / nginx issues (debug)

- Test nginx configuration:

```sh
sudo nginx -t
```
Checks nginx config syntax.

- Test certbot webroot issuance (staging):

```sh
sudo mkdir -p /var/www/autoship/<domain>/.well-known/acme-challenge
sudo chown -R nginx:nginx /var/www/autoship
sudo certbot certonly --webroot -w /var/www/autoship/<domain> -d <domain> --staging --force-renewal --rsa-key-size 2048 --config-dir /var/lib/autoship/certbot/config --work-dir /var/lib/autoship/certbot/work --logs-dir /var/lib/autoship/certbot/logs
```
Explanation: obtains a cert using webroot plugin (use staging first to avoid rate limits).

Notes about the `.spec` file
- The spec controls how rpmbuild constructs the package. Key parts:
  - Metadata (Name, Version, Requires) — used by package managers.
  - `%prep` unpacks Source0 tarball.
  - `%install` copies files into buildroot with the final target layout (`/opt/autoship`, `/etc/systemd/system/autoship.service`).
  - `%files` declares which files are included in the RPM and which are docs/configs.
  - `%post`/`%preun` scriptlets run on install/uninstall to enable/disable the systemd service.

What I will continue doing next (pick one or more):
- Create a GitHub Actions workflow that automates tarball → rpmbuild → createrepo → push to `gh-pages`.
- Add RPM signing steps and show how to import the GPG public key on clients (switch gpgcheck to 1).
- Provide an `fpm` quick-pack script and `after-install.sh` for fast iteration.
- Update `nginx_utils.py` to ensure S3 proxying does not break `nginx -t` (avoid upstream blocks) and ensure `.well-known` is served for webroot.

Which of the follow-ups should I do next? Reply with the number(s).
