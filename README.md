# README

## 步骤

1. 安装 `Oracle InstantClient` 参考 [Oracle InstantClient 配置](#oracle-instantclient-配置)
2. `go mod tidy`

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
