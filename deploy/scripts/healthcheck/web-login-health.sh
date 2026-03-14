#!/bin/sh
set -eu

node -e "fetch('http://127.0.0.1:3000/ui/v2/login/login').then(r=>process.exit(r.ok?0:1)).catch(()=>process.exit(1))"
