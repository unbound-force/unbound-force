# SPDX-License-Identifier: Apache-2.0

%global goipath github.com/unbound-force/unbound-force
%global base_url https://%{goipath}
%global debug_package %{nil}

Name:           unbound-force
# Version is updated automatically by Packit during propose_downstream
Version:        0.14.0
Release:        1%{?dist}
Summary:        AI agent swarm specification framework toolkit
License:        Apache-2.0
URL:            %{base_url}
Source0:        %{base_url}/archive/refs/tags/v%{version}.tar.gz

BuildRequires:  golang
BuildRequires:  go-rpm-macros

%gometa -f

%description
%{name} is a CLI toolkit for managing AI agent swarms themed
as a superhero team. It scaffolds agent configurations,
validates specifications, runs health checks, and manages
sandboxed development environments.

%prep
%goprep -k

%build
BUILD_DATE_GO=$(date -u +'%%Y-%%m-%%dT%%H:%%M:%%SZ')

# Set up environment variables and flags to build properly
%set_build_flags

# Ldflags match .goreleaser.yaml configuration
GO_LD_EXTRAFLAGS="-s -w \
  -X main.version=%{version} \
  -X main.commit=%{version} \
  -X main.date=${BUILD_DATE_GO}"

export GO111MODULE=on

GO_BUILD_BINDIR=./bin
mkdir -p ${GO_BUILD_BINDIR}

go build -buildmode=pie \
  -o ${GO_BUILD_BINDIR}/unbound-force \
  -ldflags="${GO_LD_EXTRAFLAGS}" \
  ./cmd/unbound-force

%install
install -d %{buildroot}%{_bindir}
install -p -m 0755 bin/unbound-force %{buildroot}%{_bindir}/unbound-force
ln -sf unbound-force %{buildroot}%{_bindir}/uf

%check
go test -race -count=1 ./...

%files
%attr(0755, root, root) %{_bindir}/unbound-force
%{_bindir}/uf
%license LICENSE
%doc README.md

%changelog
* Wed May 06 2026 Marcus Burghardt <maburgha@redhat.com> - 0.14.0-1
- Initial RPM spec for Fedora packaging via Packit
