#!/bin/bash

# Nginx and SSL Setup Script for dicedb.io
# Usage: sudo bash setup_nginx_ssl.sh <your-email>

# Check if script is run with sudo
if [[ $EUID -ne 0 ]]; then
   echo "This script must be run with sudo or as root" 
   exit 1
fi

# Email for Let's Encrypt registration
LETSENCRYPT_EMAIL=arpit@dicedb.io

# Update system packages
echo "Updating system packages..."
apt update && apt upgrade -y

# Install required packages
echo "Installing Nginx and Certbot..."
apt install -y nginx certbot python3-certbot-nginx

# Create Nginx configuration file
echo "Creating Nginx configuration..."
cat > /etc/nginx/sites-available/dicedb <<EOL
server {
    listen 80;
    server_name *.dicedb.io dicedb.io;

    location / {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host \$host;
        proxy_cache_bypass \$http_upgrade;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
    }
}
EOL

# Enable the site configuration
echo "Enabling Nginx site configuration..."
ln -sf /etc/nginx/sites-available/dicedb /etc/nginx/sites-enabled/

# Test Nginx configuration
echo "Testing Nginx configuration..."
nginx -t

# Restart Nginx to apply changes
echo "Restarting Nginx..."
systemctl restart nginx

# Obtain SSL certificate
echo "Obtaining SSL certificate..."
certbot --nginx -d dicedb.io -d '*.dicedb.io' --email "$LETSENCRYPT_EMAIL" --agree-tos --no-ew

# Enable automatic renewal
echo "Setting up automatic SSL certificate renewal..."
systemctl enable certbot.timer
systemctl start certbot.timer

# Configure UFW firewall
echo "Configuring UFW firewall..."
ufw allow 'Nginx Full'

# Verify SSL renewal
echo "Performing dry run of SSL certificate renewal..."
certbot renew --dry-run

# Final status checks
echo "Checking Nginx status..."
systemctl status nginx

echo "Setup complete!"
echo "Don't forget to:"
echo "1. Ensure your DNS is correctly configured for *.dicedb.io"
echo "2. Verify SSL configuration at https://www.ssllabs.com/ssltest/"
