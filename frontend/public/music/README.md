# 音乐播放器使用说明

## 📁 开发阶段 - 本地音乐文件

将您的音乐文件放在以下目录：
```
frontend/public/music/
├── work-lofi.mp3      # 轻松工作
├── focus-ambient.mp3  # 专注时光
└── creative-jazz.mp3  # 创意灵感
```

## 🎵 推荐的背景音乐类型

### Lo-Fi Hip Hop
- 轻松舒缓，适合编程和工作
- 推荐网站：https://www.chosic.com/free-music/lofi/

### Ambient Music
- 环境音乐，有助于专注
- 推荐网站：https://freemusicarchive.org/genre/Ambient

### 轻柔爵士
- 创意灵感，放松心情
- 推荐网站：https://www.bensound.com/royalty-free-music

## 🌐 上线部署方案

### 方案一：云存储 CDN（推荐）
```javascript
// 修改 MusicPlayer.vue 中的播放列表
const playlist = reactive<Song[]>([
  {
    id: 1,
    title: "轻松工作",
    artist: "Lo-Fi Beats",
    src: "https://your-cdn.com/music/work-lofi.mp3"
  }
])
```

**阿里云 OSS 示例配置：**
1. 创建 OSS Bucket
2. 上传音乐文件
3. 设置公共读权限
4. 获取文件 CDN 链接

**七牛云示例配置：**
```bash
# 使用七牛云 qshell 工具
qshell account <AccessKey> <SecretKey> <AccountName>
qshell qupload music_config.json
```

### 方案二：流媒体 API
```javascript
// 使用网易云音乐 API
const getMusicUrl = async (songId) => {
  const response = await fetch(`/api/song/url?id=${songId}`)
  return response.json()
}
```

### 方案三：GitHub Pages
```bash
# 将音乐文件推送到单独的 repository
git add frontend/public/music/
git commit -m "Add background music files"
git push origin main

# 访问链接格式
https://username.github.io/repository-name/music/song.mp3
```

## ⚖️ 版权注意事项

### 免版权音乐网站
- **Pixabay Music**: https://pixabay.com/music/
- **Freesound**: https://freesound.org/
- **Zapsplat**: https://www.zapsplat.com/
- **YouTube Audio Library**: https://studio.youtube.com/

### 版权声明
```javascript
// 在组件中添加版权信息
const copyrightNotice = "音乐版权归原作者所有，仅供学习交流使用"
```

## 📊 文件大小建议

- **开发环境**: 每首歌曲 < 10MB
- **生产环境**: 每首歌曲 < 5MB
- **总大小**: 所有音乐文件 < 50MB

## 🔧 优化建议

### 音频压缩
```bash
# 使用 FFmpeg 压缩音频
ffmpeg -i input.mp3 -codec:a libmp3lame -b:a 128k output.mp3
```

### 懒加载
```javascript
// 只在用户点击播放时加载音频
const loadAudio = () => {
  if (!audioRef.value.src) {
    audioRef.value.src = currentSong.value.src
  }
}
```

### 预加载策略
```javascript
// 预加载下一首歌曲
const preloadNext = () => {
  const nextIndex = (currentIndex.value + 1) % playlist.length
  const nextAudio = new Audio(playlist[nextIndex].src)
  nextAudio.preload = 'metadata'
}
```