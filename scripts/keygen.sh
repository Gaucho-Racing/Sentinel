#!/bin/bash

ssh-keygen -t rsa -b 4096 -m PEM -f jwtRS256.key
openssl rsa -in jwtRS256.key -pubout -outform PEM -out jwtRS256.key.pub

echo "RSA_PUBLIC_KEY=\"$(awk 'NF {sub(/\r/, ""); printf "%s\\n",$0;}' jwtRS256.key.pub)\""
echo "RSA_PRIVATE_KEY=\"$(awk 'NF {sub(/\r/, ""); printf "%s\\n",$0;}' jwtRS256.key)\""

# Clean up temporary files
rm jwtRS256.key jwtRS256.key.pub
