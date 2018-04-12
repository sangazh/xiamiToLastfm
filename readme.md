# Intro
Yet another go project to scrobble xiami songs to last.fm.

设定的是每三分钟去爬一次虾米主页里的最近播放歌曲，1小时以前的都会放弃，因为时间不准确了。

# How to use
在下载的目录下
```
cp config.example.toml config.toml

./xiami_to_last
```
按照提示输入虾米主页url，在需要打开last.fm的某url时，打开页面授权。放置即可。

# Todo
- 同步虾米收藏的歌曲到last.fm的fav
- MusicBiz
- 定时
- 根据特定错误code处理
