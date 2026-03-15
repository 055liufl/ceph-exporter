# CentOS 7 国内镜像源配置指南

> 验证时间：2026-03-15
>
> **重要提示**：CentOS 7 已于 2024年6月30日 停止维护，部分镜像站已将其移至 vault 目录。

## 一、可用镜像源列表

### 1. 阿里云镜像源 ⭐推荐
- **状态**：✅ 可用 (HTTP 200)
- **官网**：https://developer.aliyun.com/mirror/
- **Base 源**：https://mirrors.aliyun.com/centos/
- **EPEL 源**：https://mirrors.aliyun.com/epel/
- **特点**：速度快，稳定性高，国内访问最佳

### 2. 华为云镜像源 ⭐推荐
- **状态**：✅ 可用 (HTTP 200)
- **官网**：https://mirrors.huaweicloud.com/
- **Base 源**：https://mirrors.huaweicloud.com/centos/
- **特点**：速度快，企业级稳定性

### 3. 清华大学镜像源
- **状态**：✅ 可用 (需使用 vault 路径)
- **官网**：https://mirrors.tuna.tsinghua.edu.cn/
- **Base 源**：https://mirrors.tuna.tsinghua.edu.cn/centos-vault/
- **EPEL 源**：https://mirrors.tuna.tsinghua.edu.cn/epel/
- **特点**：教育网访问速度快

### 4. 中国科学技术大学镜像源
- **状态**：⚠️ 部分限制 (HTTP 403，但可能可用)
- **官网**：https://mirrors.ustc.edu.cn/
- **Base 源**：https://mirrors.ustc.edu.cn/centos-vault/
- **EPEL 源**：https://mirrors.ustc.edu.cn/epel/
- **特点**：教育网优质镜像

### 5. 网易镜像源
- **状态**：❌ 不可用 (HTTP 404)
- **说明**：CentOS 7 源已下线

### 6. 腾讯云镜像源
- **官网**：https://mirrors.cloud.tencent.com/
- **Base 源**：https://mirrors.cloud.tencent.com/centos/
- **特点**：腾讯云用户访问快

---

## 二、CentOS-Base.repo 配置

### 方案一：阿里云源（推荐）

```bash
# 备份原有源
sudo mv /etc/yum.repos.d/CentOS-Base.repo /etc/yum.repos.d/CentOS-Base.repo.backup

# 创建新的 CentOS-Base.repo
sudo cat > /etc/yum.repos.d/CentOS-Base.repo << 'EOF'
[base]
name=CentOS-7 - Base - mirrors.aliyun.com
baseurl=https://mirrors.aliyun.com/centos/7/os/$basearch/
gpgcheck=1
gpgkey=https://mirrors.aliyun.com/centos/RPM-GPG-KEY-CentOS-7

[updates]
name=CentOS-7 - Updates - mirrors.aliyun.com
baseurl=https://mirrors.aliyun.com/centos/7/updates/$basearch/
gpgcheck=1
gpgkey=https://mirrors.aliyun.com/centos/RPM-GPG-KEY-CentOS-7

[extras]
name=CentOS-7 - Extras - mirrors.aliyun.com
baseurl=https://mirrors.aliyun.com/centos/7/extras/$basearch/
gpgcheck=1
gpgkey=https://mirrors.aliyun.com/centos/RPM-GPG-KEY-CentOS-7

[centosplus]
name=CentOS-7 - Plus - mirrors.aliyun.com
baseurl=https://mirrors.aliyun.com/centos/7/centosplus/$basearch/
gpgcheck=1
enabled=0
gpgkey=https://mirrors.aliyun.com/centos/RPM-GPG-KEY-CentOS-7
EOF
```

### 方案二：华为云源

```bash
sudo cat > /etc/yum.repos.d/CentOS-Base.repo << 'EOF'
[base]
name=CentOS-7 - Base - mirrors.huaweicloud.com
baseurl=https://mirrors.huaweicloud.com/centos/7/os/$basearch/
gpgcheck=1
gpgkey=https://mirrors.huaweicloud.com/centos/RPM-GPG-KEY-CentOS-7

[updates]
name=CentOS-7 - Updates - mirrors.huaweicloud.com
baseurl=https://mirrors.huaweicloud.com/centos/7/updates/$basearch/
gpgcheck=1
gpgkey=https://mirrors.huaweicloud.com/centos/RPM-GPG-KEY-CentOS-7

[extras]
name=CentOS-7 - Extras - mirrors.huaweicloud.com
baseurl=https://mirrors.huaweicloud.com/centos/7/extras/$basearch/
gpgcheck=1
gpgkey=https://mirrors.huaweicloud.com/centos/RPM-GPG-KEY-CentOS-7
EOF
```

### 方案三：清华大学源（使用 vault）

```bash
sudo cat > /etc/yum.repos.d/CentOS-Base.repo << 'EOF'
[base]
name=CentOS-7 - Base - mirrors.tuna.tsinghua.edu.cn
baseurl=https://mirrors.tuna.tsinghua.edu.cn/centos-vault/7.9.2009/os/$basearch/
gpgcheck=1
gpgkey=https://mirrors.tuna.tsinghua.edu.cn/centos/RPM-GPG-KEY-CentOS-7

[updates]
name=CentOS-7 - Updates - mirrors.tuna.tsinghua.edu.cn
baseurl=https://mirrors.tuna.tsinghua.edu.cn/centos-vault/7.9.2009/updates/$basearch/
gpgcheck=1
gpgkey=https://mirrors.tuna.tsinghua.edu.cn/centos/RPM-GPG-KEY-CentOS-7

[extras]
name=CentOS-7 - Extras - mirrors.tuna.tsinghua.edu.cn
baseurl=https://mirrors.tuna.tsinghua.edu.cn/centos-vault/7.9.2009/extras/$basearch/
gpgcheck=1
gpgkey=https://mirrors.tuna.tsinghua.edu.cn/centos/RPM-GPG-KEY-CentOS-7
EOF
```

---

## 三、EPEL 源配置

### 方案一：阿里云 EPEL 源（推荐）

```bash
# 安装 EPEL 仓库
sudo yum install -y epel-release

# 备份原有 EPEL 源
sudo mv /etc/yum.repos.d/epel.repo /etc/yum.repos.d/epel.repo.backup
sudo mv /etc/yum.repos.d/epel-testing.repo /etc/yum.repos.d/epel-testing.repo.backup

# 创建新的 EPEL 源
sudo cat > /etc/yum.repos.d/epel.repo << 'EOF'
[epel]
name=Extra Packages for Enterprise Linux 7 - $basearch
baseurl=https://mirrors.aliyun.com/epel/7/$basearch
failovermethod=priority
enabled=1
gpgcheck=1
gpgkey=https://mirrors.aliyun.com/epel/RPM-GPG-KEY-EPEL-7

[epel-debuginfo]
name=Extra Packages for Enterprise Linux 7 - $basearch - Debug
baseurl=https://mirrors.aliyun.com/epel/7/$basearch/debug
failovermethod=priority
enabled=0
gpgkey=https://mirrors.aliyun.com/epel/RPM-GPG-KEY-EPEL-7
gpgcheck=1

[epel-source]
name=Extra Packages for Enterprise Linux 7 - $basearch - Source
baseurl=https://mirrors.aliyun.com/epel/7/SRPMS
failovermethod=priority
enabled=0
gpgkey=https://mirrors.aliyun.com/epel/RPM-GPG-KEY-EPEL-7
gpgcheck=1
EOF
```

### 方案二：清华大学 EPEL 源

```bash
sudo cat > /etc/yum.repos.d/epel.repo << 'EOF'
[epel]
name=Extra Packages for Enterprise Linux 7 - $basearch
baseurl=https://mirrors.tuna.tsinghua.edu.cn/epel/7/$basearch
failovermethod=priority
enabled=1
gpgcheck=1
gpgkey=https://mirrors.tuna.tsinghua.edu.cn/epel/RPM-GPG-KEY-EPEL-7
EOF
```

---

## 四、配置后的操作

### 1. 清理缓存并更新

```bash
# 清理 yum 缓存
sudo yum clean all

# 生成缓存
sudo yum makecache

# 更新系统（可选）
sudo yum update -y
```

### 2. 验证源配置

```bash
# 查看已启用的仓库
yum repolist

# 查看所有仓库（包括禁用的）
yum repolist all

# 测试安装软件包
sudo yum install -y vim
```

---

## 五、验证结果总结

| 镜像源 | CentOS Base | EPEL | 推荐度 |
|--------|-------------|------|--------|
| 阿里云 | ✅ 200 | ✅ 200 | ⭐⭐⭐⭐⭐ |
| 华为云 | ✅ 200 | - | ⭐⭐⭐⭐ |
| 清华大学 | ✅ 200 (vault) | ⚠️ 404 | ⭐⭐⭐ |
| 中科大 | ⚠️ 403 | ⚠️ 403 | ⭐⭐ |
| 网易 | ❌ 404 | ❌ 404 | ❌ |
| 腾讯云 | ✅ 未测试 | - | ⭐⭐⭐ |

---

## 六、常见问题

### 1. CentOS 7 停止维护后如何继续使用？

CentOS 7 已于 2024年6月30日 停止维护，但仍可以使用：
- 使用阿里云、华为云等仍提供 CentOS 7 镜像的源
- 使用 centos-vault 归档源（清华、中科大等）
- 考虑迁移到 Rocky Linux 或 AlmaLinux

### 2. 如何选择最快的镜像源？

可以使用以下命令测试各镜像源的速度：

```bash
time curl -o /dev/null https://mirrors.aliyun.com/centos/7/os/x86_64/repodata/repomd.xml
time curl -o /dev/null https://mirrors.huaweicloud.com/centos/7/os/x86_64/repodata/repomd.xml
```

### 3. 更换源后出现 GPG 密钥错误？

```bash
# 导入 GPG 密钥
sudo rpm --import https://mirrors.aliyun.com/centos/RPM-GPG-KEY-CentOS-7
sudo rpm --import https://mirrors.aliyun.com/epel/RPM-GPG-KEY-EPEL-7
```

---

## 参考资料

- [CentOS 7常用国内源配置：阿里云、腾讯云、华为云、清华源](https://blog.csdn.net/weixin_44149170/article/details/150574339)
- [CentOS7国内镜像源配置指南：阿里云、清华、中科大等源优化](https://comate.baidu.com/zh/page/psawrgp8wvk)
- [CentOS 7 停更后如何配置YUM 源？（Vault、EPEL）](https://blog.csdn.net/liuguizhong/article/details/154388785)
- [CentOS 7 YUM源配置指南：2025年最新国内镜像源方案](https://comate.baidu.com/zh/page/5hu1kqzh4yj)

---

**最后更新**：2026-03-15
**推荐配置**：阿里云 CentOS Base + 阿里云 EPEL
