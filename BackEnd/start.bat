@echo off
echo Starting Internet User Portrait Generate System...

:: 设置工作目录为BackEnd
cd /d %~dp0

:: 1. 启动Python ML服务
echo Starting ML Service...
pushd ml_service
start cmd /k "python -m venv venv && venv\Scripts\activate && pip install -r requirements.txt && set PYTHONPATH=%cd%\.. && python server.py"
popd

:: 等待ML服务启动
timeout /t 10

:: 2. 启动Go后端服务
echo Starting Go Backend Service...
start cmd /k "go run routers/app.go"

echo All services started successfully!
echo ML Service running on http://localhost:8000
echo Backend Service running on http://localhost:5000
echo.
echo Press any key to exit...
pause > nul 