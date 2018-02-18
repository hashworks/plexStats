build: bin/plexStats

run: bin/plexStats
	bin/plexStats

node_modules/maketag: package.json
	npm install
	touch node_modules/maketag

css/main.css: sass/*.scss
	mkdir -p css
	sassc -t compressed sass/main.scss > $@

bin/plexStats: *.go css/main.css node_modules/maketag
	mkdir -p bin
	go build -o bin/plexStats *.go
