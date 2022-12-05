#!/bin/sh
#Set environmet variable
echo VUE_APP_ROOT_API=$URL > .env
#Build js packages
npm run build
#Run web server
http-server dist
