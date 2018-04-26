# Intro
Yet another go project to scrobble your xiami recent songs to last.fm.

* Scrobbling now
* Recent played

设定为每分钟去爬一次虾米主页里的最近播放歌曲，1小时以前的都会放弃，因为时间不准确了。

# How to use
1. 随便创建一个目录，
2. 下载 `xiamiScrobble` 放进去
3. 下载 `config.example.toml` 放进去，并重命名为 `config.toml`
4. 在该目录下执行
    ```
    chmod +x xiamiScrobble
    ./xiamiScrobble
    ```
5. 按照提示，输入虾米主页url
6. 按照提示，打开last.fm的授权页。
7. Done!
8. Ctrl+C 退出

# Error
## 授权失败
如果报错。

```
Invalid session key - Please re-authenticate
```
有可能是当前`API key`失效了，你可以联系我，也可以自己去last.fm[创建一个API账户](https://www.last.fm/api/account/create)，
并将获得的`API key`和`Shared secret`分别填入`config.toml`的`shared_secret`和`api_key`，重新运行即可。

或者~

等几天再试就好了，可能last.fm对一定时间段内的授权次数有限制。

# Todo

 * [ ] ~~同步虾米收藏的歌曲到last.fm的fav~~ 收藏页面404无法访问
 * [x]  MusicBiz Track ID
 * [ ] 后台跑的进程怎么正常退出？

