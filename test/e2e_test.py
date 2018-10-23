#!/bin/env python

import urllib
import json
import time
import sys

def get_metrics(url):
    return json.loads(urllib.urlopen(url).read())

def TestE2E():
    for i in range(20):
        try:
            time.sleep(0.5)
            metrics = get_metrics('http://localhost:8080/metrics')

            blockHeight = metrics['BlockStorage.BlockHeight']['Value']
            print 'network block height', blockHeight

            if blockHeight >= 3:
                print 'Pass'
                return
        except:
            pass

    sys.exit(1)


if __name__ == '__main__':
    TestE2E()
