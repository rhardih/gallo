#!/usr/bin/env sh

##
# NAME
#      hash_name_assets
#
# SYNOPSIS
#      hash_name_assets
#
# DESCRIPTION
#      Makes copies of .js and .css files in public/assets/ with a filename
#      containing the sha256sum of the content of each file.
#
# MISC
#      2020-06-29 - René Hansen

set -e

for ext in js css; do
  cd public/assets/$ext

  # Remove previously copied files
  if [ "$(echo -- *.*.$ext)" != "-- *.*.$ext" ]; then
    rm -- *.*.$ext
  fi

  sha256sum -- *.$ext > sha256sum.txt

  while IFS="  " read -r hash filename; do
   cp "$filename" "${filename%.$ext}.$hash.$ext"
  done < sha256sum.txt

  cd -
done
