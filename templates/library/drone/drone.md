title: 本地git库连接github
date: 2014-12-05 15:21:30
tags:
categories: git
---

### 1. 注册github账户，并创建一个仓库 ##

  github地址：[https://github.com/](https://github.com/)，点击注册即可,并创建一个test仓库。

### 2. 与github建立通信

使用命令生成公钥和私钥：

    ssh-keygen -t rsa

将公钥（id_rsa.pub）复制到githib的ssh keys，执行以下命令测试连接是否成功

    ssh -T git@github.com

提示：Hi zhihongme! You've successfully authenticated, but GitHub does not provide shell access. 标示连接成功！

### 3. 设置用户信息

    $ git config --global github.user zhihongme      //github 上的用户名
    $ git config --global github.email g.success16@gmail.com

### 4. 在本地创建一个项目（仓库）

        $ makdir ~/test    //创建一个项目test
        $ cd ~/test    //打开这个项目
        $ git init    //初始化
        $ touch README
        $ git add README   //更新README文件
        $ git commit -m 'first commit'//提交更新，并注释信息“first commit”
        $ git remote add origin git@github.com:zhihongme/test.git   //连接远程github项目
        $ git push -u origin master   //将本地项目更新到github项目上去

### 5. 可能出现的错误

1. 在执行

    $ git remote addorigin git@github.com:zhihongme/test.git

错误提示：fatal: remote origin already exists.(原因是在github上已经存在test仓库了)

解决办法：

    $ git remote rm origin

然后在执行：$ git remote add origin git@github.com:zhihongme/test.git 就不会报错误了

2. 在执行

            $ git push origin master

    错误提示：error:failed to push som refs to.......

    解决办法：

                $ git pull origin master // 先把远程服务器github上面的文件pull下来，再push 上去。
