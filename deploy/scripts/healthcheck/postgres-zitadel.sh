#!/bin/sh
set -eu

pg_isready -d zitadel -U postgres
