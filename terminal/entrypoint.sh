#!/bin/bash

# Create AWS credentials file
mkdir -p /root/.aws
echo "[default]
aws_access_key_id=AKIAIOSFOFAKEPROD
aws_secret_access_key=wJalrXUtnFEMI/K7MDENG/bPxRfakePRODKEY" > /root/.aws/credentials
chmod 600 /root/.aws/credentials

# Create .ssh directory and test key file, and set appropriate permissions
mkdir -p /root/.ssh
echo "Q29uZ3JhdHVsYXRpb25zIG9uIG5hdmlnYXRpbmcgdGhyb3VnaCB0aGUgbWF6ZSBvZiB0aGUgQm9yaW5nIFBhcGVyIENvbXBhbnkhCkhlcmUncyB5b3VyIHJld2FyZCBmb3IgYWxsIHRoZSB\
oYXJkIHdvcmtzOiBGTEFHentQNHBfVHJhaWxfRGlzY292ZXJlZCF9CkxldCB1cyBzaG93IHlvdSBhcm91bmQgdGhlIG9mZmljZS4=" > /root/.ssh/test_key_file
chmod 700 /root/.ssh
chmod 600 /root/.ssh/test_key_file

# Create the fake malware script and setup cron job
echo '#!/bin/bash
echo "Simulated Malicious Activity Detected! Access granted to sensitive data." > /dev/shm/malware_output.txt' > /dev/shm/fake_malware.sh
chmod +x /dev/shm/fake_malware.sh
echo '0 3 * * * root /dev/shm/fake_malware.sh' > /etc/cron.d/fake-malware-cron
chmod 0644 /etc/cron.d/fake-malware-cron
crontab /etc/cron.d/fake-malware-cron

# Start the cron service
cron

# Script to execute Stratum protocol attack
echo '#!/bin/bash
curl -X POST -F "username=trend" stratum+tcp://stratum.fakepooltrend.trend:3333' > /dev/shm/stratum_attack.sh
chmod +x /dev/shm/stratum_attack.sh
/dev/shm/stratum_attack.sh

# Execute the main Go application
exec ./main "$@"

