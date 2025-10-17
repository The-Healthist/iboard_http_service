# 设备打印机接口文档

**版本**: 1.2.0  
**基础路径**: `http://your-domain:10031`  
**认证**: Device JWT Token (Header: `Authorization: Bearer <token>`)

---

## 1. 打印机健康检查接口

### 接口信息
- **URL**: `/api/device/client/printers/health`
- **方法**: `POST`
- **认证**: Device JWT Token
- **功能**: 上报香橙派服务状态和打印机列表，自动同步数据库（新增/删除/更新）

### 请求参数

```json
{
  "orange_pi": {
    "ip": "192.168.50.173",           // 必填（status为online/offline时）
    "port": 8080,                     // 必填（status为online/offline时）
    "status": "online",               // 必填，可选值: online, offline, not_configured
    "response_time": 45,              // 可选，响应时间（毫秒）
    "reason": "",                     // 可选，离线原因
    "error_code": ""                  // 可选，错误代码
  },
  "printers": [
    {
      "ip_address": "192.168.50.139", // 必填，打印机IP，唯一标识
      "status": "online",             // 必填，可选值: online, offline
      "display_name": "HP LaserJet",  // 可选，显示名称
      "name": "HP_Printer_139",       // 可选，CUPS打印机名称
      "state": "idle",                // 可选，打印机状态: idle, processing, stopped
      "uri": "ipp://192.168.50.139:631/ipp/print", // 可选，打印机URI
      "reason": ""                    // 可选，离线原因
    }
  ]
}
```

### 响应示例

#### 场景1：香橙派在线，同步成功
```json
{
  "success": true,
  "message": "Printers synchronized successfully",
  "orange_pi_status": "online",
  "summary": {
    "added": 2,
    "updated": 1,
    "deleted": 0,
    "unchanged": 1,
    "total": 4,
    "printers_status": {
      "online": 3,
      "offline": 1
    }
  }
}
```

#### 场景2：香橙派离线
```json
{
  "success": true,
  "message": "Orange Pi service is offline, printer statuses preserved",
  "orange_pi_status": "offline",
  "orange_pi_reason": "Connection timeout after 5s",
  "summary": {
    "added": 0,
    "updated": 0,
    "deleted": 0,
    "unchanged": 4,
    "total": 4
  },
  "note": "Existing printer records in database were not modified"
}
```

#### 场景3：香橙派未配置
```json
{
  "success": true,
  "message": "Orange Pi not configured, no printer sync performed",
  "orange_pi_status": "not_configured",
  "orange_pi_reason": "Orange Pi IP not set in device settings",
  "summary": {
    "added": 0,
    "updated": 0,
    "deleted": 0,
    "total": 0
  }
}
```

---

## 2. 打印回调接口

### 接口信息
- **URL**: `/api/device/client/printers/callback`
- **方法**: `POST`
- **认证**: Device JWT Token
- **功能**: 打印任务完成后上报结果，只更新状态，不新增/删除

### 请求参数

```json
{
  "printers": [
    {
      "ip_address": "192.168.50.139",  // 必填，打印机IP
      "status": "online",              // 必填，可选值: online, offline
      "reason": ""                     // 可选，失败原因
    }
  ]
}
```

### 响应示例

```json
{
  "success": true,
  "message": "Printer statuses updated successfully",
  "summary": {
    "updated": 2,
    "notFound": 0,
    "errors": 0
  }
}
```

---

## 3. 获取设备信息（包含OrangePi和打印机列表）

### 接口信息
- **URL**: `/api/admin/device/{id}`
- **方法**: `GET`
- **认证**: Admin JWT Token
- **功能**: 获取设备详细信息，包含OrangePi状态和打印机列表

### 响应示例

```json
{
  "message": "Get device success",
  "data": {
    "id": 32,
    "deviceId": "DEVICE_64B54E97",
    "buildingId": 2,
    "building": {
      "id": 2,
      "name": "仁英大厦"
    },
    "orangePi": {
      "ip": "192.168.50.173",
      "port": 8080,
      "status": "online",
      "response_time": 45,
      "reason": null,
      "error_code": null,
      "printers": [
        {
          "id": 5,
          "deviceId": 32,
          "display_name": "HP LaserJet P1108",
          "ip_address": "192.168.50.139",
          "name": "HP_Printer_139",
          "state": "idle",
          "uri": "ipp://192.168.50.139:631/ipp/print",
          "status": "online",
          "reason": ""
        },
        {
          "id": 6,
          "deviceId": 32,
          "display_name": "Canon LBP2900",
          "ip_address": "192.168.50.146",
          "name": "Canon_Printer_146",
          "state": "idle",
          "uri": "ipp://192.168.50.146:631/ipp/print",
          "status": "offline",
          "reason": "media-empty-report"
        }
      ]
    },
    "settings": {
      "arrearageUpdateDuration": 5,
      "noticeUpdateDuration": 10,
      "advertisementUpdateDuration": 15,
      "printPassWord": "1090119"
    },
    "status": "active"
  }
}
```

---

## 错误响应

### 401 - 未授权
```json
{
  "error": "unauthorized",
  "message": "no token found"
}
```

### 400 - 请求参数错误
```json
{
  "error": "invalid_request",
  "message": "orange_pi field is required"
}
```

### 422 - 验证错误
```json
{
  "error": "invalid_status",
  "message": "orange_pi.status must be one of: online, offline, not_configured"
}
```

---

## 字段说明

### OrangePi 状态枚举
- `online`: 香橙派服务在线
- `offline`: 香橙派服务离线
- `not_configured`: 香橙派未配置

### 打印机状态枚举
- `online`: 打印机在线可用
- `offline`: 打印机离线或故障

### 打印机 State 枚举
- `idle`: 空闲
- `processing`: 处理中
- `stopped`: 已停止

### 常见错误代码
| 错误代码 | 说明 |
|---------|------|
| `CONNECTION_TIMEOUT` | 连接超时 |
| `CONNECTION_REFUSED` | 连接被拒绝 |
| `NETWORK_UNREACHABLE` | 网络不可达 |
| `SERVICE_UNAVAILABLE` | 服务不可用 |

### 常见打印机离线原因
| Reason | 说明 |
|--------|------|
| `media-empty-report` | 纸张可能为空 |
| `media-empty` | 纸张已用完 |
| `media-jam` | 打印机卡纸 |
| `marker-supply-low` | 墨盒/碳粉不足 |
| `marker-supply-empty` | 墨盒/碳粉已用完 |
| `door-open` | 打印机门未关闭 |
| `offline` | 打印机离线 |
