## Data for test files generated with

for i in 1024 2048 3072 4096 5120 6144 7168 8192 15360 ;

 do

   openssl genrsa -des3 -passout pass:Test1ng -out test${i}.key ${i}

   openssl req -new -x509 -days 3650 -key test${i}.key -out testi${i}.crt -subj "/C=US/ST=California/L=San Diego/O=LifeRaft/OU=Testing/CN=FakeCA" -passin pass:Test1ng

done