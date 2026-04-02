# pRESTd

_p_**REST** (**P**_ostgreSQL_ **REST**), is a simple production-ready API, that delivers an instant, realtime, and high-performance application on top of your **existing or new Postgres** database.

> PostgreSQL version 9.5 or higher
> Opengauss
> Gaussdb version 506

## Problems we solve

The pREST project is the API that addresses the need for fast and efficient solution in building RESTful APIs on PostgreSQL databases. It simplifies API development by offering:

1. A **lightweight server** with easy configuration;
2. Direct **SQL queries with templating** in customizable URLs;
3. Optimizations for **high performance**;
4. **Enhanced** developer **productivity**;
5. **Authentication and authorization** features;
6. **Pluggable** custom routes and middlewares.

Overall, pREST simplifies the process of creating secure and performant RESTful APIs on top of your new or old PostgreSQL database.

## Why we built pREST

When we built pREST, we originally intended to contribute and build with the PostgREST project, although it took a lot of work as the project is in Haskell. At the time, we did not have anything similar or intended to keep working with that tech stack. We've been building production-ready Go applications for a long time, so building a similar project with Golang as its core was natural.

Additionally, as Go has taken a huge role in many other vital projects such as Kubernetes and Docker, and we've been able to use the pREST project in many different companies with success over the years, it has shown to be an excellent decision.

## How to deploy a gaussdb restful service

### Run a gaussdb in docker
- single node:
- centralize 3 nodes:
- distribution cluster:

### Login into gaussdb and create database、user
```sql
CREATE USER app_user WITH
    PASSWORD 'XXXX'
    CREATEDB
    LOGIN;

CREATE DATABASE app_db
    OWNER app_user
    ENCODING 'UTF8'
    DBCOMPATIBILITY 'A'
    CONNECTION LIMIT 50;

\c app_db

-- 授权模式权限
GRANT USAGE ON SCHEMA public TO app_user;
GRANT CREATE ON SCHEMA public TO app_user;

-- 授权现有表权限（如果已有表）
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO app_user;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO app_user;

-- 设置默认权限（新表自动继承）
ALTER DEFAULT PRIVILEGES IN SCHEMA public
    GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO app_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA public
    GRANT USAGE, SELECT ON SEQUENCES TO app_user;

```

### Create prest.toml
```toml
# Development configuration for pREST

[auth]
enabled = false
table = "prest_users"
username = "prest"
password = "prest"

[http]
host = "0.0.0.0"
port = 3000
timeout = 60

[database]
type = "gaussdb"

[pg]
host = "<gaussdb access host>"
port = <gaussdb port>
user = "app_user"
pass = "XXXXX"
database = "app_db"
ssl.mode = "disable"
cache = false
single = true

[jwt]
default = false
algo = "HS256"
whitelist = ["^\\/auth$"]

[cors]
alloworigin = ["*"]
allowheaders = ["Content-Type"]
allowmethods = ["GET", "HEAD", "POST", "PUT", "DELETE", "OPTIONS"]
allowcredentials = true

[json]
agg.type = "json_agg"

[vector]
enabled = true
default_distance = "cosine"
default_dimensions = 128
index_type = "hnsw"
max_connections = 16
ef_construction = 64
default_vector_field = "embedding"

[cache]
enabled = false
time = 10
storagepath = "./"
sufixfile = ".cache.prestd.db"

[access]
restrict = false # allow access to all tables for development

[expose]
enabled = true
tables = true
schemas = true
databases = true

[debug]
enabled = true

```

### 启动prest 服务
```shell
go run .\cmd\prestd\main.go
```
