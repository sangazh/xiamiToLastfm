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

# Todo

 * [x] MusicBiz Track ID
 * [ ] 后台跑的进程怎么正常退出？
 * [ ] 虾米album拿不到通过MusicBiz去查

