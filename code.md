# iBoard 后端接口设计文档
## 架构设计
### 分层架构
系统采用经典的四层架构设计：
- Controller 层：处理 HTTP 请求响应
- Service 层：实现业务逻辑
- Repository 层：数据访问层
- Model 层：数据模型定义
### 每层职责
#### Controller 层
- 处理 HTTP 请求和响应
- 参数校验和绑定
- 调用 Service 层处理业务逻辑
- 统一响应格式处理
- 错误处理和状态码管理
#### Service 层
- 实现核心业务逻辑
- 调用 Repository 层进行数据操作
- 数据组装和转换
- 事务管理
- 业务规则验证
#### Repository 层
- 数据库操作接口
- 实现数据的 CRUD 操作
- 数据持久化
- 查询优化
- 数据缓存处理
#### Model 层
- 定义数据结构
- 定义表关系
- 字段验证规则
- 模型关联关系
## 模块划分
系统分为基础模块(base)和关系模块(relationship)两大类：

## SuperAdmin 功能模块
### 1. Super Admin Management(超级管理员账户管理)
  - add(email,password)
  - login(emailmpassword)
  - updateP(id,newPassword)
  - get(pageSize,pageNum)
  - getOne(id)
  - delete(ids[])
### 2. Building Admin Management(楼宇管理员账户管理)
  - add(email,password,status)
  - update(id,!password,!status)
  - delete(ids[])
  - get(pageSize,pageNum)
  - getOne(id)
### 3. Building Management(楼宇信息管理)
  - add(name,ismartId,password,!remark)
  - delete(ids[])
  - update(name,!ismartId,!password,!remark)
  - get(pageSize,pageNum)
  - getOne(id)
### 4. Advertisement Management(广告管理)
  - add(title,!description,type,status,duration,startTime,endTime,isPublic,path)
  - update(id,!title,!description,!type,!status,!duration,!startTime,!endTime,!isPublic,!path)
  - delete(ids[])
  - get(pageSize,pageNum)
  - getOne(id)
### 5. Notice Management(通知管理)
  - add(title,!description,type,status,startTime,endTime,isPublic,path,fileTy)
  - addMany([{title,type,status,startTime,endTime,isPublic,path,duration}])
  - update(id,!title,!type,!status,!startTime,!endTime,!isPublic,!path)
  - delete(ids[])
  - get(pageSize,pageNum)
  - getOne(id)
### 6. File Management(文件管理)
  - getUploadParams(fileName)
  - uploadCallback(fileName,size)
  - delete(ids[])
  - get(pageSize,pageNum)
  - getOne(id)
### 7. Building-Advertisement Relationship(楼宇广告关联管理)
  - bind(buildingId,advertisementId)(advertisementIds[],buildingIds[])
  - unbind(buildingId,advertisementId)(advertisementId,buildingIds[])
  - getBuildingAdvertisements(buildingId)
  - getAdvertisementBuildings(advertisementId)
### 8. Building-Notice Relationship(楼宇通知关联管理)
  - bind(buildingId,noticeId)(noticeIds[],buildingIds[])
  - unbind(buildingIds[],noticeId),
  - getBuildingNotices(buildingId)  
  - getNoticeBuildings(noticeId)
### 9. Building-Admin Relationship(楼宇管理员关联管理)
  - bind(buildingId,adminId)
  - unbind(buildingId,adminId)
  - getBuildingAdmins(buildingId)
  - getAdminBuildings(adminId)


## BuildingAdmin 功能模块
## BuildingAdmin login 功能为下面的接口添加token

## building 功能模块
  - login(ismartId,password)
  - get_advertisements_building
  根据
  - get_notices_building 
