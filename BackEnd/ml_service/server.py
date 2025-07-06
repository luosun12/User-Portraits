from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
from typing import Dict, List, Optional
import uvicorn
import json
import os
import random
import torch
from dotenv import load_dotenv
from datetime import datetime, timedelta
from contextlib import asynccontextmanager
from fastapi.responses import JSONResponse

from data.loader import DataLoader
from data.processor import DataProcessor
from model.trainer import ModelTrainer
from model.stgnn import STGNN

load_dotenv()

def get_db_host():
    """获取数据库主机地址，包含端口号"""
    host = os.getenv('DB_HOST', '127.0.0.1')
    return f"{host}"

# 配置
CONFIG = {
    'db': {
        'host': get_db_host(),
        'user': os.getenv('DB_USER'),
        'password': os.getenv('DB_PASSWORD'),
        'database': os.getenv('DB_NAME'),
    },
    'model': {
        'checkpoint_dir': os.getenv('MODEL_SAVE_PATH', './checkpoints'),
        'latest_model': None  # 将在启动时更新
    },
    'server': {
        'host': os.getenv('SERVICE_HOST', '0.0.0.0'),
        'port': int(os.getenv('SERVICE_PORT', '8000'))
    }
}

@asynccontextmanager
async def lifespan(app: FastAPI):
    """服务生命周期管理"""
    if not load_latest_model():
        print("Warning: No pre-trained model found.")
    yield

# 创建FastAPI应用
app = FastAPI(title="STGNN-ML-Service", lifespan=lifespan)

# 创建服务实例
db_loader = DataLoader(**CONFIG['db'])
processor = DataProcessor()
trainer = ModelTrainer(db_config=CONFIG['db'], 
                      model_save_path=CONFIG['model']['checkpoint_dir'])

# 请求模型
class PredictionRequest(BaseModel):
    user_id: int
    station_id: int
    current_time: str

class TrainingRequest(BaseModel):
    start_date: str
    end_date: str
    epochs: int = 100
    batch_size: int = 32
    use_mock_data: Optional[bool] = False

# 加载最新模型
def load_latest_model():
    checkpoint_dir = CONFIG['model']['checkpoint_dir']
    if not os.path.exists(checkpoint_dir):
        return False
        
    checkpoints = [f for f in os.listdir(checkpoint_dir) if f.endswith('.pt')]
    if not checkpoints:
        return False
        
    latest_checkpoint = max(checkpoints, key=lambda x: os.path.getctime(os.path.join(checkpoint_dir, x)))
    trainer.load_model(os.path.join(checkpoint_dir, latest_checkpoint))
    CONFIG['model']['latest_model'] = latest_checkpoint
    return True

@app.post("/predict")
async def predict(request: PredictionRequest):
    """预测接口"""
    try:
        # 获取当前数据
        current_data = db_loader.load_universe_data(
            start_date=request.current_time,
            end_date=request.current_time,
            user_id=request.user_id
        )
        
        if not current_data:
            raise HTTPException(status_code=404, detail="No data found for the specified parameters")
            
        # 预测
        prediction = trainer.predict(current_data[0])
        
        # 格式化结果
        result = {
            'predicted_location': {
                'latitude': float(prediction[0, 0]),
                'longitude': float(prediction[0, 1])
            },
            'predicted_flow': float(prediction[0, 2]),
            'timestamp': datetime.now().isoformat()
        }
        
        return result
        
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

def generate_mock_data(num_samples: int = 100):
    """生成模拟训练数据"""
    mock_data = []
    
    # 生成一些随机位置
    locations = [(random.uniform(39.0, 41.0), random.uniform(116.0, 118.0)) for _ in range(10)]
    
    # 确保数据有序排列，便于时间窗口处理
    current_date = datetime.now()
    dates = []
    for i in range(num_samples):
        # 每小时一个数据点，连续num_samples小时
        sample_date = current_date - timedelta(hours=num_samples-i-1)
        dates.append(sample_date)
    
    # 按时间排序
    dates.sort()
    
    # 为每个用户生成数据
    for user_id in range(1, 6):  # 5个用户
        user_locations = locations.copy()  # 每个用户有自己的位置集合
        
        for i, sample_date in enumerate(dates):
            # 随机选择一个位置，但保持连续性（每4个点才可能改变位置）
            if i % 4 == 0:
                location_idx = random.randint(0, len(user_locations) - 1)
            lat, lon = user_locations[location_idx]
            
            # 添加一些随机性，但保持变化平滑
            lat += random.uniform(-0.02, 0.02)
            lon += random.uniform(-0.02, 0.02)
            user_locations[location_idx] = (lat, lon)  # 更新位置
            
            # 随机生成其他数据，但保持一定的连续性
            if i > 0:
                prev_flow = mock_data[-1]['flow'] if user_id == mock_data[-1]['user_id'] else random.randint(100, 5000)
                flow = max(100, min(5000, prev_flow + random.randint(-500, 500)))
                
                prev_latency = mock_data[-1]['latency'] if user_id == mock_data[-1]['user_id'] else random.randint(10, 500)
                latency = max(10, min(500, prev_latency + random.randint(-50, 50)))
                
                prev_count = mock_data[-1]['count'] if user_id == mock_data[-1]['user_id'] else random.randint(10, 1000)
                count = max(10, min(1000, prev_count + random.randint(-100, 100)))
                
                prev_err_count = mock_data[-1]['err_count'] if user_id == mock_data[-1]['user_id'] else random.randint(0, 50)
                err_count = max(0, min(50, prev_err_count + random.randint(-5, 5)))
            else:
                flow = random.randint(100, 5000)
                latency = random.randint(10, 500)
                count = random.randint(10, 1000)
                err_count = random.randint(0, 50)
            
            date_str = sample_date.strftime('%Y-%m-%d')
            period_id = sample_date.hour
            
            record = {
                'user_id': user_id,
                'latitude': lat,
                'longitude': lon,
                'flow': flow,
                'latency': latency,
                'count': count,
                'err_count': err_count,
                'period_id': period_id,
                'date': date_str,
                'district': f'District {user_id}',
                'city': f'City {user_id % 3 + 1}',
                'ip': f'192.168.{user_id}.{random.randint(1, 255)}'
            }
            mock_data.append(record)
    
    # 排序确保时间顺序
    mock_data.sort(key=lambda x: (x['date'], x['period_id']))
    
    return mock_data

@app.post("/train")
async def train(request: TrainingRequest):
    """训练接口"""
    try:
        print(f"正在准备训练数据，日期范围: {request.start_date} 到 {request.end_date}")
        print(f"是否使用模拟数据: {request.use_mock_data}")
        
        if request.use_mock_data:
            # 使用模拟数据
            print("使用模拟数据进行训练")
            mock_data = generate_mock_data(200)  # 生成200条模拟数据
            print(f"生成了 {len(mock_data)} 条模拟数据")
            
            # 直接使用训练器处理模拟数据
            training_data = trainer.prepare_training_data(
                start_date=request.start_date,
                end_date=request.end_date,
                mock_data=mock_data
            )
        else:
            # 使用真实数据
            training_data = trainer.prepare_training_data(
                start_date=request.start_date,
                end_date=request.end_date
            )
        
        if not training_data:
            return JSONResponse(
                status_code=404,
                content={"detail": "No training data found or data preprocessing failed"}
            )
        
        print(f"成功准备训练数据，样本数量: {len(training_data)}")
        
        # 打印第一个训练样本的形状信息，帮助调试
        if training_data:
            sample = training_data[0]
            print("第一个训练样本:")
            for key, value in sample.items():
                if isinstance(value, torch.Tensor):
                    print(f"  {key}: shape={value.shape}, dtype={value.dtype}")
                else:
                    print(f"  {key}: type={type(value)}")
                    
        # 训练模型
        trainer.train(
            training_data=training_data,
            epochs=request.epochs,
            batch_size=request.batch_size
        )
        
        # 更新最新模型信息
        if load_latest_model():
            print("成功加载新训练的模型")
        else:
            print("警告：无法加载新训练的模型")
        
        return {"message": "Training completed successfully"}
        
    except ValueError as e:
        print(f"训练时发生值错误: {e}")
        return JSONResponse(
            status_code=400,
            content={"detail": f"Value error: {str(e)}"}
        )
    except Exception as e:
        print(f"训练时发生错误: {e}")
        return JSONResponse(
            status_code=500,
            content={"detail": str(e)}
        )

@app.get("/model/status")
async def model_status():
    """获取模型状态"""
    return {
        "latest_model": CONFIG['model']['latest_model'],
        "model_path": CONFIG['model']['checkpoint_dir']
    }

if __name__ == "__main__":
    uvicorn.run(
        "server:app",
        host=CONFIG['server']['host'],
        port=CONFIG['server']['port'],
        reload=True
    )
