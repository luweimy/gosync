## 编译&启动
```
// 默认并发1，监听9293端口
make build && ./bin/gosync

// 或者修改默认启动参数
make build && ./bin/gosync --concurrency 2 --listen :9294
```

## 文件上传
```
// file: 要上传的文件
// dst: 上传文件到哪
// perm: 上传后文件的权限，默认777
curl -F "file=@/Users/luwei/token.md" -F "dst=/Users/luweimy/tmp" -F "perm=666" http://localhost:9293/upload
```

## 远程执行命令
```
curl -d 'ls -l' http://localhost:9293/exec
curl -d 'find . -name *.go' http://localhost:9293/exec
```

## 调整允许同时执行的命令数量(默认1)
```
curl -F "n=10" http://localhost:9293/concurrency
```