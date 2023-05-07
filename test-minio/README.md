运行 minio

```shell
docker run --rm --name mi -p 9000:9000 -p 9001:9001 quay.io/minio/minio server /data --console-address ":9001"
```

上传文件后，生成签名链接

```shell
go run main.go test api
http://127.0.0.1:9000/test/api?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=minioadmin%2F20230507%2Fus-east-1%2Fs3%2Faws4_request&X-Amz-Date=20230507T185529Z&X-Amz-Expires=86400&X-Amz-SignedHeaders=host&X-Amz-Signature=4632c1c6c0feeba7459f36331f457f7f0cbd6e02eb249d0dceb18963e2730c0a
```

分段下载

```shell
curl -s -H "Range: bytes=0-1000" -o /tmp/1 'http://127.0.0.1:9000/test/api?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=minioadmin%2F20230507%2Fus-east-1%2Fs3%2Faws4_request&X-Amz-Date=20230507T185529Z&X-Amz-Expires=86400&X-Amz-SignedHeaders=host&X-Amz-Signature=4632c1c6c0feeba7459f36331f457f7f0cbd6e02eb249d0dceb18963e2730c0a'

curl -s -H "Range: bytes=1001-1744" -o /tmp/2 'http://127.0.0.1:9000/test/api?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=minioadmin%2F20230507%2Fus-east-1%2Fs3%2Faws4_request&X-Amz-Date=20230507T185529Z&X-Amz-Expires=86400&X-Amz-SignedHeaders=host&X-Amz-Signature=4632c1c6c0feeba7459f36331f457f7f0cbd6e02eb249d0dceb18963e2730c0a'
```

检查

```shell
(base) lv@lv:split$ sha256sum /tmp/3
9c6f73b2cdb31b867d52512f966a15a4175a055e9b509208103d0115e2e221a4  /tmp/3
(base) lv@lv:split$ sha256sum ~/Desktop/api 
9c6f73b2cdb31b867d52512f966a15a4175a055e9b509208103d0115e2e221a4  /home/lv/Desktop/api
```