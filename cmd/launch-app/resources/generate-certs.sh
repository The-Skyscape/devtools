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

    # Stop and remove the old container
    docker stop sky-app
    docker rm sky-app

    # Recreate container with SSL certificates
    mkdir -p /home/coder/.skyscape/services/sky-app && \
    chmod -R 777 /home/coder/.skyscape/services/sky-app && \
    docker create \
      --name sky-app \
      --network host \
      --privileged \
      --entrypoint /app \
      -v /root/.skyscape:/root/.skyscape -v /var/run/docker.sock:/var/run/docker.sock \
      -e PORT=80 -e THEME=corporate \
      -v /var/run/docker.sock:/var/run/docker.sock \
      skyscape:latest && \
    docker cp /root/app sky-app:/app && \
    docker cp /root/fullchain.pem sky-app:/root/fullchain.pem 2>/dev/null || true && \
    docker cp /root/privkey.pem sky-app:/root/privkey.pem 2>/dev/null || true && \
    docker start sky-app
else
    exit $exit_code
fi