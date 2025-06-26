#!/bin/sh
set -eo pipefail
pip install -U pip2pi
pip2pi -S -z /packages/ -r /requirements.txt

rm -rf /packages/*.{gz,zip,whl}
