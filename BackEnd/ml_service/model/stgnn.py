"""
网络结构:
    输入层:
        - 节点特征: [batch_size, num_nodes, 8] (位置[2], 流量[1], 延迟[1], 访问计数[1], 错误计数[1], 时间信息[2])
        - 边特征: [num_edges, 2] (空间距离[1], 时间差[1])
        
    编码器:
        1. 特征投影层: Linear(8 -> 64)
        2. 时空图卷积块 × 3:
            - 空间: GAT卷积 (考虑边权重的注意力机制)
            - 时间: 自注意力层
            - 残差连接和层归一化
            
    解码器:
        1. 全局平均池化
        2. MLP预测头:
            - Linear(64 -> 64) + ReLU + Dropout(0.2)
            - Linear(64 -> 3)
            
损失函数:
    - MSE损失 (均方误差)
    - L2正则化 (权重衰减)
    
优化算法:
    - Adam优化器
    - 学习率: 0.001
    - weight_decay: 1e-5
"""

import torch
import torch.nn as nn
import torch.nn.functional as F
from torch_geometric.nn import GCNConv, GATConv
from torch_geometric.nn import global_mean_pool

class TemporalAttention(nn.Module):
    def __init__(self, hidden_dim):
        super().__init__()
        self.att_weight = nn.Parameter(torch.Tensor(hidden_dim, hidden_dim))
        self.bias = nn.Parameter(torch.Tensor(hidden_dim))
        nn.init.xavier_uniform_(self.att_weight)
        nn.init.zeros_(self.bias)
        
    def forward(self, x):
        # x shape: [batch_size, seq_len, hidden_dim]
        # 计算查询(Q)和键(K)的注意力分数
        Q = torch.matmul(x, self.att_weight)  # [batch_size, seq_len, hidden_dim]
        K = x  # [batch_size, seq_len, hidden_dim]
        
        # 计算注意力分数 (Q * K^T)
        attention_scores = torch.matmul(Q, K.transpose(-2, -1))  # [batch_size, seq_len, seq_len]
        attention_scores = attention_scores / torch.sqrt(torch.tensor(x.size(-1), dtype=torch.float32, device=x.device))
        
        # 应用softmax获取注意力权重
        attention_weights = F.softmax(attention_scores, dim=-1)  # [batch_size, seq_len, seq_len]
        
        # 应用注意力权重到值(V)
        V = x  # [batch_size, seq_len, hidden_dim]
        output = torch.matmul(attention_weights, V)  # [batch_size, seq_len, hidden_dim]
        
        # 添加偏置并返回
        output = output + self.bias.unsqueeze(0).unsqueeze(0)
        return output

class SpatialTemporalBlock(nn.Module):
    def __init__(self, in_channels, hidden_channels):
        super().__init__()
        self.conv1 = GATConv(in_channels, hidden_channels)
        self.conv2 = GATConv(hidden_channels, hidden_channels)
        self.temporal_attention = TemporalAttention(hidden_channels)
        self.norm = nn.LayerNorm(hidden_channels)
        
    def forward(self, x, edge_index, edge_attr):
        # 验证边索引
        if edge_index.numel() > 0:
            num_nodes = x.size(0)
            # 检查边索引是否超出节点范围
            if edge_index.max() >= num_nodes:
                print(f"警告: 边索引 {edge_index.max().item()} 超出节点范围 {num_nodes}")
                # 过滤掉无效的边
                mask = (edge_index[0] < num_nodes) & (edge_index[1] < num_nodes)
                edge_index = edge_index[:, mask]
                if edge_attr is not None:
                    edge_attr = edge_attr[mask]
        
        # 特殊情况处理：没有边
        if edge_index.size(1) == 0:
            # 如果没有边，使用自循环
            edge_index = torch.stack([
                torch.arange(x.size(0), device=x.device),
                torch.arange(x.size(0), device=x.device)
            ], dim=0)
            if edge_attr is not None:
                edge_attr = torch.zeros((edge_index.size(1), edge_attr.size(1)), 
                                      device=edge_attr.device, 
                                      dtype=edge_attr.dtype)
        
        # 空间信息聚合
        h = self.conv1(x, edge_index, edge_attr)
        h = F.relu(h)
        h = F.dropout(h, p=0.2, training=self.training)
        h = self.conv2(h, edge_index, edge_attr)
        
        # 时间注意力
        batch_size = 1 if len(h.shape) == 2 else h.shape[0]
        h = h.view(batch_size, -1, h.shape[-1])  # [batch_size, nodes, features]
        h = self.temporal_attention(h)
        h = h.view(-1, h.shape[-1])  # 展平回原始形状
        
        # 残差连接和归一化
        h = h + x
        h = self.norm(h)
        return h

class STGNN(nn.Module):
    def __init__(self, 
                 input_dim: int = 8,        # 输入特征维度
                 hidden_dim: int = 64,      # 隐藏层维度
                 output_dim: int = 3,       # 输出维度（位置、流量、延迟）
                 num_layers: int = 3):      # ST-GCN层数
        super().__init__()
        
        self.input_proj = nn.Linear(input_dim, hidden_dim)
        
        # 多层时空图卷积块
        self.st_blocks = nn.ModuleList([
            SpatialTemporalBlock(hidden_dim, hidden_dim)
            for _ in range(num_layers)
        ])
        
        # 预测头
        self.pred_head = nn.Sequential(
            nn.Linear(hidden_dim, hidden_dim),
            nn.ReLU(),
            nn.Dropout(0.2),
            nn.Linear(hidden_dim, output_dim)
        )
        
    def forward(self, data):
        x, edge_index, edge_attr = data.x, data.edge_index, data.edge_attr
        
        # 特征投影
        h = self.input_proj(x)
        
        # 时空图卷积
        for block in self.st_blocks:
            h = block(h, edge_index, edge_attr)
        
        # 全局池化和预测
        if hasattr(data, 'batch'):
            h = global_mean_pool(h, data.batch)
        else:
            h = h.mean(dim=0, keepdim=True)
            
        # 预测未来值
        pred = self.pred_head(h)
        
        return pred  # [batch_size, output_dim]
    
    def predict(self, data):
        """预测接口"""
        self.eval()
        with torch.no_grad():
            pred = self.forward(data)
        return pred.numpy()
