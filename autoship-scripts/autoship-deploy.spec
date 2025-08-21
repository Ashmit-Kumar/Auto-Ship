Name:           autoship-deploy
Version:        1.0.0
Release:        1%{?dist}
Summary:        AutoShip host worker and helper scripts (systemd service)
License:        MIT
URL:            https://github.com/Ashmit-Kumar/Auto-Ship
Source0:        autoship-deploy-%{version}.tar.gz
BuildArch:      noarch

# Runtime requirement: python3 must be available on target hosts. Python libraries are
# vendored and installed into a private virtualenv packaged inside the RPM so the
# package does not depend on distro-provided python modules like watchdog/requests.
Requires:       python3, nginx, certbot

%description
Host worker that watches deployment requests written by the Auto-Ship server and
configures nginx, DNS and SSL on the host. Installs scripts under /opt/autoship
and a systemd unit to run the worker. Python runtime dependencies are installed
into a bundled virtualenv contained in /opt/autoship/venv.

%prep
%setup -q -n autoship-scripts

%build
# no build step (scripts only)

%install
rm -rf %{buildroot}
# install application files into /opt/autoship
mkdir -p %{buildroot}/opt/autoship
cp -a * %{buildroot}/opt/autoship/

# install systemd unit
install -D -m 644 autoship.service %{buildroot}/etc/systemd/system/autoship.service

# install example env as config (do not include real secrets)
install -m 644 .env.example %{buildroot}/opt/autoship/.env.example

# ===== create a virtualenv and install vendored wheels =====
# If Python3 is available on the build system, create a venv in the buildroot
# and install wheels bundled under /opt/autoship/wheels (these are packaged
# into the source tarball by CI). After installation, remove the wheels to
# keep the installed payload smaller.
%{__python3} -V >/dev/null 2>&1 || true
if [ -x %{__python3} ]; then
  # create venv inside buildroot
  %{__python3} -m venv %{buildroot}/opt/autoship/venv || true
  if [ -x %{buildroot}/opt/autoship/venv/bin/pip ]; then
    %{buildroot}/opt/autoship/venv/bin/pip install --no-index --find-links %{buildroot}/opt/autoship/wheels -r %{buildroot}/opt/autoship/requirements.txt || true
  fi
  rm -rf %{buildroot}/opt/autoship/wheels || true
fi

%files
%defattr(-,root,root,-)
/opt/autoship
/opt/autoship/venv
/etc/systemd/system/autoship.service
%config(noreplace) /opt/autoship/.env.example
%doc Deploy-package.md PKG_AND_PUBLISH.md

%post
# enable and start service if systemd available
if [ -x /usr/bin/systemctl ]; then
  systemctl daemon-reload || true
  # enable but do not override existing user-managed unit
  systemctl enable --now autoship.service || true
fi

%preun
if [ $1 -eq 0 ]; then
  # package is being removed (not upgraded)
  if [ -x /usr/bin/systemctl ]; then
    systemctl stop autoship.service || true
    systemctl disable autoship.service || true
  fi
fi

%postun
# always reload systemd daemon after uninstall/upgrade
if [ -x /usr/bin/systemctl ]; then
  systemctl daemon-reload || true
fi

%changelog
* Thu Aug 21 2025 Ashmit-Kumar - 1.0.0-1
- Vendored Python dependencies into package and install into bundled virtualenv
- Require only python3, nginx and certbot at runtime
- Include systemd unit and example .env as config
