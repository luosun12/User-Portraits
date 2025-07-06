# 设置请求参数
$startDate = (Get-Date).AddDays(-30).ToString("yyyy-MM-dd")
$endDate = (Get-Date).ToString("yyyy-MM-dd")

# 是否使用模拟数据
$useMockData = $false

$body = @{
    start_date = $startDate
    end_date = $endDate
    epochs = 10  # 降低训练轮数，方便测试
    batch_size = 64
    use_mock_data = $useMockData
} | ConvertTo-Json

Write-Host "正在触发模型训练..."
Write-Host "训练数据范围: $startDate 到 $endDate"
Write-Host "参数: epochs=10, batch_size=32, use_mock_data=$useMockData"

# ML服务训练接口
$mlServiceUrl = "http://localhost:8000/train"
Write-Host "`n发送请求到: $mlServiceUrl"

try {
    # 发送POST请求到ML服务训练接口
    $response = Invoke-RestMethod -Uri $mlServiceUrl -Method Post -Body $body -ContentType "application/json"

    # 显示响应结果
    Write-Host "`n训练结果:"
    $response | ConvertTo-Json
} 
catch {
    Write-Host "`n错误: 无法连接到ML服务，请确保服务已启动" -ForegroundColor Red
    Write-Host "错误详情: $_" -ForegroundColor Red
}
