# moth-index 容器验证

基于 Go 实现的夜行昆虫诱捕监测 Web 服务项目，一款面向生态调查的后端服务，处理调查站点、诱捕器、夜间批次、环境读数与样本复核的持久化协同。

镜像使用 `golang:1.22-bookworm`，构建时固定 `GOTOOLCHAIN=local`。启动服务前通过 `MOTH_INDEX_DB` 指定 SQLite 文件路径，默认监听 `:8080`。

```bash
./build_benzhi_docker.sh moth-index-run linux/amd64
docker run --rm --platform linux/amd64 -p 8080:8080 -e MOTH_INDEX_DB=/tmp/moth-index.db benzhi/moth-index-run:latest
```

这是 Go HTTP 服务的运行入口，页面只提供夜间调查操作入口，数据和状态由后端持久化。
