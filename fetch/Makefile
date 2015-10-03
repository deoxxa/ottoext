dist-fetch.rice-box.go: dist-fetch/bundle.js
	rice embed-go

dist-fetch/bundle.js: js/bundle.js
	mkdir -p dist-fetch
	cp -a js/bundle.js dist-fetch/bundle.js

js/bundle.js: js/fetch.js js/headers.js js/index.js js/request.js js/response.js
	cd js && npm install && npm run bundle

clean:
	rm -rf dist-fetch.rice-box.go dist-fetch js/bundle.js
