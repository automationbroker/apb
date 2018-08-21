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
%global gopath /usr/share/gocode

%if 0%{?copr}
%define build_timestamp .%(date +"%Y%m%d%H%M%%S")
%else
%define build_timestamp %{nil}
%endif

# %define selinux_variants targeted
%define moduletype apps
%define modulename apb

Name: %{repo}
Version: 1.9.3
Release: 1%{build_timestamp}%{?dist}
Summary: APB CLI
License: ASL 2.0
URL: https://%{provider_prefix}
Source0: %{name}-%{version}.tar.gz

# e.g. el6 has ppc64 arch without gcc-go, so EA tag is required
#ExclusiveArch: %%{?go_arches:%%{go_arches}}%%{!?go_arches:%%{ix86} x86_64 %{arm}}
ExclusiveArch: %{ix86} x86_64 %{arm} aarch64 ppc64le %{mips} s390x
BuildRequires: golang

%if ! 0%{?with_bundled}
%endif

%description
%{summary}

%package container-scripts
Summary: scripts required for running apb in a container
BuildArch: noarch
Requires: %{name}

%description container-scripts
containers scripts for apb

%if 0%{?with_devel}
%package devel
Summary: %{summary}
BuildArch: noarch

Requires: golang

%description devel
devel for %{name}
%{import_path} prefix.
%endif

%prep
%setup -q -n %{repo}-%{version}
%if !0%{?copr}
patch -p0 < downstream.patch
%endif
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
install -m 755 apb-wrapper %{buildroot}%{_bindir}/apb-wrapper

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

%files container-scripts
%{_bindir}/apb-wrapper

%if 0%{?with_devel}
%files devel -f devel.file-list
%license LICENSE
%dir %{gopath}/src/%{provider}.%{provider_tld}/%{project}
%dir %{gopath}/src/%{import_path}
%endif

%changelog
* Tue Aug 21 2018 Dylan Murray <dymurray@redhat.com> 1.9.3-1
- Update apb_cli document (#118) (dymurray@redhat.com)
- Bug 1613720 - Make -n flag for broker namespace (#119) (dymurray@redhat.com)

* Fri Aug 17 2018 David Zager <david.j.zager@gmail.com> 1.9.2-1
- Bug 1613664 - Update shorthand flags for bundle subcommand (#117)
  (dymurray@redhat.com)
- Bug 1577769 - Add support for outputting broker catalog request in yaml, json
  (#114) (dymurray@redhat.com)
- Fix BrokerRouteName (#115) (jmontleo@redhat.com)
- Bug 1613224 - Add CLI param support for Runner field (#113)
  (dymurray@redhat.com)
- Add apb-container-scripts sub-package with wrapper (#116)
  (jmontleo@redhat.com)
- Add support for specifying tag in registry add (#110) (dymurray@redhat.com)
- Bug 1595153 - Add support for apb test that works in and out of cluster
  (#109) (dymurray@redhat.com)
- Hard code the broker prefix to osb (#112) (jmontleo@redhat.com)
- add downstream patch (#111) (jmontleo@redhat.com)
- Updates for spec reconciliation changes to bootstrap endpoint (#107)
  (dymurray@redhat.com)
- Fix loading of defaults for config package & add tests (#105)
  (dymurray@redhat.com)
- Fix quay default (#104) (dymurray@redhat.com)
- Add unit tests for config and util packages (#97) (dymurray@redhat.com)
- Add quay registry adapter (#101) (dymurray@redhat.com)
- Change default imagePullPolicy to always (#100) (dymurray@redhat.com)
- Update vendor to pick up bundle-lib 0.2.6 (#96) (dymurray@redhat.com)
- Update repo with docs folder (#93) (dymurray@redhat.com)
- Add developers.md, include sections on building and running APBs (#94)
  (derekwhatley@gmail.com)
- Add `apb config` cmd for configuring broker/catalog interaction vars (#88)
  (derekwhatley@gmail.com)
- Update deprovision command with --skip-params flag (#85)
  (dymurray@redhat.com)
- Remove go compiler macro (#87) (jmontleo@redhat.com)
- Pre-release tweaks (#84) (derekwhatley@gmail.com)
- Update vendor to pick up bundle-lib release (#83) (dymurray@redhat.com)
- Add support for `apb catalog relist` (#79) (dymurray@redhat.com)
- Match old apb commands (#78) (derekwhatley@gmail.com)
- Add apb version command with tito versioning (#80) (dymurray@redhat.com)
- Make registry subcommand experience better (#71) (dymurray@redhat.com)
- Add support for viewing bundle logs for action. (#76)
  (derekwhatley@gmail.com)
- Remove requirement to supply namespace (#72) (dymurray@redhat.com)
- Add releasers info for tito and brew (#74) (dymurray@redhat.com)

* Thu Jul 05 2018 Dylan Murray <dymurray@redhat.com> 1.9.1-1
- new package built with tito

