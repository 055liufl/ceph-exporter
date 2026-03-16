# Samba 服务中文操作指南

> **版本**: 1.0
> **最后更新**: 2026-03-11
> **适用系统**: CentOS 7 / RHEL 7 / Ubuntu / Debian

---

## 📖 目录

1. [Samba 服务概述](#samba-服务概述)
2. [安装与配置](#安装与配置)
3. [用户管理](#用户管理)
4. [共享目录配置](#共享目录配置)
5. [访问控制](#访问控制)
6. [客户端连接](#客户端连接)
7. [常用操作](#常用操作)
8. [安全配置](#安全配置)
9. [性能优化](#性能优化)
10. [故障排查](#故障排查)
11. [最佳实践](#最佳实践)
12. [命令参考](#命令参考)

---

## Samba 服务概述

### 什么是 Samba？

Samba 是一个开源软件套件，实现了 SMB/CIFS 协议，使 Linux/Unix 系统能够与 Windows 系统进行文件和打印机共享。

### 核心功能

- ✅ **文件共享**: 在 Linux 和 Windows 之间共享文件
- ✅ **打印机共享**: 共享网络打印机
- ✅ **身份验证**: 支持多种身份验证方式
- ✅ **域控制器**: 可作为 Windows 域控制器
- ✅ **跨平台**: 支持 Windows、Linux、macOS 等系统

### 主要组件

| 组件 | 说明 | 用途 |
|------|------|------|
| **smbd** | Samba 守护进程 | 提供文件和打印服务 |
| **nmbd** | NetBIOS 名称服务器 | 提供 NetBIOS 名称解析 |
| **winbindd** | Windows 域集成服务 | 与 Windows 域集成 |
| **smbclient** | Samba 客户端工具 | 访问 SMB 共享 |
| **smbpasswd** | 密码管理工具 | 管理 Samba 用户密码 |
| **testparm** | 配置测试工具 | 验证配置文件语法 |

### 端口说明

| 端口 | 协议 | 用途 |
|------|------|------|
| 139 | TCP | NetBIOS 会话服务 |
| 445 | TCP | SMB over TCP（直接托管） |
| 137 | UDP | NetBIOS 名称服务 |
| 138 | UDP | NetBIOS 数据报服务 |

---

## 安装与配置

### CentOS 7 / RHEL 7 安装

#### 1. 安装 Samba

```bash
# 安装 Samba 服务器
sudo yum install -y samba samba-client samba-common

# 查看安装版本
smbd --version
```

#### 2. 启动服务

```bash
# 启动 Samba 服务
sudo systemctl start smb
sudo systemctl start nmb

# 设置开机自启
sudo systemctl enable smb
sudo systemctl enable nmb

# 查看服务状态
sudo systemctl status smb
sudo systemctl status nmb
```

#### 3. 配置防火墙

```bash
# 开放 Samba 端口
sudo firewall-cmd --permanent --add-service=samba
sudo firewall-cmd --reload

# 或手动开放端口
sudo firewall-cmd --permanent --add-port=139/tcp
sudo firewall-cmd --permanent --add-port=445/tcp
sudo firewall-cmd --permanent --add-port=137/udp
sudo firewall-cmd --permanent --add-port=138/udp
sudo firewall-cmd --reload

# 验证防火墙规则
sudo firewall-cmd --list-all
```

#### 4. 配置 SELinux

```bash
# 查看 SELinux 状态
getenforce

# 如果启用了 SELinux，需要设置布尔值
sudo setsebool -P samba_enable_home_dirs on
sudo setsebool -P samba_export_all_rw on

# 查看 Samba 相关的 SELinux 布尔值
getsebool -a | grep samba
```

### Ubuntu / Debian 安装

```bash
# 更新软件包列表
sudo apt update

# 安装 Samba
sudo apt install -y samba samba-common-bin

# 启动服务
sudo systemctl start smbd
sudo systemctl start nmbd

# 设置开机自启
sudo systemctl enable smbd
sudo systemctl enable nmbd

# 查看服务状态
sudo systemctl status smbd
sudo systemctl status nmbd
```

### 主配置文件

**配置文件位置**: `/etc/samba/smb.conf`

#### 配置文件结构

```ini
# Samba 配置文件分为两部分：
# 1. [global] - 全局配置
# 2. [共享名] - 共享配置

[global]
    # 全局配置项

[共享名称]
    # 共享配置项
```

#### 基本全局配置

```ini
[global]
    # 工作组名称（Windows 网络中的工作组）
    workgroup = WORKGROUP

    # 服务器描述
    server string = Samba Server %v

    # NetBIOS 名称
    netbios name = SAMBA-SERVER

    # 安全级别
    security = user

    # 密码后端
    passdb backend = tdbsam

    # 日志文件
    log file = /var/log/samba/log.%m

    # 最大日志大小（KB）
    max log size = 50

    # 接口绑定
    interfaces = lo eth0
    bind interfaces only = yes

    # 主机访问控制
    hosts allow = 192.168.1. 127.
    hosts deny = 0.0.0.0/0
```

#### 验证配置

```bash
# 验证配置文件语法
sudo testparm

# 验证配置并显示详细信息
sudo testparm -v

# 验证特定共享
sudo testparm -s --section-name=共享名称
```

---

## 用户管理

### Samba 用户概念

- Samba 用户必须是 Linux 系统用户
- Samba 有独立的密码数据库
- 需要同时创建 Linux 用户和 Samba 用户

### 创建 Samba 用户

#### 方式 1: 创建新用户

```bash
# 1. 创建 Linux 系统用户（不允许登录）
sudo useradd -M -s /sbin/nologin sambauser

# 2. 创建 Samba 用户并设置密码
sudo smbpasswd -a sambauser
# 输入密码（两次）

# 3. 启用 Samba 用户
sudo smbpasswd -e sambauser
```

#### 方式 2: 为现有用户添加 Samba 访问

```bash
# 为现有 Linux 用户添加 Samba 密码
sudo smbpasswd -a existing_user

# 启用用户
sudo smbpasswd -e existing_user
```

### 管理 Samba 用户

```bash
# 列出所有 Samba 用户
sudo pdbedit -L

# 查看用户详细信息
sudo pdbedit -L -v username

# 修改用户密码
sudo smbpasswd username

# 禁用用户
sudo smbpasswd -d username

# 启用用户
sudo smbpasswd -e username

# 删除 Samba 用户
sudo smbpasswd -x username

# 删除 Linux 用户（如果需要）
sudo userdel username
```

### 用户组管理

```bash
# 创建用户组
sudo groupadd smbgroup

# 将用户添加到组
sudo usermod -aG smbgroup username

# 查看用户所属组
groups username

# 查看组成员
getent group smbgroup
```

---

## 共享目录配置

### 基本共享配置

#### 1. 公共共享（匿名访问）

```ini
[Public]
    # 共享描述
    comment = Public Share

    # 共享路径
    path = /srv/samba/public

    # 可浏览
    browseable = yes

    # 可写
    writable = yes

    # 允许访客访问
    guest ok = yes

    # 创建文件的权限掩码
    create mask = 0755

    # 创建目录的权限掩码
    directory mask = 0755
```

**创建共享目录**:
```bash
# 创建目录
sudo mkdir -p /srv/samba/public

# 设置权限
sudo chmod 777 /srv/samba/public

# 设置 SELinux 上下文（如果启用）
sudo chcon -t samba_share_t /srv/samba/public
```

#### 2. 私有共享（需要认证）

```ini
[Private]
    comment = Private Share
    path = /srv/samba/private
    browseable = yes
    writable = yes

    # 不允许访客
    guest ok = no

    # 有效用户
    valid users = user1, user2, @smbgroup

    # 只读用户
    read list = user3

    # 可写用户
    write list = user1, user2

    create mask = 0660
    directory mask = 0770
```

**创建共享目录**:
```bash
# 创建目录
sudo mkdir -p /srv/samba/private

# 设置所有者
sudo chown -R root:smbgroup /srv/samba/private

# 设置权限
sudo chmod 770 /srv/samba/private

# 设置 SELinux 上下文
sudo chcon -t samba_share_t /srv/samba/private
```

#### 3. 用户主目录共享

```ini
[homes]
    comment = Home Directories
    browseable = no
    writable = yes
    valid users = %S
    create mask = 0700
    directory mask = 0700
```

#### 4. 只读共享

```ini
[ReadOnly]
    comment = Read Only Share
    path = /srv/samba/readonly
    browseable = yes

    # 只读
    read only = yes

    # 或使用 writable = no
    # writable = no

    guest ok = yes
```

### 高级共享配置

#### 1. 回收站功能

```ini
[Share]
    comment = Share with Recycle Bin
    path = /srv/samba/share
    writable = yes

    # 启用 VFS 回收站模块
    vfs objects = recycle

    # 回收站配置
    recycle:repository = .recycle
    recycle:keeptree = yes
    recycle:versions = yes
    recycle:touch = yes
    recycle:exclude = *.tmp, *.temp
    recycle:exclude_dir = /tmp, /cache
```

#### 2. 审计日志

```ini
[Share]
    comment = Share with Audit
    path = /srv/samba/share
    writable = yes

    # 启用审计模块
    vfs objects = full_audit

    # 审计配置
    full_audit:prefix = %u|%I|%m|%S
    full_audit:success = mkdir rmdir read write rename
    full_audit:failure = none
    full_audit:facility = local5
    full_audit:priority = notice
```

#### 3. 隐藏文件

```ini
[Share]
    comment = Share with Hidden Files
    path = /srv/samba/share
    writable = yes

    # 隐藏以点开头的文件
    hide dot files = yes

    # 隐藏特定文件
    veto files = /*.exe/*.com/*.bat/

    # 删除被否决的文件
    delete veto files = yes
```

### 应用配置更改

```bash
# 验证配置
sudo testparm

# 重新加载配置（不中断连接）
sudo smbcontrol all reload-config

# 或重启服务
sudo systemctl restart smb
sudo systemctl restart nmb
```

---

## 访问控制

### 基于用户的访问控制

```ini
[Share]
    path = /srv/samba/share

    # 允许访问的用户
    valid users = user1, user2, @group1

    # 拒绝访问的用户
    invalid users = baduser, @badgroup

    # 只读用户列表
    read list = user3, @readonly_group

    # 可写用户列表
    write list = user1, user2, @write_group

    # 管理员用户（完全控制）
    admin users = admin, @admin_group
```

### 基于主机的访问控制

```ini
[Share]
    path = /srv/samba/share

    # 允许访问的主机
    hosts allow = 192.168.1. 10.0.0. localhost

    # 拒绝访问的主机
    hosts deny = 0.0.0.0/0
```

### 基于 IP 范围的访问控制

```ini
[Share]
    path = /srv/samba/share

    # CIDR 表示法
    hosts allow = 192.168.1.0/24 10.0.0.0/8

    # 排除特定 IP
    hosts deny = 192.168.1.100
```

### 文件权限控制

```ini
[Share]
    path = /srv/samba/share
    writable = yes

    # 创建文件的权限掩码
    create mask = 0664

    # 强制创建文件的权限
    force create mode = 0664

    # 创建目录的权限掩码
    directory mask = 0775

    # 强制创建目录的权限
    force directory mode = 0775

    # 强制用户
    force user = smbuser

    # 强制组
    force group = smbgroup
```

---

## 客户端连接

### Windows 客户端

#### 方式 1: 文件资源管理器

```
1. 打开文件资源管理器
2. 在地址栏输入: \\服务器IP\共享名
   例如: \\192.168.1.100\Public
3. 输入用户名和密码（如果需要）
4. 点击"记住我的凭据"（可选）
```

#### 方式 2: 映射网络驱动器

```
1. 右键点击"此电脑" → "映射网络驱动器"
2. 选择驱动器号（如 Z:）
3. 输入文件夹路径: \\192.168.1.100\Share
4. 勾选"登录时重新连接"
5. 勾选"使用其他凭据连接"（如果需要）
6. 点击"完成"
7. 输入用户名和密码
```

#### 方式 3: 命令行

```cmd
# 查看共享列表
net view \\192.168.1.100

# 连接共享
net use Z: \\192.168.1.100\Share /user:username password

# 断开连接
net use Z: /delete

# 查看已连接的共享
net use
```

### Linux 客户端

#### 方式 1: smbclient 命令行工具

```bash
# 安装 smbclient
sudo yum install -y samba-client  # CentOS
sudo apt install -y smbclient     # Ubuntu

# 列出服务器上的共享
smbclient -L //192.168.1.100 -U username

# 连接到共享
smbclient //192.168.1.100/Share -U username

# 在 smbclient 中的常用命令
smb: \> ls              # 列出文件
smb: \> cd directory    # 切换目录
smb: \> get file        # 下载文件
smb: \> put file        # 上传文件
smb: \> mget *.txt      # 批量下载
smb: \> mput *.txt      # 批量上传
smb: \> quit            # 退出
```

#### 方式 2: 挂载 SMB 共享

```bash
# 安装 cifs-utils
sudo yum install -y cifs-utils  # CentOS
sudo apt install -y cifs-utils  # Ubuntu

# 创建挂载点
sudo mkdir -p /mnt/samba

# 临时挂载
sudo mount -t cifs //192.168.1.100/Share /mnt/samba -o username=user,password=pass

# 或使用凭据文件（更安全）
# 创建凭据文件
sudo vi /root/.smbcredentials
# 内容:
# username=sambauser
# password=password

# 设置权限
sudo chmod 600 /root/.smbcredentials

# 使用凭据文件挂载
sudo mount -t cifs //192.168.1.100/Share /mnt/samba -o credentials=/root/.smbcredentials

# 永久挂载（编辑 /etc/fstab）
sudo vi /etc/fstab
# 添加:
//192.168.1.100/Share  /mnt/samba  cifs  credentials=/root/.smbcredentials,uid=1000,gid=1000  0  0

# 测试 fstab
sudo mount -a

# 卸载
sudo umount /mnt/samba
```

#### 方式 3: 文件管理器（GUI）

**Nautilus (GNOME)**:
```
1. 打开文件管理器
2. 点击"其他位置"
3. 在底部输入: smb://192.168.1.100/Share
4. 输入用户名和密码
```

**Dolphin (KDE)**:
```
1. 打开 Dolphin
2. 在地址栏输入: smb://192.168.1.100/Share
3. 输入用户名和密码
```

### macOS 客户端

#### 方式 1: Finder

```
1. 打开 Finder
2. 按 Command + K
3. 输入服务器地址: smb://192.168.1.100/Share
4. 点击"连接"
5. 输入用户名和密码
```

#### 方式 2: 命令行

```bash
# 挂载 SMB 共享
mkdir ~/samba
mount_smbfs //username:password@192.168.1.100/Share ~/samba

# 卸载
umount ~/samba
```

---

## 常用操作

### 查看 Samba 状态

```bash
# 查看服务状态
sudo systemctl status smb
sudo systemctl status nmb

# 查看 Samba 进程
ps aux | grep smbd
ps aux | grep nmbd

# 查看监听端口
sudo netstat -tlnp | grep smbd
sudo ss -tlnp | grep smbd

# 查看当前连接
sudo smbstatus

# 查看连接的用户
sudo smbstatus -b

# 查看共享列表
sudo smbstatus -S

# 查看锁定的文件
sudo smbstatus -L
```

### 管理 Samba 服务

```bash
# 启动服务
sudo systemctl start smb nmb

# 停止服务
sudo systemctl stop smb nmb

# 重启服务
sudo systemctl restart smb nmb

# 重新加载配置
sudo systemctl reload smb

# 查看服务日志
sudo journalctl -u smb -f
sudo journalctl -u nmb -f

# 查看 Samba 日志文件
sudo tail -f /var/log/samba/log.smbd
sudo tail -f /var/log/samba/log.nmbd
```

### 测试和诊断

```bash
# 验证配置文件
sudo testparm

# 测试用户认证
smbclient -L localhost -U username

# 测试共享访问
smbclient //localhost/Share -U username

# 查看 NetBIOS 名称
nmblookup -A 192.168.1.100

# 查看工作组中的主机
nmblookup -M WORKGROUP

# 测试网络连接
ping 192.168.1.100
telnet 192.168.1.100 445
```

### 备份和恢复

```bash
# 备份配置文件
sudo cp /etc/samba/smb.conf /etc/samba/smb.conf.backup.$(date +%Y%m%d)

# 备份用户数据库
sudo tdbbackup /var/lib/samba/private/passdb.tdb

# 备份共享数据
sudo tar -czf /backup/samba-data-$(date +%Y%m%d).tar.gz /srv/samba/

# 恢复配置文件
sudo cp /etc/samba/smb.conf.backup.20260311 /etc/samba/smb.conf

# 恢复用户数据库
sudo cp /var/lib/samba/private/passdb.tdb.bak /var/lib/samba/private/passdb.tdb

# 重启服务
sudo systemctl restart smb nmb
```

---

## 安全配置

### 密码策略

```ini
[global]
    # 最小密码长度
    min password length = 8

    # 密码历史记录
    password history = 5

    # 密码复杂度
    check password script = /usr/local/bin/check_password.sh
```

### 加密传输

```ini
[global]
    # 启用 SMB 加密
    smb encrypt = required

    # 或仅对特定共享加密
    [Share]
        smb encrypt = desired
```

### 限制协议版本

```ini
[global]
    # 只允许 SMB2 和 SMB3
    server min protocol = SMB2
    server max protocol = SMB3

    # 禁用 SMB1（不安全）
    server min protocol = SMB2_02
```

### 审计日志

```ini
[global]
    # 启用详细日志
    log level = 2

    # 日志文件大小限制
    max log size = 1000

    # 每个客户端单独的日志文件
    log file = /var/log/samba/%m.log
```

### 防止暴力破解

```bash
# 使用 fail2ban 防止暴力破解
sudo yum install -y fail2ban

# 创建 Samba 过滤器
sudo vi /etc/fail2ban/filter.d/samba.conf
```

内容:
```ini
[Definition]
failregex = ^.*smbd.*: .*authentication for user \[.*\] -> \[<HOST>\] FAILED.*$
ignoreregex =
```

配置 jail:
```bash
sudo vi /etc/fail2ban/jail.local
```

内容:
```ini
[samba]
enabled = true
port = 139,445
filter = samba
logpath = /var/log/samba/log.smbd
maxretry = 3
bantime = 3600
```

启动 fail2ban:
```bash
sudo systemctl start fail2ban
sudo systemctl enable fail2ban
```

---

## 性能优化

### 网络优化

```ini
[global]
    # 套接字选项
    socket options = TCP_NODELAY IPTOS_LOWDELAY SO_RCVBUF=131072 SO_SNDBUF=131072

    # 读取大小
    read raw = yes
    write raw = yes

    # 最大传输大小
    max xmit = 65535

    # 异步 I/O
    aio read size = 16384
    aio write size = 16384
```

### 缓存优化

```ini
[global]
    # 启用 sendfile
    use sendfile = yes

    # 启用内核变更通知
    kernel change notify = yes

    # 启用内核 oplocks
    kernel oplocks = yes
```

### 大文件优化

```ini
[Share]
    # 严格分配
    strict allocate = yes

    # 严格同步
    strict sync = no

    # 同步总是
    sync always = no
```

### 性能监控

```bash
# 查看 Samba 性能统计
sudo smbstatus -p

# 查看网络流量
sudo iftop -i eth0

# 查看磁盘 I/O
sudo iostat -x 1

# 查看系统负载
top
htop
```

---


## 故障排查

### 常见问题

#### 问题 1: 无法连接到 Samba 服务器

**症状**: 客户端无法访问 `\\服务器IP\共享名`

**排查步骤**:

```bash
# 1. 检查服务是否运行
sudo systemctl status smb
sudo systemctl status nmb

# 2. 检查端口是否监听
sudo netstat -tlnp | grep -E '139|445'

# 3. 检查防火墙
sudo firewall-cmd --list-all

# 4. 测试网络连接
ping 服务器IP
telnet 服务器IP 445

# 5. 检查 SELinux
getenforce
sudo ausearch -m avc -ts recent | grep smbd
```

**解决方案**:

```bash
# 启动服务
sudo systemctl start smb nmb

# 开放防火墙端口
sudo firewall-cmd --permanent --add-service=samba
sudo firewall-cmd --reload

# 临时禁用 SELinux 测试
sudo setenforce 0

# 如果禁用 SELinux 后可以访问，设置正确的 SELinux 上下文
sudo setsebool -P samba_enable_home_dirs on
sudo setsebool -P samba_export_all_rw on
```

#### 问题 2: 用户认证失败

**症状**: 输入用户名和密码后提示"登录失败"

**排查步骤**:

```bash
# 1. 检查用户是否存在
sudo pdbedit -L | grep username

# 2. 检查用户是否启用
sudo pdbedit -L -v username | grep "Account Flags"

# 3. 测试用户认证
smbclient -L localhost -U username

# 4. 查看日志
sudo tail -f /var/log/samba/log.smbd
```

**解决方案**:

```bash
# 重置用户密码
sudo smbpasswd username

# 启用用户
sudo smbpasswd -e username

# 如果用户不存在，创建用户
sudo useradd -M -s /sbin/nologin username
sudo smbpasswd -a username
```

#### 问题 3: 权限被拒绝

**症状**: 可以连接但无法读写文件

**排查步骤**:

```bash
# 1. 检查共享配置
sudo testparm -s --section-name=共享名

# 2. 检查目录权限
ls -ld /srv/samba/share

# 3. 检查文件权限
ls -l /srv/samba/share/

# 4. 检查 SELinux 上下文
ls -Z /srv/samba/share
```

**解决方案**:

```bash
# 修改目录权限
sudo chmod 775 /srv/samba/share
sudo chown -R root:smbgroup /srv/samba/share

# 设置 SELinux 上下文
sudo chcon -t samba_share_t /srv/samba/share

# 或永久设置
sudo semanage fcontext -a -t samba_share_t "/srv/samba/share(/.*)?"
sudo restorecon -Rv /srv/samba/share

# 修改共享配置
sudo vi /etc/samba/smb.conf
# 添加或修改:
# writable = yes
# valid users = username

# 重新加载配置
sudo systemctl reload smb
```

#### 问题 4: 共享不可见

**症状**: 在网络邻居中看不到共享

**排查步骤**:

```bash
# 1. 检查 nmbd 服务
sudo systemctl status nmb

# 2. 检查共享配置
sudo testparm -s

# 3. 检查 browseable 设置
grep -A 10 "\[共享名\]" /etc/samba/smb.conf | grep browseable
```

**解决方案**:

```bash
# 启动 nmbd 服务
sudo systemctl start nmb

# 修改共享配置
sudo vi /etc/samba/smb.conf
# 设置:
# browseable = yes

# 重启服务
sudo systemctl restart smb nmb
```

#### 问题 5: 性能缓慢

**症状**: 文件传输速度很慢

**排查步骤**:

```bash
# 1. 检查网络速度
iperf3 -s  # 在服务器上
iperf3 -c 服务器IP  # 在客户端上

# 2. 检查磁盘 I/O
sudo iostat -x 1

# 3. 检查系统负载
top
htop

# 4. 查看 Samba 日志
sudo tail -f /var/log/samba/log.smbd
```

**解决方案**:

```bash
# 优化配置
sudo vi /etc/samba/smb.conf
# 添加:
# socket options = TCP_NODELAY IPTOS_LOWDELAY SO_RCVBUF=131072 SO_SNDBUF=131072
# use sendfile = yes
# read raw = yes
# write raw = yes

# 重启服务
sudo systemctl restart smb
```

#### 问题 6: 配置文件错误

**症状**: 服务无法启动或配置不生效

**排查步骤**:

```bash
# 验证配置文件语法
sudo testparm

# 查看详细错误信息
sudo testparm -v
```

**解决方案**:

```bash
# 恢复备份配置
sudo cp /etc/samba/smb.conf.backup /etc/samba/smb.conf

# 或重新生成默认配置
sudo mv /etc/samba/smb.conf /etc/samba/smb.conf.broken
sudo cp /usr/share/doc/samba-*/smb.conf.example /etc/samba/smb.conf

# 重启服务
sudo systemctl restart smb nmb
```

### 日志分析

#### 日志位置

```bash
# 主日志文件
/var/log/samba/log.smbd      # smbd 日志
/var/log/samba/log.nmbd      # nmbd 日志
/var/log/samba/log.%m        # 每个客户端的日志
```

#### 查看日志

```bash
# 实时查看日志
sudo tail -f /var/log/samba/log.smbd

# 查看最近的错误
sudo grep -i error /var/log/samba/log.smbd | tail -20

# 查看认证失败
sudo grep -i "authentication failed" /var/log/samba/log.smbd

# 查看特定用户的日志
sudo grep "username" /var/log/samba/log.smbd

# 使用 journalctl 查看日志
sudo journalctl -u smb -f
sudo journalctl -u nmb -f
```

#### 增加日志级别

```bash
# 临时增加日志级别
sudo smbcontrol smbd debug 3

# 永久修改（编辑配置文件）
sudo vi /etc/samba/smb.conf
# 添加:
# log level = 3

# 重新加载配置
sudo systemctl reload smb
```

### 诊断工具

```bash
# 1. testparm - 验证配置
sudo testparm

# 2. smbclient - 测试连接
smbclient -L localhost -U username

# 3. smbstatus - 查看状态
sudo smbstatus

# 4. nmblookup - NetBIOS 名称查询
nmblookup -A 192.168.1.100

# 5. rpcclient - RPC 客户端
rpcclient -U username //localhost

# 6. net - 网络管理工具
net rpc info -U username

# 7. pdbedit - 用户数据库管理
sudo pdbedit -L -v

# 8. smbpasswd - 密码管理
sudo smbpasswd -a username
```

---

## 最佳实践

### 安全最佳实践

#### 1. 用户和权限管理

```bash
# 使用专用的 Samba 用户（不允许系统登录）
sudo useradd -M -s /sbin/nologin sambauser

# 使用强密码
sudo smbpasswd -a sambauser
# 输入复杂密码（至少 12 位，包含大小写字母、数字、特殊字符）

# 定期更换密码
sudo smbpasswd sambauser

# 最小权限原则
# 只授予必要的权限，使用 read list 和 write list
```

#### 2. 网络安全

```ini
[global]
    # 限制访问的主机
    hosts allow = 192.168.1.0/24 127.0.0.1
    hosts deny = 0.0.0.0/0

    # 禁用 SMB1（不安全）
    server min protocol = SMB2_02

    # 启用 SMB 加密
    smb encrypt = required

    # 绑定到特定接口
    interfaces = lo eth0
    bind interfaces only = yes
```

#### 3. 审计和监控

```ini
[global]
    # 启用详细日志
    log level = 2
    log file = /var/log/samba/%m.log
    max log size = 1000

[Share]
    # 启用审计
    vfs objects = full_audit
    full_audit:prefix = %u|%I|%m|%S
    full_audit:success = mkdir rmdir read write rename
    full_audit:failure = all
    full_audit:facility = local5
    full_audit:priority = notice
```

#### 4. 备份策略

```bash
# 每日备份脚本
#!/bin/bash
BACKUP_DIR="/backup/samba"
DATE=$(date +%Y%m%d)

# 备份配置
cp /etc/samba/smb.conf ${BACKUP_DIR}/smb.conf.${DATE}

# 备份用户数据库
tdbbackup /var/lib/samba/private/passdb.tdb

# 备份共享数据
tar -czf ${BACKUP_DIR}/samba-data-${DATE}.tar.gz /srv/samba/

# 保留最近 7 天的备份
find ${BACKUP_DIR} -name "*.tar.gz" -mtime +7 -delete
```

### 性能最佳实践

#### 1. 网络优化

```ini
[global]
    # 套接字选项
    socket options = TCP_NODELAY IPTOS_LOWDELAY SO_RCVBUF=131072 SO_SNDBUF=131072

    # 启用 sendfile
    use sendfile = yes

    # 最大传输大小
    max xmit = 65535

    # 异步 I/O
    aio read size = 16384
    aio write size = 16384
```

#### 2. 磁盘优化

```bash
# 使用高性能文件系统（ext4, xfs）
sudo mkfs.xfs /dev/sdb1

# 挂载选项优化
sudo vi /etc/fstab
# 添加:
# /dev/sdb1  /srv/samba  xfs  defaults,noatime,nodiratime  0  0

# 禁用不必要的同步
```

```ini
[Share]
    strict sync = no
    sync always = no
```

#### 3. 缓存优化

```ini
[global]
    # 启用内核变更通知
    kernel change notify = yes

    # 启用内核 oplocks
    kernel oplocks = yes

    # 启用 oplocks
    oplocks = yes
    level2 oplocks = yes
```

### 维护最佳实践

#### 1. 定期维护任务

```bash
# 每周维护脚本
#!/bin/bash

# 1. 检查服务状态
systemctl status smb nmb

# 2. 验证配置
testparm -s

# 3. 检查磁盘空间
df -h /srv/samba

# 4. 查看活跃连接
smbstatus

# 5. 检查日志错误
grep -i error /var/log/samba/log.smbd | tail -20

# 6. 备份配置和数据
cp /etc/samba/smb.conf /backup/smb.conf.$(date +%Y%m%d)
tar -czf /backup/samba-data-$(date +%Y%m%d).tar.gz /srv/samba/

# 7. 清理旧日志
find /var/log/samba -name "*.log" -mtime +30 -delete
```

#### 2. 监控指标

```bash
# 监控脚本
#!/bin/bash

# 检查服务状态
if ! systemctl is-active --quiet smb; then
    echo "WARNING: Samba service is not running"
    systemctl start smb
fi

# 检查磁盘使用率
USAGE=$(df -h /srv/samba | awk 'NR==2 {print $5}' | sed 's/%//')
if [ $USAGE -gt 90 ]; then
    echo "WARNING: Disk usage is ${USAGE}%"
fi

# 检查活跃连接数
CONNECTIONS=$(smbstatus -b | wc -l)
echo "Active connections: $CONNECTIONS"

# 检查日志错误
ERRORS=$(grep -c "error" /var/log/samba/log.smbd)
if [ $ERRORS -gt 10 ]; then
    echo "WARNING: Found $ERRORS errors in log"
fi
```

#### 3. 更新和升级

```bash
# 更新前备份
sudo cp /etc/samba/smb.conf /etc/samba/smb.conf.backup
sudo tdbbackup /var/lib/samba/private/passdb.tdb

# 更新 Samba
sudo yum update samba samba-client samba-common  # CentOS
sudo apt update && sudo apt upgrade samba        # Ubuntu

# 验证配置
sudo testparm

# 重启服务
sudo systemctl restart smb nmb

# 验证服务
sudo systemctl status smb nmb
smbclient -L localhost -U username
```

### 文档和记录

#### 1. 配置文档

```bash
# 记录配置更改
sudo vi /etc/samba/CHANGELOG
# 内容:
# 2026-03-11: 添加新共享 /srv/samba/projects
# 2026-03-10: 更新用户 john 的权限
# 2026-03-09: 启用 SMB 加密
```

#### 2. 用户文档

创建用户手册，包含:
- 如何连接到共享
- 用户名和密码管理
- 常见问题解决
- 联系支持的方式

#### 3. 运维文档

记录:
- 服务器配置信息
- 共享列表和权限
- 备份和恢复流程
- 故障排查步骤
- 联系人信息

---

## 命令参考

### Samba 服务管理

```bash
# 启动服务
sudo systemctl start smb
sudo systemctl start nmb

# 停止服务
sudo systemctl stop smb
sudo systemctl stop nmb

# 重启服务
sudo systemctl restart smb
sudo systemctl restart nmb

# 重新加载配置
sudo systemctl reload smb
sudo smbcontrol all reload-config

# 查看服务状态
sudo systemctl status smb
sudo systemctl status nmb

# 设置开机自启
sudo systemctl enable smb
sudo systemctl enable nmb

# 禁用开机自启
sudo systemctl disable smb
sudo systemctl disable nmb
```

### 用户管理命令

```bash
# 添加 Samba 用户
sudo smbpasswd -a username

# 删除 Samba 用户
sudo smbpasswd -x username

# 修改用户密码
sudo smbpasswd username

# 启用用户
sudo smbpasswd -e username

# 禁用用户
sudo smbpasswd -d username

# 列出所有用户
sudo pdbedit -L

# 查看用户详细信息
sudo pdbedit -L -v username

# 删除用户（pdbedit）
sudo pdbedit -x username
```

### 配置管理命令

```bash
# 验证配置文件
sudo testparm

# 验证并显示详细信息
sudo testparm -v

# 验证特定共享
sudo testparm -s --section-name=共享名

# 显示配置摘要
sudo testparm -s

# 检查配置文件语法
sudo testparm --suppress-prompt
```

### 状态查询命令

```bash
# 查看所有连接
sudo smbstatus

# 查看简要信息
sudo smbstatus -b

# 查看共享列表
sudo smbstatus -S

# 查看锁定的文件
sudo smbstatus -L

# 查看进程信息
sudo smbstatus -p

# 查看用户信息
sudo smbstatus -u username
```

### 客户端命令

```bash
# 列出服务器共享
smbclient -L //服务器IP -U username

# 连接到共享
smbclient //服务器IP/共享名 -U username

# 匿名连接
smbclient -L //服务器IP -N

# 执行命令
smbclient //服务器IP/共享名 -U username -c "ls"

# 挂载共享
sudo mount -t cifs //服务器IP/共享名 /mnt/samba -o username=user,password=pass

# 卸载共享
sudo umount /mnt/samba
```

### 网络诊断命令

```bash
# NetBIOS 名称查询
nmblookup -A 192.168.1.100

# 查询工作组主浏览器
nmblookup -M WORKGROUP

# 查询特定名称
nmblookup 服务器名称

# 测试端口连接
telnet 192.168.1.100 445
nc -zv 192.168.1.100 445

# 查看监听端口
sudo netstat -tlnp | grep -E '139|445'
sudo ss -tlnp | grep -E '139|445'
```

### 日志管理命令

```bash
# 查看实时日志
sudo tail -f /var/log/samba/log.smbd
sudo tail -f /var/log/samba/log.nmbd

# 查看系统日志
sudo journalctl -u smb -f
sudo journalctl -u nmb -f

# 查看最近的日志
sudo journalctl -u smb --since "1 hour ago"

# 查看错误日志
sudo grep -i error /var/log/samba/log.smbd

# 设置日志级别
sudo smbcontrol smbd debug 3
```

### 防火墙命令

```bash
# 开放 Samba 服务
sudo firewall-cmd --permanent --add-service=samba
sudo firewall-cmd --reload

# 开放特定端口
sudo firewall-cmd --permanent --add-port=139/tcp
sudo firewall-cmd --permanent --add-port=445/tcp
sudo firewall-cmd --permanent --add-port=137/udp
sudo firewall-cmd --permanent --add-port=138/udp
sudo firewall-cmd --reload

# 查看防火墙规则
sudo firewall-cmd --list-all

# 删除规则
sudo firewall-cmd --permanent --remove-service=samba
sudo firewall-cmd --reload
```

### SELinux 命令

```bash
# 查看 SELinux 状态
getenforce

# 临时禁用 SELinux
sudo setenforce 0

# 永久禁用 SELinux
sudo vi /etc/selinux/config
# 设置: SELINUX=disabled

# 设置 Samba 布尔值
sudo setsebool -P samba_enable_home_dirs on
sudo setsebool -P samba_export_all_rw on

# 查看 Samba 布尔值
getsebool -a | grep samba

# 设置文件上下文
sudo chcon -t samba_share_t /srv/samba/share

# 永久设置上下文
sudo semanage fcontext -a -t samba_share_t "/srv/samba/share(/.*)?"
sudo restorecon -Rv /srv/samba/share

# 查看 SELinux 日志
sudo ausearch -m avc -ts recent | grep smbd
```

---

## 附录

### A. 配置文件模板

#### 基本配置模板

```ini
[global]
    workgroup = WORKGROUP
    server string = Samba Server %v
    netbios name = SAMBA-SERVER
    security = user
    passdb backend = tdbsam
    log file = /var/log/samba/log.%m
    max log size = 50
    interfaces = lo eth0
    bind interfaces only = yes
    hosts allow = 192.168.1. 127.
    server min protocol = SMB2
    smb encrypt = desired

[Public]
    comment = Public Share
    path = /srv/samba/public
    browseable = yes
    writable = yes
    guest ok = yes
    create mask = 0755
    directory mask = 0755

[Private]
    comment = Private Share
    path = /srv/samba/private
    browseable = yes
    writable = yes
    guest ok = no
    valid users = @smbgroup
    create mask = 0660
    directory mask = 0770

[homes]
    comment = Home Directories
    browseable = no
    writable = yes
    valid users = %S
    create mask = 0700
    directory mask = 0700
```

### B. 常用脚本

#### 自动化部署脚本

```bash
#!/bin/bash
# Samba 自动化部署脚本

# 安装 Samba
sudo yum install -y samba samba-client samba-common

# 备份原配置
sudo cp /etc/samba/smb.conf /etc/samba/smb.conf.original

# 创建共享目录
sudo mkdir -p /srv/samba/{public,private}

# 设置权限
sudo chmod 777 /srv/samba/public
sudo chmod 770 /srv/samba/private

# 创建用户组
sudo groupadd smbgroup

# 创建用户
sudo useradd -M -s /sbin/nologin -G smbgroup smbuser
echo "password" | sudo smbpasswd -a -s smbuser

# 配置防火墙
sudo firewall-cmd --permanent --add-service=samba
sudo firewall-cmd --reload

# 配置 SELinux
sudo setsebool -P samba_enable_home_dirs on
sudo setsebool -P samba_export_all_rw on
sudo chcon -t samba_share_t /srv/samba/public
sudo chcon -t samba_share_t /srv/samba/private

# 启动服务
sudo systemctl start smb nmb
sudo systemctl enable smb nmb

# 验证
sudo testparm -s
sudo smbstatus

echo "Samba 部署完成！"
```

### C. 快速参考

#### 常用端口

| 端口 | 协议 | 服务 |
|------|------|------|
| 139 | TCP | NetBIOS 会话服务 |
| 445 | TCP | SMB over TCP |
| 137 | UDP | NetBIOS 名称服务 |
| 138 | UDP | NetBIOS 数据报服务 |

#### 常用路径

| 路径 | 说明 |
|------|------|
| `/etc/samba/smb.conf` | 主配置文件 |
| `/var/log/samba/` | 日志目录 |
| `/var/lib/samba/` | 数据库目录 |
| `/var/lib/samba/private/passdb.tdb` | 用户数据库 |
| `/usr/share/doc/samba-*/` | 文档目录 |

#### 常用配置参数

| 参数 | 说明 | 示例值 |
|------|------|--------|
| `workgroup` | 工作组名称 | WORKGROUP |
| `security` | 安全级别 | user |
| `browseable` | 是否可浏览 | yes/no |
| `writable` | 是否可写 | yes/no |
| `guest ok` | 允许访客 | yes/no |
| `valid users` | 有效用户 | user1, @group1 |
| `create mask` | 文件权限掩码 | 0755 |
| `directory mask` | 目录权限掩码 | 0755 |

### D. 相关资源

#### 官方文档

- Samba 官方网站: https://www.samba.org/
- Samba Wiki: https://wiki.samba.org/
- Samba 文档: https://www.samba.org/samba/docs/

#### 社区资源

- Samba 邮件列表: https://lists.samba.org/
- Samba Bug 追踪: https://bugzilla.samba.org/
- Stack Overflow: https://stackoverflow.com/questions/tagged/samba

#### 推荐阅读

- 《Using Samba》 - O'Reilly
- 《Samba-3 by Example》
- Red Hat 官方文档
- Ubuntu Server 指南

---

**文档版本**: 1.0
**最后更新**: 2026-03-11
**作者**: Samba 运维团队

---

**结束语**

本文档提供了 Samba 服务的完整操作指南，涵盖了从安装配置到故障排查的所有方面。通过本指南，您应该能够：

- ✅ 快速部署和配置 Samba 服务器
- ✅ 熟练管理用户和共享
- ✅ 掌握访问控制和安全配置
- ✅ 独立排查和解决常见问题
- ✅ 遵循最佳实践进行系统维护

如有任何问题或建议，欢迎反馈！
