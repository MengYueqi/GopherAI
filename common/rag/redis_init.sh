# docker 拉取支持向量搜索的 Redis 镜像并运行在 6381 端口
docker run -d \
  --name redis-vector-6381 \
  -p 6381:6379 \
  -p 8002:8001 \
  redis/redis-stack:latest

# 创建向量索引
FT.CREATE idx:rag_data ON HASH PREFIX 1 rag:data: SCHEMA content TEXT embedding VECTOR HNSW 6 TYPE FLOAT32 DIM 768 DISTANCE_METRIC COSINE