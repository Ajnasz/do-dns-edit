#!/bin/sh

if [ -z "$DO_TOKEN" ];then
	>&2 echo "DO_TOKEN environment variable required"
	exit 1
fi

BASEDIR=$(dirname $0)

$BASEDIR/../do-dns-edit \
	-domain="$CERTBOT_DOMAIN" \
	-token="$DO_TOKEN" \
	-recordType="TXT" \
	-recordName="_acme-challenge" \
	-recordData="$CERTBOT_VALIDATION" \
	-delete="true" \

