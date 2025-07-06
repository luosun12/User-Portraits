import pymysql
from typing import List, Dict, Optional
import pandas as pd
from datetime import datetime, timedelta

class DataLoader:
    def __init__(self, host: str, user: str, password: str, database: str):
        """初始化数据库连接
        
        Args:
            host: 数据库主机地址
            user: 数据库用户名
            password: 数据库密码
            database: 数据库名称
        """
        self.connection = pymysql.connect(
            host=host,
            user=user,
            password=password,
            database=database,
            charset='utf8mb4',
            cursorclass=pymysql.cursors.DictCursor
        )
        
    def load_universe_data(self, start_date: str, end_date: str, user_id: Optional[int] = None) -> List[Dict]:
        """加载Universe表数据
        
        Args:
            start_date: 开始日期 (YYYY-MM-DD)
            end_date: 结束日期 (YYYY-MM-DD)
            user_id: 可选的用户ID过滤
            
        Returns:
            List[Dict]: Universe表记录列表
        """
        with self.connection.cursor() as cursor:
            query = """
                SELECT *
                FROM universe1 
                WHERE date BETWEEN %s AND %s
            """
            params = [start_date, end_date]
            
            if user_id is not None:
                query += " AND user_id = %s"
                params.append(user_id)
                
            cursor.execute(query, params)
            return cursor.fetchall()
            
    def load_base_station_data(self, station_id: int, start_date: str, end_date: str) -> List[Dict]:
        """加载基站数据
        
        Args:
            station_id: 基站ID
            start_date: 开始日期
            end_date: 结束日期
            
        Returns:
            List[Dict]: 基站记录列表
        """
        table_name = f"base_station{station_id}"
        with self.connection.cursor() as cursor:
            query = f"""
                SELECT conn_count, err_count, date, period_id, 
                       total_flow, ave_latency, loss_rate
                FROM {table_name}
                WHERE date BETWEEN %s AND %s
            """
            cursor.execute(query, [start_date, end_date])
            return cursor.fetchall()
            
    def load_recent_data(self, hours: int = 24) -> Dict[str, List[Dict]]:
        """加载最近n小时的数据
        
        Args:
            hours: 需要加载的小时数
            
        Returns:
            Dict: 包含Universe和BaseStation数据的字典
        """
        end_date = datetime.now()
        start_date = end_date - timedelta(hours=hours)
        
        # 加载Universe数据
        universe_data = self.load_universe_data(
            start_date.strftime('%Y-%m-%d'),
            end_date.strftime('%Y-%m-%d')
        )
        
        # 加载所有基站数据
        station_data = {}
        for station_id in range(1, 5):  # 假设有4个基站
            station_data[f'station_{station_id}'] = self.load_base_station_data(
                station_id,
                start_date.strftime('%Y-%m-%d'),
                end_date.strftime('%Y-%m-%d')
            )
            
        return {
            'universe': universe_data,
            'base_stations': station_data
        }
        
    def close(self):
        """关闭数据库连接"""
        self.connection.close()