#!/bin/bash

# 1. Generate CA's private key and self-signed certificate
openssl req -x509 -nodes -sha256 -days 365 -newkey rsa:2048 -keyout cert/ca-key.pem -out cert/ca-cert.pem  -subj "/C=BR/ST=Minas Gerais/L=Divinopolis/O=TCM Software Certificate Authority/CN=*.tcmsoftware.com/emailAddress=tcmsoftwareltda@gmail.com"

# 2. Generate server's private key and certificate signing request (CSR)
openssl req -new -sha256 -keyout cert/server-key.pem -nodes -out cert/server-req.pem -subj "/C=BR/ST=Minas Gerais/L=Divinopolis/O=Book Management Service/CN=*.tcmsoftware.bookservice.com/emailAddress=bookservice@tcmsoftware.com"

# 3. Use CA's private key to sign server's CSR and get back the signed certificate
openssl x509 -req -in cert/server-req.pem -sha256 -days 60 -CA cert/ca-cert.pem -CAkey cert/ca-key.pem -CAcreateserial -out cert/server-cert.pem -extfile cert/config/server-ext.cnf

# 4. Generate client's private key and certificate signing request (CSR)
openssl req -new -sha256 -keyout cert/client-key.pem  -nodes -out cert/client-req.pem -subj "/C=BR/ST=Minas Gerais/L=Divinopolis/O=Book Management Service client/CN=*.tcmsoftware.bookservice.client.com/emailAddress=client.bookservice@tcmsoftware.com"

# 5. Use CA's private key to sign client's CSR and get back the signed certificate
openssl x509 -req -in cert/client-req.pem -sha256 -days 60 -CA cert/ca-cert.pem -CAkey cert/ca-key.pem -CAcreateserial -out cert/client-cert.pem -extfile cert/config/client-ext.cnf

echo "Finished."