#!/bin/env python

import urllib
import json
import time
import unittest

def get_metrics(url):
    return json.loads(urllib.urlopen(url).read())

class TestSuite(unittest.TestCase):
    def testE2E(self):
        for i in range(20):
            try:
                time.sleep(0.5)
                metrics = get_metrics('http://localhost:8080/metrics')

                blockHeight = metrics['BlockStorage.BlockHeight']['Value']
                print 'network block height', blockHeight

                if blockHeight > 0:
                    break
            except:
                pass

        assert(blockHeight != 0)


if __name__ == '__main__':
    unittest.main()
