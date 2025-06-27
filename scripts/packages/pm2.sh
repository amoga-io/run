# Detect the user to set up PM2
if [ "$SUDO_USER" ]; then
  TARGET_USER="$SUDO_USER"
else
  TARGET_USER="$(whoami)"
fi
TARGET_HOME="$(eval echo "~$TARGET_USER")"

DEFAULT_VERSION="latest"
version="$DEFAULT_VERSION"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --version)
      shift
      version="$1"
      shift
      ;;
    *)
      shift
      ;;
  esac
done

export PATH="$TARGET_HOME/.npm-global/bin:$PATH"

if [ "$version" = "latest" ]; then
  sudo -u "$TARGET_USER" npm install -g pm2
else
  sudo -u "$TARGET_USER" npm install -g pm2@"$version"
fi

# Source the user's profile to update PATH and check pm2
if ! sudo -u "$TARGET_USER" bash -c "source $TARGET_HOME/.profile && which pm2" > /dev/null; then
  echo "pm2 not found in PATH for $TARGET_USER after install"
  exit 1
fi

sudo -u "$TARGET_USER" bash -c "source $TARGET_HOME/.profile && pm2 save"
sudo chmod 755 "$(sudo -u "$TARGET_USER" bash -c 'source $TARGET_HOME/.profile && which pm2')"
sudo chmod -R 755 "$(dirname "$(sudo -u "$TARGET_USER" bash -c 'source $TARGET_HOME/.profile && which pm2')")/../lib/node_modules/pm2"
sudo mkdir -p /var/log/pm2
sudo chmod 777 /var/log/pm2
sudo -u "$TARGET_USER" bash -c "source $TARGET_HOME/.profile && pm2 startup systemd"