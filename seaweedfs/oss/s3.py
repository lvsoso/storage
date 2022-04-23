import io
import os
import boto3
from botocore.exceptions import ClientError
import logging

oss_endpoint_url = os.getenv('ENDPOINT_URL', 'http://localhost')
oss_access_key = os.getenv('ACCESS_KEY', '')
oss_secret_key = os.getenv('SECRET_KEY', '')


s3_client = boto3.client(
    's3',
    aws_access_key_id=oss_access_key,
    aws_secret_access_key=oss_secret_key,
    endpoint_url=oss_endpoint_url
)

def create_bucket(bucket_name):
    try:
        s3_client.create_bucket(Bucket=bucket_name)
    except ClientError as e:
        logging.error(e.response)
        raise e from e

def upload_file(file_name, bucket, object_name):
    try:
        _ = s3_client.upload_file(file_name, bucket, object_name)
    except ClientError as e:
        logging.error(e)
        raise e from e


def upload_fileobj(data_bytes, bucket, object_name):
    try:
        data = io.ByteIO(data_bytes)
        _ = s3_client.upload_fileobj(data, bucket, object_name)
    except ClientError as e:
        logging.error(e)
        raise e from e
