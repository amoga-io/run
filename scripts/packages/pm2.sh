# Detect the user to set up PM2
if [ "$SUDO_USER" ]; then
  TARGET_USER="$SUDO_USER"
else
  TARGET_USER="$(whoami)"
fi

sudo -u "$TARGET_USER" npm install -g pm2
sudo -u "$TARGET_USER" pm2 save
sudo chmod 755 $(which pm2)
sudo chmod -R 755 $(dirname $(which pm2))/../lib/node_modules/pm2
sudo mkdir -p /var/log/pm2
sudo chmod 777 /var/log/pm2
sudo -u "$TARGET_USER" pm2 startup systemd