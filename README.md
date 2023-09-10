# README

## 使用步骤

1. 安装 `Oracle InstantClient` 参考 [Oracle InstantClient 配置](#oracle-instantclient-配置)
2. `go mod tidy`
3. `go build`
4. `cd configs && cp config.example.yaml config.yaml`
5. 修改 `configs/config.yaml`
6. 执行 `oracletypeconverter` 或者 `oracletypeconverter.exe`

## 数据类型映射关系

| Oracle数据类型 | go-oci8映射 | goracle映射 | Go数据类型 |
| ------------- | ----------- | ----------- | ---------- |
| VARCHAR2(size [BYTE/CHAR]) | string | string | string |
| NVARCHAR2(size) | string | string | string |
| NUMBER [(p [, s])] | float64 / int64 | float64 / int64 | float64 / int64 |
| FLOAT [(p)] | float64 | float64 | float64 |
| LONG | string | string | string |
| DATE | time.Time | time.Time | time.Time |
| BINARY_FLOAT | float32 | float32 | float32 |
| BINARY_DOUBLE | float64 | float64 | float64 |
| TIMESTAMP [(fractional_seconds_precision)] | time.Time | time.Time | time.Time |
| TIMESTAMP [(fractional_seconds_precision)] WITH TIME ZONE | time.Time | time.Time | time.Time |
| TIMESTAMP [(fractional_seconds_precision)] WITH LOCAL TIME ZONE | time.Time | time.Time | time.Time |
| INTERVAL YEAR [(year_precision)] TO MONTH | Not supported | Not supported | Not supported |
| INTERVAL DAY [(day_precision)] TO SECOND [(fractional_seconds_precision)] | Not supported | Not supported | Not supported |
| RAW(size) | []byte | []byte | []byte |
| LONG RAW | []byte | []byte | []byte |
| ROWID | string | string | string |
| UROWID [(size)] | string | string | string |
| CHAR [(size [BYTE/CHAR])] | string | string | string |
| NCHAR [(size)] | string | string | string |
| CLOB | string | string | string |
| NCLOB | string | string | string |
| BLOB | []byte | []byte | []byte |
| BFILE | Not supported | Not supported | Not supported |

## Oracle InstantClient 配置

### Windows

+ 下载 <https://raw.githubusercontent.com/tload/binaries/main/instantclient/instantclient-basiclite-windows.x64-19.18.0.0.0dbru.zip>
+ 解压到 C:\oracle\instantclient_19_18
+ 配置 `C:\oracle\instantclient_19_18` 到系统的 Path

### Linux

```bash
cd /opt

curl -O https://raw.githubusercontent.com/tload/binaries/main/instantclient/instantclient-basiclite-linux.x64-19.18.0.0.0dbru.zip

unzip instantclient-basiclite-linux.x64-19.18.0.0.0dbru.zip

echo "/opt/instantclient_19_18" > /etc/ld.so.conf.d/oracle-instantclient.conf
ldconfig
```

### macOS

+ 下载并打开 <https://raw.githubusercontent.com/tload/binaries/main/instantclient/instantclient-basiclite-macos.x64-19.8.0.0.0dbru.dmg> 文件，将自动挂载
+ 命令行输入 `/Volumes/instantclient-basiclite-macos.x64-19.8.0.0.0dbru/install_ic.sh`
+ 复制 `/Users/${USER}/Downloads/instantclient_19_8` 到 `/opt/oracle`，执行命令 `sudo mkdir -p /opt/oracle && sudo cp -rf /Users/${USER}/Downloads/instantclient_19_8 /opt/oracle`
+ 链接到 `/usr/local/lib/`，执行 `ln -s /opt/oracle/instantclient_19_8/libclntsh.dylib /usr/local/lib/`

## 一起共建

如果实现有不正确的地方，欢迎讨论、提交PR
