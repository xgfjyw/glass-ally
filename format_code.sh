#!/bin/sh

find . -name "*.py" | grep -v venv | xargs black -l 120
find . -name "*.py" | grep -v venv | xargs isort -sl -fss -w 120
