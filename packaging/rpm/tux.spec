Name:           tux
Version:        ${VERSION}
Release:        1%{?dist}
Summary:        Terminal pet penguin

License:        MIT
URL:            https://github.com/imns/tux
Source0:        %{name}-%{version}.tar.gz

%description
Tux is a virtual pet penguin that lives in your terminal.
Interact with it through simple CLI commands: feed, play, and sleep.
The daemon runs in the background, maintaining Tux's state over time.

%prep
%setup -q

%build
go build -o tux ./cmd/tux
go build -o tuxd ./cmd/tuxd

%install
install -D -m 0755 tux %{buildroot}%{_bindir}/tux
install -D -m 0755 tuxd %{buildroot}%{_bindir}/tuxd
install -D -m 0644 packaging/tux.service %{buildroot}%{_unitdir}/tux.service

%files
%{_bindir}/tux
%{_bindir}/tuxd
%{_unitdir}/tux.service

%changelog
* ${DATE} Builder <builder@example.com> - ${VERSION}-1
- Initial package
