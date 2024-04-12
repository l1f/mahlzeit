default: build

TOOLS_AIR_VERSION := "v1.43.0"
TOOLS_DBMATE_VERSION := "v2.2.0"
TOOLS_SQLC_VERSION := "1.18.0"

# it's probably a good idea to test the versions of the tools... but that's a problem for the future Nina 
_check_requirements:
	air -v
	dbmate -v
	sqlc version

_install-deps:
	go mod download

# Build the application, it's saved under ./mahlzeit
build: _install-deps
	go generate ./web/...
	go build -o mahlzeit ./cmd/mahlzeit

tmpdir  := `mktemp -d`
version := "0.0.1"
tarfile := "mahlzeit-" + version + "-" + os() + "-" + arch()
tardir  := tmpdir / tarfile
tarball := tardir + ".tar.gz"

# Package builds the application and compresses the binary and all necessary files
# into a single .tar.gz archive.
package: build
	mkdir -p {{tardir}}
	cp README.md LICENSE.md config.toml {{tardir}}
	cp mahlzeit {{tardir}}
	mkdir -p {{tardir}}/db/migrations/
	cp -r db/migrations/* {{tardir}}/db/migrations/
	cd {{tardir}} && tar zcvf {{tarball}} .
	cp {{tarball}} {{invocation_directory()}}
	rm -rf {{tarball}} {{tardir}}

# Apply all pending database migrations.
migrate:
    docker compose up -d
    dbmate --wait up

# Installs the dependencies and applies all database migrations.
prepare: _check_requirements migrate

# Start the watch mode for local development
dev: migrate
    air
