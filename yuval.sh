go build -ldflags '-w -extldflags -static -X github.com/orbs-network/boyarin/version.SemanticVersion=v1.11.2 -X github.com/orbs-network/boyarin/version.CommitVersion=4e5b04d55b26df368ee4d02aabe0bfc1e2a22219' -tags ' usergo netgo' -o _bin/boyar-v1.11.2.bin -a ./boyar/main/main.gogot

# run existing
./main --keys /opt/orbs/keys.json --max-reload-time-delay 0m --bootstrap-reset-timeout 30m --status /var/efs/boyar-status/status.json --management-config ~/boyar/opt/orbs/management-config.json --auto-update --shutdown-after-update

# _bin
./boyar-v1.12.1.bin --keys /opt/orbs/keys.json --max-reload-time-delay 0m --bootstrap-reset-timeout 30m --status /var/efs/boyar-status/status.json --management-config ~/boyar/opt/orbs/management-config.json --auto-update --shutdown-after-update

# run boyar on validator
/usr/bin/boyar --keys /opt/orbs/keys.json --max-reload-time-delay 0m --bootstrap-reset-timeout 30m --status /var/efs/boyar-status/status.json    --management-config /opt/orbs/management-config.json --auto-update --shutdown-after-update &

 scp -i ../.ssh/id_rsa_guardians -o StrictHostKeyChecking=no ../../boyarin/_bin/boyar-v1.12.1.bin ubuntu@0xDEV:/usr/bin


https://github.com/orbs-network/mainnet-deployment/tree/main/boyar_agent/node/0x9f0988Cd37f14dfe95d44cf21f9987526d6147Ba/main.sh 
https://deployment.orbs.network/boyar_agent/node/0x9f0988Cd37f14dfe95d44cf21f9987526d6147Ba/main.sh 


### cioy local boyar to 0xdev
scp -i ../.ssh/id_rsa_guardians -o StrictHostKeyChecking=no ../../boyarin/_bin/boyar-v1.12.1.bin ubuntu@0xDEV:"~"
cd /usr/bin

cp /home/ubuntu/boyar-v1.12.1.bin .
supervisorctl stop boyar
rm boyar
mv boyar-v1.12.1.bin boyar

sudo mv current current.bak2

supervisorctl start boyar




 
 #okex
iptables -A INPUT -i eth0 -p tcp --dport 80 -j ACCEPT

iptables -A INPUT -i eth0 -p tcp --dport 8080 -j ACCEPT

iptables -A PREROUTING -t nat -i eth0 -p tcp --dport 8080 -j REDIRECT --to-port 80
iptables -A PREROUTING -t nat -i eth0 -p tcp --dport 80 -j REDIRECT --to-port 8080

#list
iptables -t nat -v -L -n --line-number
iptables -t nat -v -L PREROUTING -n --line-number

#darko
iptables -t nat -A PREROUTING -p tcp -i eth0 --dport 5000 -j DNAT --to-destination 127.0.0.1:80
iptables -A FORWARD -p tcp -d 127.0.0.1 --dport 80 -m state --state NEW,ESTABLISHED,RELATED -j ACCEPT

#delete 
iptables -t nat -D PREROUTING 

# 2
socat TCP-LISTEN:8080,fork TCP:127.0.0.1:80


# Add entries in ip table
iptables -A INPUT -i eth0 -p tcp --dport 80 -j ACCEPT
iptables -A INPUT -i eth0 -p tcp --dport 18888 -j ACCEPT
# pre rout 
iptables -A PREROUTING -t nat -i eth0 -p tcp --dport 18888 -j REDIRECT --to-port 80
# if not installed
apt-get socat 
# for listen on new port
socat TCP-LISTEN:18888,fork TCP:127.0.0.1:80