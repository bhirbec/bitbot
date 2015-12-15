SHELL=/bin/bash

.PHONY: all js

all: js

js: public/build/main.js

public/build/main.js: $(shell find client -name "*.js" -o -name "*.jsx")
	mkdir -p ./public/build
	./node_modules/.bin/browserify -t reactify client/main.js > $@
