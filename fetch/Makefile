bindata.go: js/bundle.js
	go-bindata -o bindata.go -pkg fetch -ignore 'node_modules|fetch|headers|index|request|response|webpack|package' js

js/bundle.js: js/fetch.js js/headers.js js/index.js js/request.js js/response.js
	cd js && npm install && npm run bundle

clean:
	rm -f bindata.go js/bundle.js
