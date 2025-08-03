apt-get install -y certbot python3-certbot-dns-digitalocean

echo 'dns_digitalocean_token=%[3]s' > ~/certbot-creds.ini
chmod 600 ~/certbot-creds.ini

certbot certonly \
  --dns-digitalocean --dns-digitalocean-credentials ~/certbot-creds.ini \
  -d %[1]s --non-interactive --expand \
  --agree-tos --email %[2]s

exit_code=$?

if [ $exit_code -eq 0 ]; then
    cp /etc/letsencrypt/live/%[1]s/fullchain.pem /root/fullchain.pem
    cp /etc/letsencrypt/live/%[1]s/privkey.pem /root/privkey.pem

    docker cp -L /root/fullchain.pem "sky-app:/root/fullchain.pem"
    docker cp -L /root/privkey.pem "sky-app:/root/privkey.pem"
else
    exit $exit_code
fi