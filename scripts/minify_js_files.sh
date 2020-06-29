#!/usr/bin/env sh

##
# NAME
#      minify_js_files
#
# SYNOPSIS
#      minify_js_files [-w]
#
# DESCRIPTION
#      Runs minification on javascript files found in app/assets/js and outputs
#      them to public/assets/js. Optionally watches for file changes with -w.
#
# MISC
#      2020-06-29 - René Hansen

command -v inotifywait >/dev/null 2>&1 || {
  apk add -u inotify-tools
}

set -e

# watch
watch=false

usage() {
	echo "Usage: $0 [-w]" 1>&2
	exit 1
}

while getopts "w" o; do
	case "${o}" in
		w)
      watch=true
			;;
		*)
			usage
			;;
	esac
done

mkdir -p public/assets/js
npm install

if [ "$watch" = true ]; then
  inotifywait -e moved_to,create -m app/assets/js |
  while read -r dir event name; do
    case $name in *.js)
      npm run minify-js -- --output public/assets/js/${name%.js}.js -- \
        "$dir$name"
    esac
  done
else
  for filename in app/assets/js/*.js; do
    name=$(basename $filename)
    npm run minify-js -- --output public/assets/js/${name%.js}.js -- \
      $filename
  done
fi
