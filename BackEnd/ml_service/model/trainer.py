import torch
import torch.nn as nn
import torch.optim as optim
from torch_geometric.loader import DataLoader
from torch_geometric.data import Data
import numpy as np
from typing import List, Dict, Optional, Tuple
import os
from datetime import datetime

from data.processor import DataProcessor
from data.loader import DataLoader as DBLoader
from model.stgnn import STGNN

class ModelTrainer:
    def __init__(self,
                 model: Optional[STGNN] = None,
                 db_config: Dict = None,
                 model_save_path: str = "./checkpoints"):
        """初始化训练器
        
        Args:
            model: STGNN模型实例
            db_config: 数据库配置
            model_save_path: 模型保存路径
        """
        self.model = model if model else STGNN()
        self.processor = DataProcessor()
        self.db_loader = DBLoader(**db_config) if db_config else None
        self.model_save_path = model_save_path
        self.device = torch.device('cuda' if torch.cuda.is_available() else 'cpu')
        
        # 创建模型保存目录
        os.makedirs(model_save_path, exist_ok=True)
        
    def _create_window_edges(self, window_features: torch.Tensor) -> Tuple[torch.Tensor, torch.Tensor]:
        """为窗口创建边索引和边属性
        
        Args:
            window_features: 窗口特征张量 [window_size, num_features]
            
        Returns:
            edge_index: 边索引张量 [2, num_edges]
            edge_attr: 边属性张量 [num_edges, 2]
        """
        num_nodes = window_features.shape[0]
        edge_index = []
        edge_attr = []
        
        # 按时间顺序连接相邻节点
        for i in range(num_nodes - 1):
            j = i + 1
            edge_index.extend([[i, j], [j, i]])  # 双向边
            
            # 边特征：距离、时间差
            dist = torch.norm(window_features[i, :2] - window_features[j, :2]).item()
            time_diff = abs(window_features[i, 6].item() - window_features[j, 6].item())
            edge_attr.extend([[dist, time_diff], [dist, time_diff]])
        
        # 添加自循环
        for i in range(num_nodes):
            edge_index.append([i, i])
            edge_attr.append([0.0, 0.0])
            
        edge_index = torch.tensor(edge_index, dtype=torch.long).t()
        edge_attr = torch.tensor(edge_attr, dtype=torch.float)
        
        return edge_index, edge_attr
        
    def prepare_training_data(self, start_date: str, end_date: str, mock_data=None) -> List[Dict]:
        """准备训练数据
        
        Args:
            start_date: 开始日期
            end_date: 结束日期
            mock_data: 可选的模拟数据，如果提供则使用模拟数据而不是从数据库加载
            
        Returns:
            训练数据列表
        """
        # 处理数据
        try:
            # 获取数据
            if mock_data is not None:
                print(f"使用提供的模拟数据，数量: {len(mock_data)}")
                raw_data = mock_data
            else:
                # 加载原始数据
                raw_data = self.db_loader.load_universe_data(start_date, end_date)
                print(f"从数据库加载了 {len(raw_data)} 条数据")
            
            if not raw_data:
                print("警告: 没有数据")
                return []
            
            print(f"原始数据数量: {len(raw_data)}")
            
            # 按用户ID分组处理数据
            all_training_data = []
            user_ids = set(record['user_id'] for record in raw_data)
            print(f"数据中包含 {len(user_ids)} 个用户")
            
            for user_id in user_ids:
                user_data = [record for record in raw_data if record['user_id'] == user_id]
                print(f"用户 {user_id} 的数据数量: {len(user_data)}")
                
                # 处理单个用户的数据
                try:
                    # 按时间排序
                    user_data.sort(key=lambda x: (x['date'], x['period_id']))
                    
                    node_features, edge_index, edge_attr = self.processor.process_universe_data(user_data)
                    
                    if len(node_features) == 0:
                        print(f"警告: 用户 {user_id} 处理后没有有效的节点特征")
                        continue
                        
                    print(f"用户 {user_id} 处理后的节点特征数量: {len(node_features)}")
                    
                    # 创建时间窗口
                    window_size = 24  # 24小时作为一个窗口
                    windows = self.processor.create_time_windows(node_features, window_size)
                    
                    if not windows or len(windows) < 2:
                        print(f"警告: 用户 {user_id} 没有足够的时间窗口进行训练")
                        continue
                        
                    print(f"用户 {user_id} 创建的时间窗口数量: {len(windows)}")
                    
                    # 构建训练样本
                    for i in range(len(windows) - 1):
                        try:
                            # 确保窗口是tensor类型
                            curr_window = windows[i]
                            next_window = windows[i+1]
                            
                            # 为当前窗口创建适当的边索引和边属性
                            window_edge_index, window_edge_attr = self._create_window_edges(curr_window)
                            
                            # 安全地获取目标索引
                            y_indices = [0, 1, 2]  # 经度、纬度、流量
                            
                            # 确保特征索引在范围内
                            if next_window.shape[1] <= max(y_indices):
                                print(f"警告: 窗口特征维度 {next_window.shape[1]} 不足以获取索引 {y_indices}")
                                continue
                                
                            # 使用下一个窗口的第一个时间点作为预测目标
                            y = next_window[0, y_indices]
                            
                            data = {
                                'x': curr_window,
                                'edge_index': window_edge_index,
                                'edge_attr': window_edge_attr,
                                'y': y
                            }
                            
                            all_training_data.append(data)
                            
                        except Exception as e:
                            print(f"用户 {user_id} 构建第 {i} 个训练样本失败: {e}")
                            continue
                            
                except Exception as e:
                    print(f"处理用户 {user_id} 的数据时出错: {e}")
                    continue
            
            if not all_training_data:
                print("警告: 没有成功构建任何训练样本")
            else:
                print(f"成功构建了 {len(all_training_data)} 个训练样本")
                
            return all_training_data
            
        except Exception as e:
            print(f"准备训练数据时出错: {e}")
            import traceback
            traceback.print_exc()
            return []
        
    def train(self,
              training_data: List,
              epochs: int = 100,
              batch_size: int = 32,
              learning_rate: float = 0.001,
              save_interval: int = 10):
        """训练模型"""
        self.model = self.model.to(self.device)
        optimizer = optim.Adam(self.model.parameters(), lr=learning_rate)
        criterion = nn.MSELoss()
        
        # 验证和清理训练数据
        processed_data = []
        
        for item_idx, item in enumerate(training_data):
            try:
                # 获取节点数量
                num_nodes = item['x'].shape[0]
                
                # 检查边索引是否超出节点范围
                edge_index = item['edge_index']
                if edge_index.shape[0] > 0:
                    max_node_idx = edge_index.max().item()
                    if max_node_idx >= num_nodes:
                        print(f"警告: 样本 {item_idx} 的边索引 {max_node_idx} 超出节点范围 {num_nodes}，跳过此样本")
                        continue
                
                # 将数据移动到设备
                data = Data(
                    x=item['x'].to(self.device),
                    edge_index=item['edge_index'].to(self.device),
                    edge_attr=item['edge_attr'].to(self.device),
                    y=item['y'].to(self.device)
                )
                processed_data.append(data)
            except Exception as e:
                print(f"处理训练样本 {item_idx} 时出错: {e}")
                continue
                
        if not processed_data:
            print("警告: 没有可用的训练数据")
            return
        
        print(f"处理后的训练样本数量: {len(processed_data)}")
        
        # 创建数据加载器
        dataloader = DataLoader(processed_data, batch_size=batch_size, shuffle=True)
        
        # 训练循环
        try:
            for epoch in range(epochs):
                self.model.train()
                total_loss = 0
                
                for batch_idx, batch in enumerate(dataloader):
                    try:
                        # 数据已经在device上，不需要再次移动
                        optimizer.zero_grad()
                        
                        # 前向传播
                        pred = self.model(batch)
                        loss = criterion(pred, batch.y)
                        
                        # 反向传播
                        loss.backward()
                        optimizer.step()
                        
                        total_loss += loss.item()
                    except Exception as e:
                        print(f"处理批次 {batch_idx} 时出错: {e}")
                        import traceback
                        traceback.print_exc()
                        continue
                    
                # 打印训练信息
                avg_loss = total_loss / len(dataloader) if len(dataloader) > 0 else float('inf')
                print(f"Epoch {epoch+1}/{epochs}, Loss: {avg_loss:.4f}")
                
                # 定期保存模型
                if (epoch + 1) % save_interval == 0 or epoch == epochs - 1:
                    self.save_model(f"model_epoch_{epoch+1}")
        except Exception as e:
            print(f"训练时发生错误: {e}")
            import traceback
            traceback.print_exc()
        
    def save_model(self, name: str):
        """保存模型"""
        timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
        save_path = os.path.join(self.model_save_path, f"{name}_{timestamp}.pt")
        
        # 保存模型状态和配置
        torch.save({
            'model_state_dict': self.model.state_dict(),
            'model_config': {
                'input_dim': self.model.input_proj.in_features,
                'hidden_dim': self.model.st_blocks[0].conv1.in_channels,
                'output_dim': self.model.pred_head[-1].out_features,
                'num_layers': len(self.model.st_blocks)
            }
        }, save_path)
        
    def load_model(self, model_path: str):
        """加载模型"""
        checkpoint = torch.load(model_path, map_location=self.device)
        
        # 重新创建模型
        config = checkpoint['model_config']
        self.model = STGNN(**config)
        
        # 加载模型参数
        self.model.load_state_dict(checkpoint['model_state_dict'])
        self.model.to(self.device)
        
    def predict(self, current_data: Dict) -> np.ndarray:
        """预测接口"""
        self.model.eval()
        
        # 准备预测数据
        data = self.processor.prepare_prediction_data(current_data)
        data = data.to(self.device)
        
        # 预测
        with torch.no_grad():
            pred = self.model(data)
            
        return pred.cpu().numpy()
