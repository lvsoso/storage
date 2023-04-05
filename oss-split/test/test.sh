
curl -v -X POST -H " application/vnd.git-lfs+json" \
 -d '{
  "operation": "upload",
  "objects": [
    {
      "oid": "c93dda5f38ad123e7244e88eca3a3037d4dddda22574f5b994a7b00b5f37bed0",
      "size": 1024000000
    }
  ],
  "transfers": ["multipart", "basic"]
}'  http://127.0.0.1:9999/oss/new_multipart


dd if=/dev/urandom of=random_file bs=1M count=40