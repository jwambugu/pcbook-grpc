rm *.pem

# 1. Generate CA's private key and self-signed certificate
# -nodes leaves the private key unencrypted
openssl req -x509 -newkey rsa:4096 -nodes -days 365 -keyout ca-key.pem -out ca-cert.pem -subj "/C=KE/ST=Nairobi/L=Nairobi/O=JWambugu/OU=Software Development/CN=*.jwambugu.dev/emailAddress=hi@jwambugu.dev"

echo "[*] CA's self-signed certificate:"
openssl x509 -in ca-cert.pem -noout -text

# 2. Generate web server's private key and certificate signing request (CSR)
openssl req -newkey rsa:4096 -nodes -keyout server-key.pem -out server-req.pem -subj "/C=KE/ST=Nairobi/L=Nairobi/O=PC Book/OU=Software Development/CN=*.pcbook.dev/emailAddress=hi@pcbook.dev"

# 3. Use CA's private key to sign web server's CSR and get back the signed certificate
echo "[*] Server's self-signed certificate:"
openssl x509 -req -in server-req.pem -days 60 -CA ca-cert.pem -CAkey ca-key.pem -CAcreateserial -out server-cert.pem -extfile server-ext.cnf

echo "[*] Verify server's certificate:"
openssl verify -CAfile ca-cert.pem server-cert.pem
