#!/bin/bash

# Usage: sudo bash setup_nginx_ssl.sh <your-email>

# Check if script is run with sudo
if [[ $EUID -ne 0 ]]; then
   echo "This script must be run with sudo or as root" 
   exit 1
fis

# Check if email argument is provided
if [ $# -eq 0 ]; then
    echo "Please provide an email address for Let's Encrypt registration"
    echo "Usage: sudo bash $0 your-email@example.com"
    exit 1
fi

# Email for Let's Encrypt registration
LETSENCRYPT_EMAIL=$1

# Update system packages
echo "Updating system packages..."
apt update

# Install required packages
echo "Installing Nginx and Certbot..."
apt install -y nginx certbot python3-certbot-nginx

# Create Nginx configuration file
echo "Creating Nginx configuration..."
cat > /etc/nginx/sites-available/dicedb <<EOL
server {
    listen 80;
    listen [::]:80;
    server_name playground-mono.dicedb.io;
    return 301 https://\$server_name\$request_uri;
}

server {
    listen 443 ssl http2;
    listen [::]:443 ssl http2;
    server_name playground-mono.dicedb.io;

    ssl_certificate /etc/letsencrypt/live/playground-mono.dicedb.io/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/playground-mono.dicedb.io/privkey.pem;
    ssl_session_timeout 1d;
    ssl_session_cache shared:SSL:50m;
    ssl_session_tickets off;

    # Modern configuration
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384:ECDHE-ECDSA-CHACHA20-POLY1305:ECDHE-RSA-CHACHA20-POLY1305:DHE-RSA-AES128-GCM-SHA256:DHE-RSA-AES256-GCM-SHA384;
    ssl_prefer_server_ciphers off;

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
certbot --nginx -d 'playground-mono.dicedb.io' --email "$LETSENCRYPT_EMAIL" --agree-tos

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
echo "1. Ensure your DNS is correctly configured for playground-mono.dicedb.io"
echo "2. Verify SSL configuration at https://www.ssllabs.com/ssltest/"
