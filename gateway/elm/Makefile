all: build minify

build:
	# elm make --optimize src/*.elm --output ../static/js/app.js
	elm make src/*.elm --output ../static/js/app.js

minify:
	uglifyjs ../static/js/app.js --compress 'pure_funcs="F2,F3,F4,F5,F6,F7,F8,F9,A2,A3,A4,A5,A6,A7,A8,A9",pure_getters,keep_fargs=false,unsafe_comps,unsafe' | uglifyjs --mangle --output=../static/js/app.min.js
