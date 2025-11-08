"""
Neo4j 数据库客户端
"""

import os
from typing import Any, Dict, List, Optional, Union

from dotenv import load_dotenv
from neo4j import GraphDatabase, Driver, Session, Transaction


class Neo4jClient:
    """
    Neo4j 数据库客户端类
    """
    def __init__(self, uri: Optional[str] = None, username: Optional[str] = None, 
                 password: Optional[str] = None):
        """
        初始化 Neo4j 客户端
        
        Args:
            uri: Neo4j 数据库连接URI，默认从环境变量 NEO4J_URI 读取
            username: 用户名，默认从环境变量 NEO4J_USERNAME 读取
            password: 密码，默认从环境变量 NEO4J_PASSWORD 读取
        """
        # 加载环境变量
        load_dotenv()
        
        # 设置连接参数
        self._uri = uri or os.getenv('NEO4J_URI', 'bolt://localhost:7687')
        self._username = username or os.getenv('NEO4J_USERNAME', 'neo4j')
        self._password = password or os.getenv('NEO4J_PASSWORD', 'password')
        
        # 创建驱动实例
        self._driver: Driver = GraphDatabase.driver(
            self._uri, 
            auth=(self._username, self._password)
        )

    def __enter__(self) -> 'Neo4jClient':
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        self.close()

    def close(self):
        """
        关闭数据库连接
        """
        self._driver.close()

    def _run_query(self, query: str, parameters: Optional[Dict[str, Any]] = None) -> List[Dict[str, Any]]:
        """
        执行 Cypher 查询
        
        Args:
            query: Cypher 查询语句
            parameters: 查询参数

        Returns:
            查询结果列表
        """
        with self._driver.session() as session:
            result = session.run(query, parameters or {})
            return [record.data() for record in result]

    def create_node(self, label: str, properties: Dict[str, Any]) -> Dict[str, Any]:
        """
        创建节点
        
        Args:
            label: 节点标签
            properties: 节点属性

        Returns:
            创建的节点数据
        """
        query = f"CREATE (n:{label} $props) RETURN n"
        result = self._run_query(query, {"props": properties})
        return result[0]['n'] if result else {}

    def get_node(self, label: str, properties: Dict[str, Any]) -> Optional[Dict[str, Any]]:
        """
        获取节点
        
        Args:
            label: 节点标签
            properties: 匹配属性

        Returns:
            节点数据，如果不存在则返回 None
        """
        conditions = " AND ".join(f"n.{k} = ${k}" for k in properties.keys())
        query = f"MATCH (n:{label}) WHERE {conditions} RETURN n"
        result = self._run_query(query, properties)
        return result[0]['n'] if result else None

    def create_relationship(self, from_label: str, from_props: Dict[str, Any],
                          to_label: str, to_props: Dict[str, Any],
                          rel_type: str, rel_props: Optional[Dict[str, Any]] = None) -> Dict[str, Any]:
        """
        创建关系
        
        Args:
            from_label: 起始节点标签
            from_props: 起始节点匹配属性
            to_label: 目标节点标签
            to_props: 目标节点匹配属性
            rel_type: 关系类型
            rel_props: 关系属性

        Returns:
            创建的关系数据
        """
        from_conditions = " AND ".join(f"n1.{k} = $from_{k}" for k in from_props.keys())
        to_conditions = " AND ".join(f"n2.{k} = $to_{k}" for k in to_props.keys())
        
        # 构建参数字典
        params = {
            f"from_{k}": v for k, v in from_props.items()
        }
        params.update({f"to_{k}": v for k, v in to_props.items()})
        if rel_props:
            params["rel_props"] = rel_props

        # 构建查询语句
        query = f"""
        MATCH (n1:{from_label}), (n2:{to_label})
        WHERE {from_conditions} AND {to_conditions}
        CREATE (n1)-[r:{rel_type} $rel_props]->(n2)
        RETURN r
        """
        
        result = self._run_query(query, params)
        return result[0]['r'] if result else {}