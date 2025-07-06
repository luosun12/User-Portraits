import numpy as np
import pandas as pd
from typing import Dict, List, Tuple
import torch
from torch_geometric.data import Data
import datetime

class DataProcessor:
    def __init__(self):
        self.location_scaler = None
        self.flow_scaler = None
        self.latency_scaler = None
        
    def process_universe_data(self, universe_records: List[Dict]) -> Tuple[torch.Tensor, torch.Tensor, torch.Tensor]:
        """处理Universe表数据，构建图数据结构
        
        Args:
            universe_records: Universe表记录列表
            
        Returns:
            node_features: 节点特征张量 [num_nodes, num_features]
            edge_index: 边连接信息 [2, num_edges]
            edge_attr: 边特征张量 [num_edges, num_edge_features]
        """
        # 提取节点特征
        node_features = []
        for record in universe_records:
            try:
                # 确保数值类型字段为浮点数
                lat = float(record['latitude']) if isinstance(record['latitude'], str) else float(record['latitude'])
                lon = float(record['longitude']) if isinstance(record['longitude'], str) else float(record['longitude'])
                flow = float(record['flow']) if isinstance(record['flow'], str) else float(record['flow'])
                latency = float(record['latency']) if isinstance(record['latency'], str) else float(record['latency'])
                count = float(record['count']) if isinstance(record['count'], str) else float(record['count'])
                err_count = float(record['err_count']) if isinstance(record['err_count'], str) else float(record['err_count'])
                period_id = int(record['period_id']) if isinstance(record['period_id'], str) else int(record['period_id'])
                
                # 正确处理日期
                date_str = record['date']
                try:
                    # 尝试转换日期为时间戳
                    date_ts = float(datetime.datetime.strptime(date_str, '%Y-%m-%d').timestamp())
                except ValueError:
                    # 如果格式不对，尝试其他格式或使用当前时间
                    print(f"无法解析日期: {date_str}，使用当前时间代替")
                    date_ts = float(datetime.datetime.now().timestamp())
                
                features = [
                    lat,
                    lon,
                    flow,
                    latency,
                    count,
                    err_count,
                    period_id,
                    date_ts
                ]
                node_features.append(features)
            except Exception as e:
                print(f"处理记录时出错: {e}, 记录: {record}")
                continue
            
        if not node_features:
            # 如果没有有效特征，返回空张量
            return torch.tensor([], dtype=torch.float), torch.tensor([], dtype=torch.long), torch.tensor([], dtype=torch.float)
            
        node_features = torch.tensor(node_features, dtype=torch.float)
        
        # 构建边连接
        num_nodes = len(node_features)
        edge_index = []
        edge_attr = []
        
        # 改进边构建逻辑
        try:
            # 按时间顺序连接
            for i in range(num_nodes - 1):
                j = i + 1  # 只连接相邻时间点
                edge_index.extend([[i, j], [j, i]])  # 双向边
                # 边特征：距离、时间差
                dist = torch.norm(node_features[i, :2] - node_features[j, :2]).item()
                time_diff = abs(node_features[i, 6].item() - node_features[j, 6].item())
                edge_attr.extend([[dist, time_diff], [dist, time_diff]])
            
            # 验证边索引不超出范围
            for edge in edge_index:
                if edge[0] >= num_nodes or edge[1] >= num_nodes:
                    print(f"警告: 边索引 {edge} 超出节点范围 {num_nodes}")
                    edge_index.remove(edge)
                    # 如果是双向边，一次移除两条
                    reverse_edge = [edge[1], edge[0]]
                    if reverse_edge in edge_index:
                        edge_index.remove(reverse_edge)
        except Exception as e:
            print(f"构建边时出错: {e}")
            # 出错时使用安全的默认边
            edge_index = []
            edge_attr = []
        
        # 确保边存在
        if not edge_index:
            # 如果没有符合条件的边，创建自循环
            edge_index = [[i, i] for i in range(num_nodes)]
            edge_attr = [[0.0, 0.0] for _ in range(num_nodes)]
        
        edge_index = torch.tensor(edge_index, dtype=torch.long).t() if edge_index else torch.zeros((2, 0), dtype=torch.long)
        edge_attr = torch.tensor(edge_attr, dtype=torch.float) if edge_attr else torch.zeros((0, 2), dtype=torch.float)
        
        return node_features, edge_index, edge_attr
    
    def normalize_features(self, node_features: torch.Tensor) -> torch.Tensor:
        """特征标准化"""
        mean = node_features.mean(dim=0, keepdim=True)
        std = node_features.std(dim=0, keepdim=True)
        return (node_features - mean) / (std + 1e-8)
    
    def create_time_windows(self, node_features: torch.Tensor, window_size: int = 24) -> List[torch.Tensor]:
        """创建时间窗口特征"""
        num_nodes = len(node_features)
        windows = []
        
        if num_nodes == 0:
            return windows
            
        # 如果数据数量少于窗口大小，则调整窗口大小
        actual_window_size = min(window_size, num_nodes)
        
        if actual_window_size == 1:
            # 如果只有一条数据，直接返回
            return [node_features]
            
        for i in range(num_nodes - actual_window_size + 1):
            windows.append(node_features[i:i+actual_window_size])
            
        return windows
    
    def prepare_prediction_data(self, current_data: Dict) -> Data:
        """准备预测用的数据"""
        node_features, edge_index, edge_attr = self.process_universe_data([current_data])
        node_features = self.normalize_features(node_features)
        
        return Data(
            x=node_features,
            edge_index=edge_index,
            edge_attr=edge_attr
        )
