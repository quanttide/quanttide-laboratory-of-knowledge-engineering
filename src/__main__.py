"""
示例用法
"""

from database import Neo4jClient


def main():
    # 创建 Neo4j 客户端示例（使用默认连接参数）
    with Neo4jClient() as client:
        # 创建示例节点
        person = client.create_node("Person", {
            "name": "张三",
            "age": 30
        })
        print("创建的人物节点:", person)

        # 获取节点
        found_person = client.get_node("Person", {"name": "张三"})
        print("查询到的人物节点:", found_person)

        # 创建关系示例
        client.create_relationship(
            from_label="Person",
            from_props={"name": "张三"},
            to_label="Person",
            to_props={"name": "李四"},
            rel_type="KNOWS",
            rel_props={"since": 2023}
        )
        print("已创建 '认识' 关系")


if __name__ == "__main__":
    main()
