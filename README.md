Create, update, delete dns record on digital ocean

Dry run:
```
DO_TOKEN="digital ocean token" /path/to/certbot-auto certonly --dry-run --preferred-challenges dns --manual --manual-cleanup-hook=/path/to/cleanup.sh --manual-auth-hook=/path/to/authenticator.sh -d something.example.com

