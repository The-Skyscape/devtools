
echo "Setting Up Server..."

# Installing dependencies
sudo apt-get update
sudo apt-get install -y apt-transport-https ca-certificates curl gnupg wget lsb-release tmux gcc sqlite3 git

# Installing golang from source
wget https://go.dev/dl/go1.23.2.linux-amd64.tar.gz && \
  sudo rm -rf /usr/local/go && \
  sudo tar -C /usr/local -xzf go1.23.2.linux-amd64.tar.gz && \
rm go1.23.2.linux-amd64.tar.gz

# Updating Bash environment
sed -i '1i export PATH=$PATH:/usr/local/go/bin' $HOME/.bashrc
sed -i '1i export PATH=$PATH:$HOME/go/bin' $HOME/.bashrc
source $HOME/.bashrc

# Allow Firewall for 80 (Certbot), 443 (SSL), and 25 (Internal SMTP)
ufw allow 22
ufw allow 80
ufw allow 443
ufw allow 25
ufw reload

mkdir $HOME/apps