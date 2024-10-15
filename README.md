# 网络用户画像跨时空生成
---
### 一、后端进展情况：
#### 1.1 已完成接口情况：
https://apifox.com/apidoc/shared-ca923973-80f7-491b-be11-834537731b51

#### 1.2 业务逻辑情况：
1. 用户端登录注册；页面访问token验证；
2. 用户空流时分布：根据解包脚本对空时分布表Universe的CRUD；
3. 简单的用户评分机制，及每日均分接口(后续需要更近评分选项等内容)；
4. 基站空流时分布：完善基站表,获取基站全面统计信息；

#### 1.3 其他说明：
1. 为保证隐私安全，config库下的变量未上传到仓库；