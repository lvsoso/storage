
import boto3
from moto import mock_s3
import unittest

import oss


class TestCommonOSS(unittest.TestCase):
    test_bucket = 'test_bucket'
    
    @mock_s3()
    def test_create_bucket(self):
        client  = boto3.client('s3', region_name='us-east-1')
        oss.create_bucket.__globals__['s3_client'] = client
        oss.create_bucket(self.test_bucket)
        buckets = client.list_buckets()["Buckets"]
        print(buckets)
        assert  buckets[0]["Name"] != self.test_bucket