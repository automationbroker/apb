%if 0%{?fedora} || 0%{?rhel} >= 6
%global with_devel 1
%global with_bundled 0
%global with_debug 0
%global with_check 0
%global with_unit_test 0
%else
%global with_devel 0
%global with_bundled 0
%global with_debug 0
%global with_check 0
%global with_unit_test 0
%endif

%if 0%{?with_debug}
%global _dwz_low_mem_die_limit 0
%else
%global	debug_package	%{nil}
%endif

%global	provider github
%global	provider_tld com
%global project automationbroker
%global repo apb

%global provider_prefix %{provider}.%{provider_tld}/%{project}/%{repo}
%global import_path %{provider_prefix}

%if 0%{?copr}
%define build_timestamp .%(date +"%Y%m%d%H%M%%S")
%else
%define build_timestamp %{nil}
%endif

# %define selinux_variants targeted
%define moduletype apps
%define modulename apb

Name: %{repo}
Version: 1.9.0
Release: 1%{build_timestamp}%{?dist}
Summary: APB CLI
License: ASL 2.0
URL: https://%{provider_prefix}
Source0: %{name}-%{version}.tar.gz

# e.g. el6 has ppc64 arch without gcc-go, so EA tag is required
#ExclusiveArch: %%{?go_arches:%%{go_arches}}%%{!?go_arches:%%{ix86} x86_64 %{arm}}
ExclusiveArch: %{ix86} x86_64 %{arm} aarch64 ppc64le %{mips} s390x
# If go_compiler is not set to 1, there is no virtual provide. Use golang instead.
BuildRequires: %{?go_compiler:compiler(go-compiler)}%{!?go_compiler:golang}

%if ! 0%{?with_bundled}
%endif

%description
%{summary}

%if 0%{?with_devel}
%package devel
Summary: %{summary}
BuildArch: noarch

Requires: %{?go_compiler:compiler(go-compiler)}%{!?go_compiler:golang}

%description devel
devel for %{name}
%{import_path} prefix.
%endif

%prep
%setup -q -n %{repo}-%{version}
ln -sf vendor src
mkdir -p src/github.com/automationbroker/apb
cp -r pkg src/github.com/automationbroker/apb
cp -r cmd src/github.com/automationbroker/apb

%build
export GOPATH=$(pwd):%{gopath}
go build -i -ldflags "-s -w" -o apb

%if 0%{?rhel} >= 6
    distver=rhel%{rhel}
%endif
%if 0%{?fedora} >= 18
    distver=fedora%{fedora}
%endif

rm -rf src

%install
install -d -p %{buildroot}%{_bindir}
install -p -m 755 apb %{buildroot}%{_bindir}/apb

# source codes for building projects
%if 0%{?with_devel}
install -d -p %{buildroot}/%{gopath}/src/%{import_path}/
# find all *.go but no *_test.go files and generate devel.file-list
for file in $(find . -iname "*.go" \! -iname "*_test.go" | grep -v "^./Godeps") ; do
    echo "%%dir %%{gopath}/src/%%{import_path}/$(dirname $file)" >> devel.file-list
    install -d -p %{buildroot}/%{gopath}/src/%{import_path}/$(dirname $file)
    cp -pav $file %{buildroot}/%{gopath}/src/%{import_path}/$file
    echo "%%{gopath}/src/%%{import_path}/$file" >> devel.file-list
done
for file in $(find . -iname "*.proto" | grep -v "^./Godeps") ; do
    echo "%%dir %%{gopath}/src/%%{import_path}/$(dirname $file)" >> devel.file-list
    install -d -p %{buildroot}/%{gopath}/src/%{import_path}/$(dirname $file)
    cp -pav $file %{buildroot}/%{gopath}/src/%{import_path}/$file
    echo "%%{gopath}/src/%%{import_path}/$file" >> devel.file-list
done
%endif

%if 0%{?with_devel}
sort -u -o devel.file-list devel.file-list
%endif

%files
%license LICENSE
%{_bindir}/apb

%if 0%{?with_devel}
%files devel -f devel.file-list
%license LICENSE
%dir %{gopath}/src/%{provider}.%{provider_tld}/%{project}
%dir %{gopath}/src/%{import_path}
%endif

%changelog
