# 创建测试用的静音音频文件
# 这个脚本可以生成 30 秒的静音 MP3 文件用于测试

# 如果您有 FFmpeg，可以运行以下命令：
# ffmpeg -f lavfi -i anullsrc=r=44100:cl=stereo -t 30 -c:a libmp3lame -b:a 128k work-lofi.mp3
# ffmpeg -f lavfi -i anullsrc=r=44100:cl=stereo -t 30 -c:a libmp3lame -b:a 128k focus-ambient.mp3
# ffmpeg -f lavfi -i anullsrc=r=44100:cl=stereo -t 30 -c:a libmp3lame -b:a 128k creative-jazz.mp3

# 或者从免费音乐网站下载：
echo "请从以下网站下载免费音乐并重命名："
echo "1. https://pixabay.com/music/ - 下载 Lo-Fi 音乐重命名为 work-lofi.mp3"
echo "2. https://freesound.org/ - 下载环境音乐重命名为 focus-ambient.mp3"
echo "3. https://www.chosic.com/free-music/ - 下载爵士乐重命名为 creative-jazz.mp3"
echo ""
echo "文件应该放在: frontend/public/music/ 目录下"