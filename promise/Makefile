dist-promise.rice-box.go: dist-promise/bundle.js
	rice embed-go

dist-promise/bundle.js: js/bundle.js
	mkdir -p dist-promise
	cp -a js/bundle.js dist-promise/bundle.js

js/bundle.js: js/index.js js/package.json
	cd js && npm install && npm run bundle

clean:
	rm -rf dist-promise.rice-box.go dist-promise js/bundle.js
