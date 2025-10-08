#!/bin/bash

# 获取时间戳
timestamp=$(date +"%Y%m%d_%H%M%S")

# 创建保存目录
mkdir -p ./pprof/$timestamp

echo "📦 保存路径：./pprof/$timestamp"

# 采样 CPU（10秒）
echo "🧠 采样 CPU..."
curl -o ./pprof/$timestamp/profile.out http://localhost:6060/debug/pprof/profile?seconds=20

# 采样 heap
echo "📦 采样 Heap..."
curl -o ./pprof/$timestamp/heap.out http://localhost:6060/debug/pprof/heap

# 采样 trace（5秒）
echo "⏱️ 采样 Trace..."
curl -o ./pprof/$timestamp/trace.out http://localhost:6060/debug/pprof/trace?seconds=20



# 生成 SVG 火焰图
echo "🔥 生成火焰图 SVG..."
go tool pprof -svg ./pprof/$timestamp/profile.out > ./pprof/$timestamp/flamegraph.svg

# 启动 Web UI
echo "🌐 启动 Web UI..."
go tool pprof -http=:10001 ./pprof/$timestamp/profile.out

echo "✅ 所有采样完成！火焰图已生成：./pprof/$timestamp/flamegraph.svg"
echo "🌐 在浏览器访问：http://localhost:10001/ui/"
