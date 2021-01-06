# Intro
虾米服务将于2月5日停止。官方提供了用户数据的导出，包括收藏歌单、专辑、艺人、歌曲。
只有收藏的歌曲能对应last.fm的 `track.love` API了。

本分支 `export` 可以把导出的 `id_song.json` 文件解析同步到last.fm上。

# How to use
1. 随便创建一个目录，
2. 下载 `xiamiScrobble` 放进去
3. 下载 `config.example.toml` 放进去，并重命名为 `config.toml`
4. 把 `id_song.json` 重命名为 `song.json` 放进去。
5. 在该目录下执行
    ```
    chmod +x xiamiScrobble
    ./xiamiScrobble
    ```
6. 按照提示，打开last.fm的授权页。
7. Done!
