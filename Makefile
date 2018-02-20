BINARY_PRE=*.go bindata.go css/main.css node_modules/chart.js/dist/Chart.min.js
BINARY_SOURCE=*.go

BINDATA_DATA=css/main.css sql/* templates/* node_modules/chart.js/dist/Chart.min.js
BINDATA_FLAGS=
BINDATA_FLAGS_DEBUG=-debug

SASS_PRE=sass/*.scss
SASS_SOURCE=sass/main.scss
SASS_FLAGS=-t compressed
SASS_FLAGS_DEBUG=-t nested -l


build: bin/plexStats

run: bin/plexStats
	bin/plexStats

debug: SASS_FLAGS=$(SASS_FLAGS_DEBUG)
debug: BINDATA_FLAGS=$(BINDATA_FLAGS_DEBUG)
debug: build


clean:
	rm -Rf ./bin ./css ./bindata.go ./node_modules


node_modules/chart.js/dist/Chart.min.js: package.json
	npm install
	touch node_modules/chart.js/dist/Chart.min.js


debug-css: SASS_FLAGS=$(SASS_FLAGS_DEBUG)
debug-css: css/main.css

css/main.css: $(SASS_PRE)
	mkdir -p css
	sassc $(SASS_FLAGS) $(SASS_SOURCE) $@


debug-bindata: BINDATA_FLAGS=$(BINDATA_FLAGS_DEBUG)
debug-bindata: bindata.go

bindata.go: $(BINDATA_DATA)
	go-bindata $(BINDATA_FLAGS) -o $@ $(BINDATA_DATA)


bin/plexStats: $(BINARY_PRE)
	mkdir -p bin
	go build -o bin/plexStats $(BINARY_SOURCE)
