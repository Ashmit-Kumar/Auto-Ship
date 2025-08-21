Name:           autoship-deploy
Version:        1.0.0
Release:        1%{?dist}
Summary:        AutoShip host worker and helper scripts (systemd service)
License:        MIT
URL:            https://github.com/Ashmit-Kumar/Auto-Ship
Source0:        autoship-deploy-%{version}.tar.gz
BuildArch:      noarch

# Adjust Requires for your target distro (Amazon Linux / CentOS / Fedora)
Requires:       python3, python3-watchdog, python3-requests, nginx, certbot

%description
Host worker that watches deployment requests written by the Auto-Ship server and
configures nginx, DNS and SSL on the host. Installs scripts under /opt/autoship
and a systemd unit to run the worker.

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

%files
%defattr(-,root,root,-)
/opt/autoship
/etc/systemd/system/autoship.service
%doc Deploy-package.md PKG_AND_PUBLISH.md
# mark .env.example as config (optional replacement behavior handled by RPM)
# %config(noreplace) /opt/autoship/.env.example

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
- Initial package
