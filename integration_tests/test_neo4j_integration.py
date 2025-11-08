"""
Neo4j 集成测试
运行方式：pytest tests/integration/test_neo4j_integration.py -v

注意：运行此测试需要有可用的 Neo4j 数据库连接
"""

import os
import pytest
from src.database import Neo4jClient


def test_neo4j_connection():
    """测试 Neo4j 连接是否正常"""
    client = Neo4jClient()
    try:
        # 尝试创建一个测试节点
        node = client.create_node("TestNode", {"test_id": "connection_test"})
        assert node is not None, "应该成功创建测试节点"
    except Exception as e:
        pytest.fail(f"Neo4j 连接测试失败: {str(e)}")
    finally:
        client.close()


def test_full_crud_operations():
    """测试完整的 CRUD 操作流程"""
    with Neo4jClient() as client:
        # 1. 创建测试节点
        test_props = {
            "name": "集成测试",
            "test_id": "integration_test",
            "timestamp": "2025-11-08"
        }
        created_node = client.create_node("TestNode", test_props)
        assert created_node, "节点应该被成功创建"
        
        # 2. 查询节点
        found_node = client.get_node("TestNode", {"test_id": "integration_test"})
        assert found_node, "应该能找到创建的节点"
        assert found_node.get("name") == "集成测试", "节点属性应该匹配"
        
        # 3. 创建关系
        other_props = {
            "name": "关联节点",
            "test_id": "related_node"
        }
        # 创建关联节点
        client.create_node("TestNode", other_props)
        
        # 创建关系
        rel = client.create_relationship(
            from_label="TestNode",
            from_props={"test_id": "integration_test"},
            to_label="TestNode",
            to_props={"test_id": "related_node"},
            rel_type="TEST_RELATION",
            rel_props={"test_date": "2025-11-08"}
        )
        assert rel, "关系应该被成功创建"


def test_error_handling():
    """测试错误处理"""
    with Neo4jClient() as client:
        # 测试查询不存在的节点
        non_existent = client.get_node("TestNode", {"test_id": "non_existent"})
        assert non_existent is None, "不存在的节点应该返回 None"
        
        # 测试创建重复节点（如果有唯一约束的话）
        try:
            client.create_node("TestNode", {"test_id": "duplicate"})
            client.create_node("TestNode", {"test_id": "duplicate"})
        except Exception as e:
            # 如果数据库配置了唯一约束，这里会抛出异常
            print(f"预期的重复键错误: {str(e)}")


if __name__ == "__main__":
    # 可以直接运行此文件来执行集成测试
    pytest.main([__file__, "-v"])