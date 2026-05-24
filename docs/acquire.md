好的，回到 Python。你是对的——数据密集型工作，先验证逻辑再考虑性能优化。既然已有 extract 命令，我们就在现有结构上增加 acquire，作为 v0.3.0 的核心功能。

—

acquire 命令设计（Python 实现方案）

1. 在现有 CLI 中的位置

假设你的命令行工具已有类似结构：

```
qknow
├── extract   # 已有：基于宪法的知识抽取
└── acquire   # 新增：获取 + 认知密度过滤
```

acquire 负责从原材料中提取可用的知识，它内部包含认知密度判断作为前置过滤。

—

2. 核心模块

```
knowledge_cloud/
├── cli/
│   ├── acquire.py       # acquire 命令入口
│   └── extract.py       # 已有
├── acquisition/
│   ├── density.py       # 认知密度评估器
│   ├── splitter.py      # 文本分块
│   ├── llm_client.py    # LLM 调用封装
│   └── models.py        # 数据结构
└── ...
```

—

3. 数据结构（models.py）

```python
from enum import Enum
from dataclasses import dataclass, field
from typing import Optional

class DensityLevel(str, Enum):
    HIGH = ”high“
    MEDIUM = ”medium“
    LOW = ”low“

@dataclass
class NovelPoint:
    content: str
    confidence: float

@dataclass
class ChunkAssessment:
    chunk_index: int
    density_level: DensityLevel
    novelty_score: float  # 0.0 - 1.0
    novel_points: list[NovelPoint]
    reasoning: str

@dataclass
class FileAssessment:
    file_path: str
    overall_density: DensityLevel
    average_novelty_score: float
    chunks: list[ChunkAssessment]
    novel_points_aggregated: list[NovelPoint]
    suggestions: str
```

—

4. 认知密度评估器（density.py）

这是核心。用你的提示词思想：判断大模型知识库没有的内容。

```python
import json
from .llm_client import LLMClient
from .models import DensityLevel, ChunkAssessment, NovelPoint

PROMPT_TEMPLATE = ”“”You are a cognitive density assessor. Your task is to evaluate how much new, unique, or non-obvious information a piece of text contains, relative to what a general large language model already knows.

Rules:
1. Identify statements that are:
   - Novel frameworks, definitions, or models
   - Specific operational insights (e.g., “I found that X must be separated from Y because...”)
   - Personal discoveries or counter-intuitive observations
   - Domain-specific patterns that are not widely documented
2. Ignore common knowledge, widely documented practices, and generic explanations.
3. Provide a density level (high/medium/low) based on the proportion of novel content.
4. Output must be strictly valid JSON with the following structure:

{{
  “density_level”: “high” | “medium” | “low”,
  “novelty_score”: 0.0-1.0,
  “novel_points”: [
    {{
      “content”: “short description of the new insight”,
      “confidence”: 0.0-1.0
    }}
  ],
  “reasoning”: “Brief explanation of the assessment.”
}}

Now evaluate the following text:
—TEXT START—
{content}
—TEXT END—“”“

class DensityAssessor:
    def __init__(self, llm_client: LLMClient):
        self.llm = llm_client

    def assess_chunk(self, text: str, chunk_index: int = 0) -> ChunkAssessment:
        prompt = PROMPT_TEMPLATE.format(content=text)
        response = self.llm.complete(prompt)  # 返回文本，需解析 JSON
        try:
            data = json.loads(response)
            return ChunkAssessment(
                chunk_index=chunk_index,
                density_level=DensityLevel(data[”density_level“]),
                novelty_score=data[”novelty_score“],
                novel_points=[NovelPoint(**p) for p in data[”novel_points“]],
                reasoning=data[”reasoning“]
            )
        except (json.JSONDecodeError, KeyError) as e:
            # 失败时默认返回低密度，安全兜底
            return ChunkAssessment(
                chunk_index=chunk_index,
                density_level=DensityLevel.LOW,
                novelty_score=0.0,
                novel_points=[],
                reasoning=f”Assessment failed: {e}“
            )
```

—

5. 文件处理流程（acquire.py 命令主体）

```python
import asyncio
from pathlib import Path
from .splitter import split_text
from .density import DensityAssessor
from .models import FileAssessment, DensityLevel

async def evaluate_file(
    file_path: Path,
    assessor: DensityAssessor,
    chunk_size: int = 3000,
    whole_file: bool = False
) -> FileAssessment:
    text = file_path.read_text(encoding=”utf-8“)
    chunks = [text] if whole_file else split_text(text, chunk_size)
    
    assessments = []
    for i, chunk in enumerate(chunks):
        # 这里可并发调 LLM，但注意速率限制
        assessment = assessor.assess_chunk(chunk, chunk_index=i)
        assessments.append(assessment)
    
    # 聚合
    avg_score = sum(a.novelty_score for a in assessments) / len(assessments)
    if avg_score > 0.6:
        overall = DensityLevel.HIGH
    elif avg_score > 0.3:
        overall = DensityLevel.MEDIUM
    else:
        overall = DensityLevel.LOW
    
    # 去重新颖点
    all_points = [p for a in assessments for p in a.novel_points]
    unique_points = deduplicate_points(all_points)  # 简单实现：按 content 相似度去重
    
    suggestions = {
        DensityLevel.HIGH: ”可直接进入建模/抽取流程“,
        DensityLevel.MEDIUM: ”建议人工复核后进入抽取“,
        DensityLevel.LOW: ”建议跳过或仅作为背景资料“
    }.get(overall, ”“)
    
    return FileAssessment(
        file_path=str(file_path),
        overall_density=overall,
        average_novelty_score=avg_score,
        chunks=assessments,
        novel_points_aggregated=unique_points,
        suggestions=suggestions
    )
```

—

6. CLI 集成

```python
# cli/acquire.py
import click
from pathlib import Path
from acquisition.density import DensityAssessor
from acquisition.llm_client import OpenAIClient

@click.command()
@click.argument(”paths“, nargs=-1, type=click.Path(exists=True))
@click.option(”—chunk-size“, default=3000, help=”最大分块字符数“)
@click.option(”—whole“, is_flag=True, help=”不分块，整体评估“)
@click.option(”—filter“, type=click.Choice([”high“, ”medium“, ”low“]), help=”只输出指定密度及以上的文件“)
@click.option(”—output“, default=”json“, type=click.Choice([”json“, ”summary“]))
def acquire(paths, chunk_size, whole, filter, output):
    ”“”获取知识：对文件进行认知密度评估，并提取高密度材料“”“
    client = OpenAIClient()  # 从配置读取 API key
    assessor = DensityAssessor(client)
    
    for path_str in paths:
        path = Path(path_str)
        if path.is_dir():
            files = list(path.glob(”*.md“))  # 可根据需要扩展
        else:
            files = [path]
        
        for file in files:
            assessment = asyncio.run(evaluate_file(file, assessor, chunk_size, whole))
            
            if filter:
                if assessment.overall_density.value < filter:
                    continue
            
            if output == ”json“:
                click.echo(json.dumps(assessment.__dict__, default=str))
            else:
                click.echo(f”{assessment.file_path}: {assessment.overall_density.value} (score: {assessment.average_novelty_score:.2f})“)
```

—

7. 与 extract 的衔接

acquire 输出的高密度文件或新颖点，可以直接作为 extract 的输入。你可以在工作流中这样编排：

```bash
# 先获取高密度材料
qknow acquire ./materials —filter high —output json > high_density.json

# 然后将新颖点或文件列表传给 extract
qknow extract —source high_density.json —standard audit_model
```

或者，在 acquire 内部集成一个“自动传递给 extract”的选项：

```python
@click.option(”—auto-extract“, is_flag=True, help=”对高密度文件直接运行抽取“)
```

—

8. 验证思路

在 v0.3.0 发布前，先用你自己的高密度上下文做几组对比测试：

1. 已知高密度文本：你的思考记录 → 应评为 HIGH，新颖点准确。
2. 已知低密度文本：常见的技术文档摘要、模板化周报 → 应评为 LOW 或 MEDIUM。
3. 混合文本：穿插段落，验证分块和聚合逻辑是否合理。

通过这个小闭环，确认提示词和解析逻辑可靠后，再正式纳入流水线。这符合你“先手动测，再自动化”的演化原则。