# Release & Publish: autoship-deploy RPM

This document records the changes, build and publish steps, and troubleshooting notes for the autoship-deploy RPM and the GitHub Pages repository used as a yum repo.

Status (22 Aug 2025)
- Package: autoship-deploy
- Current version built: 1.0.1
- Published to branch: `gh-pages` (files served from https://Ashmit-Kumar.github.io/Auto-Ship/)
- RPM installed successfully on a test host (but pip wheel ABI mismatch was observed at install-time; service started after fixes)

What was changed
- Spec: `autoship-deploy.spec`
  - Bumped `Version` to `1.0.1` and updated changelog.
  - Removed duplicate `%files` entries to avoid warnings (keep `/opt/autoship` listed rather than listing nested files twice).
  - Switched to packaging vendored Python wheels inside the source tarball (CI downloads wheels into `autoship-scripts/wheels`) and installing dependencies by creating a virtualenv at `%post` and installing from the bundled wheels (no network required for runtime installs).
  - The spec requires `python3`, `python3-pip`, `nginx`, `certbot` on target hosts.

- CI Workflow: `.github/workflows/build-rpm.yml`
  - CI runs in an Amazon Linux container, installs `python3` + `pip`, downloads wheels listed in `autoship-scripts/requirements.txt` and creates a source tarball that contains `wheels/`.
  - `rpmbuild` creates the RPM. The workflow copies produced RPM(s) to `repo/`, runs `createrepo` and publishes `repo/` to the `gh-pages` branch.

Build & Publish (local reproduction)
1. Prepare build tools (Fedora example):
   sudo dnf install -y rpm-build redhat-rpm-config createrepo_c python3-pip

2. Create vendored source tarball and build RPM:
   ```bash
   VERSION=1.0.1
   rm -rf rpmbuild
   mkdir -p rpmbuild/{BUILD,RPMS,SOURCES,SPECS,SRPMS}
   STAGEDIR=$(mktemp -d)
   cp -a autoship-scripts "${STAGEDIR}/autoship-scripts"
   mkdir -p "${STAGEDIR}/autoship-scripts/wheels"
   python3 -m pip download -r autoship-scripts/requirements.txt -d "${STAGEDIR}/autoship-scripts/wheels"
   tar czf autoship-deploy-${VERSION}.tar.gz -C "${STAGEDIR}" autoship-scripts
   mv autoship-deploy-${VERSION}.tar.gz rpmbuild/SOURCES/
   cp autoship-scripts/autoship-deploy.spec rpmbuild/SPECS/
   rpmbuild -ba --define "_topdir $PWD/rpmbuild" rpmbuild/SPECS/autoship-deploy.spec
   ls -la rpmbuild/RPMS/*/*.rpm
   ```

3. Create `repo/` and repodata, then publish to `gh-pages`:
   ```bash
   mkdir -p repo
   cp -v rpmbuild/RPMS/*/*.rpm repo/
   createrepo_c repo/ || createrepo repo/

   # publish (safe method using a temporary clone)
   git clone --branch gh-pages --single-branch https://github.com/Ashmit-Kumar/Auto-Ship.git /tmp/gh-pages
   cd /tmp/gh-pages
   rm -rf ./*
   cp -a /home/ashmit/code/Auto-Ship/repo/* .
   git add -A
   git commit -m "Publish autoship-deploy ${VERSION} RPM + repodata" || true
   git push origin gh-pages --force
   rm -rf /tmp/gh-pages
   ```

Verification
- Raw branch: https://raw.githubusercontent.com/Ashmit-Kumar/Auto-Ship/gh-pages/
- Pages URLs (examples):
  - https://Ashmit-Kumar.github.io/Auto-Ship/repodata/repomd.xml
  - https://Ashmit-Kumar.github.io/Auto-Ship/autoship-deploy-1.0.1-1.fc41.noarch.rpm

Install on host (example)
- Add repo file `/etc/yum.repos.d/autoship-deploy.repo` with baseurl set to the GitHub Pages URL (root or `/repo/` depending how you published):
  ```ini
  [autoship-deploy]
  name=Autoship Deploy Repo
  baseurl=https://Ashmit-Kumar.github.io/Auto-Ship/
  enabled=1
  gpgcheck=0
  ```
- Then:
  sudo dnf clean all && sudo dnf makecache
  sudo dnf install -y autoship-deploy

Runtime issues observed and fixes
1. Wheel ABI mismatch during `%post` pip install
   - Error: pip could not find a matching distribution (e.g. charset_normalizer) when installing from bundled wheels.
   - Cause: CI downloaded wheels for the build environment Python ABI; target host Python may be a different minor version/ABI.
   - Fix: download wheels targeted to the target host Python ABI using pip download flags: `--platform`, `--python-version`, `--implementation`, `--abi`. Example (for CPython 3.11 x86_64):
     ```bash
     python3 -m pip download -r autoship-scripts/requirements.txt -d staged/wheels \
       --platform manylinux_2_31_x86_64 --only-binary=:all: \
       --python-version 311 --implementation cp --abi cp311
     ```
   - Recreate tarball, rebuild RPM, republish.

2. systemd NAMESPACE errors
   - Error in journal: "Failed to set up mount namespacing: /run/systemd/unit-root/etc/letsencrypt: No such file or directory"
   - Cause: systemd unit used Private/namespace settings referencing directories that did not exist on the host or ExecStart path mismatch.
   - Quick host fix: create the directory so systemd mount namespacing can set up, or adjust the unit. Example:
     ```bash
     sudo mkdir -p /etc/letsencrypt
     sudo systemctl daemon-reload
     sudo systemctl restart autoship.service
     ```
   - Consider updating `autoship.service` ExecStart to `/usr/bin/env python3 /opt/autoship/main.py` to avoid absolute python path mismatches.

Commits & branches
- Changes were made and committed locally then pushed to the repository (some edits were created directly in main). Typical commit messages used:
  - "spec: avoid duplicate /opt/autoship entries in %files"
  - "spec: bump version to 1.0.1"
  - "spec: update changelog for 1.0.1"
  - "ci: vendor python wheels into tarball; package venv in RPM"
  - `gh-pages` commits: "Publish autoship-deploy RPM + repodata" / "Publish autoship-deploy 1.0.1 RPM + repodata"

If you want me to:
- Generate a cleaned, human-readable release notes file and commit it (I can add `RELEASE_AND_PUBLISH.md` — already created).
- Recreate wheels for a specific target Python ABI — tell me the host's `python3 --version` and I will provide the exact `pip download` flags.
- Create a PR to merge these spec/workflow changes into `main`.

```
Create annotated tag: git tag -a v1.0.2 -m "Release v1.0.2"

Push that tag to the remote: git push origin v1.0.2
```
