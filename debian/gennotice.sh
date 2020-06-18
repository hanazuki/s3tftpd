#!/bin/bash

find vendor \( -name NOTICE -o -name NOTICE.\* \) -print0 | \
    while IFS= read -r -d $'\0' file; do
        echo "$file"
        echo "=========="
        cat "$file"
        echo
    done
