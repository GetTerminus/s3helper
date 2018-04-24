Testing

Create a large number of objects in s3
```
dd if=/dev/urandom of=bigfile bs=1024 count=5000
split -b 1024 -a 3 bigfile
aws s3 sync . s3://<bucket> --exclude bigfile
```

Delete them
```
s3helper empty-bucket -b <bucket>
```
