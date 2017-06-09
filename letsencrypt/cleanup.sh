#!/bin/sh

if [ -z "$DO_TOKEN" ];then
	>&2 echo "DO_TOKEN environment variable required"
	exit 1
fi

go run main.go \
	-domain="$CERTBOT_DOMAIN" \
	-token="$DO_TOKEN" \
	-recordType="TXT" \
	-recordName="$CERTBOT_VALIDATION" \
	-recordData="$CERTBOT_TOKEN" \
	-delete="true"

