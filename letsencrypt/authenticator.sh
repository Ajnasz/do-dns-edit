#!/bin/sh

if [ -z "$DO_TOKEN" ];then
	>&2 echo "DO_TOKEN environment variable required"
	exit 1
fi

BASEDIR=$(dirname $0)

$BASEDIR/../do-dns-edit \
	-recordType="TXT" \
	-token="$DO_TOKEN" \
	-domain="$CERTBOT_DOMAIN" \
	-recordName="_acme-challenge" \
	-recordData="$CERTBOT_VALIDATION" \
	-create="true" \
