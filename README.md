# 量潮知识工程示例仓库

这是一个使用 Neo4j 图数据库的 Python 项目示例。

## 项目结构

```
quanttide-example-of-knowledge-engineering/
├── src/
│   └── database/
├── tests/                 # 单元测试目录
│   └── test_neo4j_client.py
├── integration_tests/     # 集成测试目录
│   └── test_neo4j_integration.py
├── pytest.ini
└── README.md
```

## 功能特性

- Neo4j 数据库连接管理
- 节点的创建和查询
- 关系的创建和管理
- 完整的单元测试
- 环境变量配置支持

## 快速开始

1. 安装依赖

```bash
# 使用 pip
pip install -e .

# 开发环境额外依赖（用于运行测试）
pip install -e ".[dev]"
```

2. 配置环境变量

复制 `.env.example` 文件到 `.env` 并修改配置：

```bash
cp .env.example .env
# 编辑 .env 文件，设置你的 Neo4j 连接信息
```

3. 运行示例

```bash
python -m src
```

## 开发

### 运行测试

项目包含两类测试，分别位于不同目录：
- 单元测试：`tests/` 目录
- 集成测试：`integration_tests/` 目录

运行测试的方式：

1. 运行单元测试（不需要数据库）：
```bash
pytest tests/
```

2. 运行集成测试（需要 Neo4j 数据库）：
```bash
pytest integration_tests/
```

3. 运行所有测试：
```bash
pytest tests/ integration_tests/
```

4. 运行特定测试文件：
```bash
# 运行特定单元测试
pytest tests/test_neo4j_client.py -v

# 运行特定集成测试
pytest integration_tests/test_neo4j_integration.py -v
```

注意：运行集成测试前请确保：
1. Neo4j 数据库已启动且可访问
2. `.env` 文件中的连接信息配置正确

## Neo4j 客户端使用示例

```python
from src.database import Neo4jClient

# 创建客户端实例
with Neo4jClient() as client:
    # 创建节点
    person = client.create_node("Person", {
        "name": "张三",
        "age": 30
    })
    
    # 获取节点
    found = client.get_node("Person", {"name": "张三"})
    
    # 创建关系
    client.create_relationship(
        from_label="Person",
        from_props={"name": "张三"},
        to_label="Person",
        to_props={"name": "李四"},
        rel_type="KNOWS",
        rel_props={"since": 2023}
    )
