Summary:        PostgreSQL database engine
Name:           postgresql96
Version:        9.6.21
Release:        1%{?dist}
License:        PostgreSQL
URL:            www.postgresql.org
Group:          Applications/Databases
Vendor:         VMware, Inc.
Distribution:   Photon

Source0:        http://ftp.postgresql.org/pub/source/v%{version}/%{name}-%{version}.tar.bz2
%define sha1    postgresql=e24333824d361968958613f546ae06011d9d1dfc

# Customized location of pg96
%global pgbaseinstdir   /usr/local/pg96

# Common libraries needed
BuildRequires:  krb5-devel
BuildRequires:  libxml2-devel
BuildRequires:  openldap
BuildRequires:  perl
BuildRequires:  readline-devel
BuildRequires:  openssl-devel
BuildRequires:  zlib-devel
BuildRequires:  tzdata
BuildRequires:  bzip2
BuildRequires:  sudo
Requires:       krb5
Requires:       libxml2
Requires:       openldap
Requires:       openssl
Requires:       readline
Requires:       zlib
Requires:       tzdata
Requires:       bzip2
Requires:       sudo

Requires:   %{name}-libs = %{version}-%{release}

%description
PostgreSQL is an object-relational database management system.

%package libs
Summary:    Libraries for use with PostgreSQL
Group:      Applications/Databases

%description libs
The postgresql-libs package provides the essential shared libraries for any
PostgreSQL client program or interface. You will need to install this package
to use any other PostgreSQL package or any clients that need to connect to a
PostgreSQL server.

%package        devel
Summary:        Development files for postgresql.
Group:          Development/Libraries
Requires:       postgresql = %{version}-%{release}

%description    devel
The postgresql-devel package contains libraries and header files for
developing applications that use postgresql.

%prep
%setup -q

%build
ls -la
sed -i '/DEFAULT_PGSOCKET_DIR/s@/tmp@/run/postgresql@' src/include/pg_config_manual.h &&
./configure \
    --prefix=%{pgbaseinstdir} \
    --with-includes=%{pgbaseinstdir}/include \
    --with-libraries=%{pgbaseinstdir}/lib \
    --datarootdir=%{pgbaseinstdir}/share \
    --enable-thread-safety \
    --with-ldap \
    --with-libxml \
    --with-openssl \
    --with-gssapi \
    --with-readline \
    --with-system-tzdata=%{_datadir}/zoneinfo \
    --docdir=%{pgbaseinstdir}/doc/postgresql
make %{?_smp_mflags}
cd contrib && make %{?_smp_mflags}

%install
[ %{buildroot} != "/"] && rm -rf %{buildroot}/*
make install DESTDIR=%{buildroot}
cd contrib && make install DESTDIR=%{buildroot}

%{_fixperms} %{buildroot}/*

%check
sed -i '2219s/",/  ; EXIT_STATUS=$? ; sleep 5 ; exit $EXIT_STATUS",/g'  src/test/regress/pg_regress.c
chown -Rv nobody .
sudo -u nobody -s /bin/bash -c "PATH=$PATH make -k check"

%post   -p /sbin/ldconfig
%postun -p /sbin/ldconfig
%clean
rm -rf %{buildroot}/*

%files
%defattr(-,root,root)
%{pgbaseinstdir}/bin/initdb
%{pgbaseinstdir}/bin/oid2name
%{pgbaseinstdir}/bin/pg_archivecleanup
%{pgbaseinstdir}/bin/pg_basebackup
%{pgbaseinstdir}/bin/pg_controldata
%{pgbaseinstdir}/bin/pg_ctl
%{pgbaseinstdir}/bin/pg_receivexlog
%{pgbaseinstdir}/bin/pg_recvlogical
%{pgbaseinstdir}/bin/pg_resetxlog
%{pgbaseinstdir}/bin/pg_rewind
%{pgbaseinstdir}/bin/pg_standby
%{pgbaseinstdir}/bin/pg_test_fsync
%{pgbaseinstdir}/bin/pg_test_timing
%{pgbaseinstdir}/bin/pg_upgrade
%{pgbaseinstdir}/bin/pg_xlogdump
%{pgbaseinstdir}/bin/pgbench
%{pgbaseinstdir}/bin/postgres
%{pgbaseinstdir}/bin/postmaster
%{pgbaseinstdir}/bin/vacuumlo
%{pgbaseinstdir}/share/postgresql/*
%{pgbaseinstdir}/lib/postgresql/*
%{pgbaseinstdir}/doc/postgresql/extension/*.example
%exclude %{pgbaseinstdir}/share/postgresql/pg_service.conf.sample
%exclude %{pgbaseinstdir}/share/postgresql/psqlrc.sample

%files libs
%{pgbaseinstdir}/bin/clusterdb
%{pgbaseinstdir}/bin/createdb
%{pgbaseinstdir}/bin/createlang
%{pgbaseinstdir}/bin/createuser
%{pgbaseinstdir}/bin/dropdb
%{pgbaseinstdir}/bin/droplang
%{pgbaseinstdir}/bin/dropuser
%{pgbaseinstdir}/bin/ecpg
%{pgbaseinstdir}/bin/pg_config
%{pgbaseinstdir}/bin/pg_dump
%{pgbaseinstdir}/bin/pg_dumpall
%{pgbaseinstdir}/bin/pg_isready
%{pgbaseinstdir}/bin/pg_restore
%{pgbaseinstdir}/bin/psql
%{pgbaseinstdir}/bin/reindexdb
%{pgbaseinstdir}/bin/vacuumdb
%{pgbaseinstdir}/lib/libecpg*.so.*
%{pgbaseinstdir}/lib/libpgtypes*.so.*
%{pgbaseinstdir}/lib/libpq*.so.*
%{pgbaseinstdir}/share/postgresql/pg_service.conf.sample
%{pgbaseinstdir}/share/postgresql/psqlrc.sample

%files devel
%defattr(-,root,root)
%{pgbaseinstdir}/include/*
%{pgbaseinstdir}/lib/pkgconfig/*
%{pgbaseinstdir}/lib/libecpg*.so
%{pgbaseinstdir}/lib/libpgtypes*.so
%{pgbaseinstdir}/lib/libpq*.so
%{pgbaseinstdir}/lib/libpgcommon.a
%{pgbaseinstdir}/lib/libpgfeutils.a
%{pgbaseinstdir}/lib/libpgport.a
%{pgbaseinstdir}/lib/libpq.a
%{pgbaseinstdir}/lib/libecpg.a
%{pgbaseinstdir}/lib/libecpg_compat.a
%{pgbaseinstdir}/lib/libpgtypes.a

%changelog
*   Yan Wang <wangyan@vmware.com>
-   Customize postgres 96 from original spec
