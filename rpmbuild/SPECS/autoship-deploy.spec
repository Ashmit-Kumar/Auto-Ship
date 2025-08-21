Name:           autoship-deploy
Version:        1.0.0
Release:        1%{?dist}
Summary:        AutoShip host worker and helper scripts (systemd service)
License:        MIT
URL:            https://github.com/Ashmit-Kumar/Auto-Ship
Source0:        autoship-deploy-%{version}.tar.gz
BuildArch:      noarch

# Vendored wheels are packaged into the source tarball by CI. At install time
# the package will create a private virtualenv under /opt/autoship/venv and
# install the wheels from /opt/autoship/wheels so no distro RPMs are required.
Requires:       python3, python3-pip, nginx, certbot

%description
Host worker that watches deployment requests written by the Auto-Ship server and
configures nginx, DNS and SSL on the host. Installs scripts under /opt/autoship
and a systemd unit to run the worker. Python runtime dependencies are installed
into a virtualenv created on the target host during package postinstall.

%prep
%setup -q -n autoship-scripts

%build
# no build step (scripts only)

%install
rm -rf %{buildroot}
# install application files into /opt/autoship (includes wheels/ when CI vendors them)
mkdir -p %{buildroot}/opt/autoship
cp -a * %{buildroot}/opt/autoship/

# install systemd unit
install -D -m 644 autoship.service %{buildroot}/etc/systemd/system/autoship.service

# install example env as config (do not include real secrets)
install -m 644 .env.example %{buildroot}/opt/autoship/.env.example

%files
%defattr(-,root,root,-)
/opt/autoship
/opt/autoship/wheels
/etc/systemd/system/autoship.service
%config(noreplace) /opt/autoship/.env.example
%doc Deploy-package.md PKG_AND_PUBLISH.md

%post
# Create a private virtualenv and install vendored wheels from the package (no network)
if [ -x /usr/bin/python3 ]; then
  if [ ! -d /opt/autoship/venv ]; then
    /usr/bin/python3 -m venv /opt/autoship/venv || true
  fi
  if [ -x /opt/autoship/venv/bin/pip ]; then
    if [ -d /opt/autoship/wheels ]; then
      /opt/autoship/venv/bin/pip install --no-index --find-links /opt/autoship/wheels -r /opt/autoship/requirements.txt || true
      # optionally remove wheels after install to save space
      rm -rf /opt/autoship/wheels || true
    fi
  fi
fi

# enable/start service if systemd available
if [ -x /usr/bin/systemctl ]; then
  systemctl daemon-reload || true
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
- Bundle python wheels in source tarball; create venv at %post and install from wheels
- Require python3 and python3-pip; avoid BUILDROOT venv contamination