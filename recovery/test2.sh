#!/bin/bash
if command -v curl &> /dev/null; then
    curl -XPOST -H "Content-Type: application/json" "http://logs.orbs.network:3001/putes/boyar-recovery" -d '{ "node": "0xSTAGING", "script":"disk cleanup1", "stage":"start" }'
fi

if command -v journalctl &> /dev/null; then
    journalctl --vacuum-size=200M
else
    echo "journalctl could not be found"    
fi    
    
# Removes old revisions of snaps
# CLOSE ALL SNAPS BEFORE RUNNING THIS
if command -v snap &> /dev/null; then
    set -eu
    LANG=C snap list --all | awk '/disabled/{print $1, $3}' |
        while read snapname revision; do
            snap remove "$snapname" --revision="$revision"
        done
else
    echo "snap could not be found"    
fi    

# apt-get cleanup
if command -v apt-get &> /dev/null; then
    # apt-get cleanup - sudo ommited as boyar is already sudo
    apt-get clean
    # clean apt cache 
    apt-get autoclean
    # unnecessary packages
    apt-get autoremove
    # snapd
    apt purge snapd
else
    echo "apt-get could not be found"
fi

# old kernel versions
if command -v dpkg &> /dev/null; then
    dpkg --get-selections | grep linux-image
else
    echo "apt-get could not be found"
fi

# NOT WORKINGDelete "Dead" or "Exited" containers.
# docker rm $(docker ps -a | grep "Dead\|Exited" | awk '{print $1}')
# #Delete dangling docker images.
# docker rmi -f $(docker images -qf dangling=true)
# #Delete or clean up unused docker volumes.
# docker rmi -f $(docker volume ls -qf dangling=true)

if command -v curl &> /dev/null; then
    curl -XPOST -H "Content-Type: application/json" "http://logs.orbs.network:3001/putes/boyar-recovery" -d '{ "node": "0xSTAGING", "script":"disk cleanup1", "stage":"end" }'
fi