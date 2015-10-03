bindata.go: js/bundle.js
	go-bindata -o bindata.go -pkg promise -ignore 'node_modules|index|webpack|package' js

js/bundle.js: js/index.js js/package.json
	cd js && npm install && npm run bundle

clean:
	rm -f bindata.go js/bundle.js
