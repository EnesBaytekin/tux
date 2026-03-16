Name:           tux
Version:        ${VERSION}
Release:        1%{?dist}
Summary:        Terminal pet penguin

License:        MIT
URL:            https://github.com/enesbaytekin/tux
Source0:        %{name}-%{version}.tar.gz

%description
Tux is a virtual pet penguin that lives in your terminal.
Interact with it through simple CLI commands: feed, play, and sleep.

%prep
%setup -q

%build
go build -o tux ./cmd/tux

%install
install -D -m 0755 tux %{buildroot}%{_bindir}/tux

%files
%{_bindir}/tux

%changelog
* ${DATE} Builder <enes@baytekin.dev> - ${VERSION}-1
- Initial package
