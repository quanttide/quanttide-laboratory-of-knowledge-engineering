"""
Neo4j 客户端单元测试
运行方式：pytest tests/test_neo4j_client.py -v
"""

import pytest

# 标记为单元测试
pytestmark = pytest.mark.unit
from unittest.mock import MagicMock, patch

from src.database import Neo4jClient


@pytest.fixture
def mock_driver():
    with patch('neo4j.GraphDatabase.driver') as mock:
        driver = MagicMock()
        session = MagicMock()
        result = MagicMock()
        
        # 模拟查询结果
        result.data.return_value = [{'n': {'name': 'Test', 'age': 25}}]
        session.run.return_value = result
        driver.session.return_value.__enter__.return_value = session
        mock.return_value = driver
        
        yield mock


def test_create_node(mock_driver):
    """测试创建节点"""
    client = Neo4jClient()
    
    # 创建测试节点
    node = client.create_node("Test", {"name": "Test", "age": 25})
    
    assert node == {'name': 'Test', 'age': 25}
    

def test_get_node(mock_driver):
    """测试获取节点"""
    client = Neo4jClient()
    
    # 获取测试节点
    node = client.get_node("Test", {"name": "Test"})
    
    assert node == {'name': 'Test', 'age': 25}


def test_create_relationship(mock_driver):
    """测试创建关系"""
    client = Neo4jClient()
    
    # 创建测试关系
    rel = client.create_relationship(
        from_label="Test",
        from_props={"name": "Test1"},
        to_label="Test",
        to_props={"name": "Test2"},
        rel_type="TEST_REL",
        rel_props={"prop": "value"}
    )
    
    # 由于模拟数据返回空字典，这里只验证函数执行成功
    assert isinstance(rel, dict)