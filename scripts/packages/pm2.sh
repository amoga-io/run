# Install and configure pm2
sudo npm install -g pm2
sudo -u azureuser pm2 save
sudo chmod 755 $(which pm2)
sudo chmod -R 755 $(dirname $(which pm2))/../lib/node_modules/pm2
sudo mkdir -p /var/log/pm2
sudo chmod 777 /var/log/pm2
sudo -u azureuser pm2 startup systemd