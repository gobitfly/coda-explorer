GITCOMMIT=`git describe --always`
VERSION=$$(git describe 2>/dev/null || echo "0.0.0-${GITCOMMIT}")
GITDATE=`TZ=UTC git show -s --date=iso-strict-local --format=%cd HEAD`
BUILDDATE=`date -u +"%Y-%m-%dT%H:%M:%S%:z"`
PACKAGE=coda-explorer
LDFLAGS="-X ${PACKAGE}/version.Version=${VERSION} -X ${PACKAGE}/version.BuildDate=${BUILDDATE} -X ${PACKAGE}/version.GitCommit=${GITCOMMIT} -X ${PACKAGE}/version.GitDate=${GITDATE}"

all: explorer frontend statistics

lint:
	golint ./...

explorer:
	go build --ldflags=${LDFLAGS} -o bin/indexer cmd/indexer/main.go

statistics:
	go build --ldflags=${LDFLAGS} -o bin/statistics cmd/statistics/main.go

frontend:
	rm -rf bin/templates
	rm -rf bin/static
	rm -rf bin/ip2location
	mkdir -p bin/templates/
	mkdir -p bin/static/
	mkdir -p bin/ip2location/
	cp -r templates/ bin/
	cp -r static/ bin/
	cp -r ip2location/ bin/
	go build --ldflags=${LDFLAGS} -o bin/frontend cmd/frontend/main.go